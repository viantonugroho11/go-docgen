package pdf

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// EngineConfig configures the PDF engine.
type EngineConfig struct {
	Timeout time.Duration
	// Mode defaults to RenderModeAuto when zero.
	Mode RenderMode
	// MaxConcurrency limits concurrent Chromium tab renders.
	// Defaults to runtime.GOMAXPROCS(0). Has no effect for RenderModeLight.
	MaxConcurrency int
}

type config struct {
	timeout        time.Duration
	mode           RenderMode
	maxConcurrency int
}

type engine struct {
	cfg config
	sem chan struct{} // bounded concurrency gate; nil when mode == RenderModeLight

	// Persistent Chromium process — guarded by mu.
	// The allocator and browser are created once (lazily on first render) and
	// reused across all calls. Chrome starts in ~100–500 ms; subsequent renders
	// only pay the cost of opening a new tab (~10–50 ms).
	mu            sync.Mutex
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	browserCtx    context.Context
	browserCancel context.CancelFunc
}

// New builds a PDF Engine from cfg. A non-positive Timeout defaults to 10 seconds.
func New(cfg EngineConfig) Engine {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	mode := cfg.Mode
	if mode > RenderModeLight {
		mode = RenderModeAuto
	}
	maxConc := cfg.MaxConcurrency
	if maxConc <= 0 {
		maxConc = runtime.GOMAXPROCS(0)
	}

	e := &engine{
		cfg: config{
			timeout:        timeout,
			mode:           mode,
			maxConcurrency: maxConc,
		},
	}
	if mode != RenderModeLight {
		e.sem = make(chan struct{}, maxConc)
	}
	return e
}

// Close shuts down the persistent browser process. Safe to call multiple times.
// After Close, the next Render call restarts the browser automatically.
func (e *engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.closeLocked()
	return nil
}

func (e *engine) closeLocked() {
	if e.browserCancel != nil {
		e.browserCancel()
		e.browserCancel = nil
		e.browserCtx = nil
	}
	if e.allocCancel != nil {
		e.allocCancel()
		e.allocCancel = nil
		e.allocCtx = nil
	}
}

// ensureBrowser initialises or reinitialises the persistent Chromium allocator and
// browser context. All callers are serialised during startup (one-time cost); the
// fast path (browser already live) takes only a mutex acquire + channel peek.
//
// Chrome starts lazily on the first renderInTab call — not here — so the mutex
// hold time is always short (context allocation only, no I/O).
func (e *engine) ensureBrowser() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.browserCtx != nil {
		select {
		case <-e.browserCtx.Done():
			// Browser died (crash or Close); fall through to reinit.
		default:
			return nil // Fast path: browser alive.
		}
	}

	e.closeLocked()

	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	browserCtx, browserCancel := chromedp.NewContext(allocCtx)

	e.allocCtx = allocCtx
	e.allocCancel = allocCancel
	e.browserCtx = browserCtx
	e.browserCancel = browserCancel
	return nil
}

func (e *engine) Render(ctx context.Context, html string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.timeout)
	defer cancel()

	switch e.cfg.mode {
	case RenderModeChromium:
		return e.renderChromium(ctx, html)
	case RenderModeLight:
		return light(ctx, html)
	default: // RenderModeAuto
		out, err := e.renderChromium(ctx, html)
		if err == nil {
			return out, nil
		}
		return light(ctx, html)
	}
}

func (e *engine) renderChromium(ctx context.Context, html string) ([]byte, error) {
	if err := e.ensureBrowser(); err != nil {
		return nil, err
	}

	// Gate concurrent tab renders; respect cancellation while waiting.
	select {
	case e.sem <- struct{}{}:
		defer func() { <-e.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	e.mu.Lock()
	browserCtx := e.browserCtx
	e.mu.Unlock()

	// Open a new tab in the shared browser.
	tabCtx, tabCancel := chromedp.NewContext(browserCtx)
	defer tabCancel()

	// Bind the render deadline to the tab so a slow render cannot block indefinitely
	// while keeping the shared browser process unaffected by the timeout.
	if dl, ok := ctx.Deadline(); ok {
		var cancel context.CancelFunc
		tabCtx, cancel = context.WithDeadline(tabCtx, dl)
		defer cancel()
	}

	return renderInTab(tabCtx, html)
}
