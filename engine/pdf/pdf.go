package pdf

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// EngineConfig configures the PDF engine.
type EngineConfig struct {
	Timeout time.Duration
	// Mode defaults to RenderModeAuto when zero.
	Mode RenderMode
	// MaxConcurrency limits concurrent Chromium tab renders and sizes the
	// reusable tab pool. Defaults to runtime.GOMAXPROCS(0). Has no effect
	// for RenderModeLight.
	MaxConcurrency int
	// ChromePath is the absolute path to a Chromium-family executable. When
	// empty, chromedp searches PATH for the standard Chrome/Chromium binaries.
	// Set this to point at chrome-headless-shell (~80 MB, vs ~200 MB for full
	// Chrome) for slim container images. Has no effect for RenderModeLight.
	ChromePath string
	// ExtraFlags are appended to the Chromium argv after the chromedp defaults.
	// Use for site-specific tuning (e.g. --font-render-hinting=none for
	// deterministic screenshots). Has no effect for RenderModeLight.
	ExtraFlags []string
}

type config struct {
	timeout        time.Duration
	mode           RenderMode
	maxConcurrency int
	chromePath     string
	extraFlags     []string
}

// tabHandle is a persistent Chromium tab (target). The pool holds either a
// materialised handle or nil (a token to be materialised on first use / after
// invalidation). Tabs are never closed during normal operation — Close on
// the engine cancels the browser context, which cascades to all pooled tabs.
type tabHandle struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type engine struct {
	cfg  config
	pool chan *tabHandle // buffered maxConcurrency; nil entries are lazy slots

	// Persistent Chromium process — guarded by mu.
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
			chromePath:     cfg.ChromePath,
			extraFlags:     append([]string(nil), cfg.ExtraFlags...),
		},
	}
	if mode != RenderModeLight {
		e.pool = make(chan *tabHandle, maxConc)
		for i := 0; i < maxConc; i++ {
			e.pool <- nil // lazy slot; materialised on first render
		}
	}
	return e
}

// Close shuts down the persistent browser process. Safe to call multiple times.
// Cancelling the browser context cascades to every pooled tab, so pool entries
// become dead-but-valid tokens; the next Render observes ctx.Err() != nil and
// materialises a fresh tab against the freshly-started browser.
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

// ensureBrowser initialises or reinitialises the persistent Chromium allocator
// and browser context. Fast path (browser alive) is a mutex acquire + channel peek.
func (e *engine) ensureBrowser() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.browserCtx != nil {
		select {
		case <-e.browserCtx.Done():
			// Browser died; fall through to reinit.
		default:
			return nil
		}
	}

	e.closeLocked()

	opts := append([]chromedp.ExecAllocatorOption(nil), chromedp.DefaultExecAllocatorOptions[:]...)
	if e.cfg.chromePath != "" {
		opts = append(opts, chromedp.ExecPath(e.cfg.chromePath))
	}
	for _, f := range e.cfg.extraFlags {
		name, value, hasValue := strings.Cut(f, "=")
		if hasValue {
			opts = append(opts, chromedp.Flag(name, value))
		} else {
			opts = append(opts, chromedp.Flag(name, true))
		}
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		opts...,
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

// renderChromium checks out a reusable tab from the pool (creating one lazily
// if the slot is empty or the previous handle died), runs the render pipeline
// against it, then returns the handle. Concurrency is bounded by pool
// capacity — the pool acts as both semaphore and tab cache.
//
// Reusing tabs skips the ~200-500 ms chromedp.NewContext (target + Page.enable)
// cost per render after warmup. renderInTab resets the DOM implicitly via
// page.SetDocumentContent, so no state leaks between renders.
func (e *engine) renderChromium(ctx context.Context, html string) ([]byte, error) {
	if err := e.ensureBrowser(); err != nil {
		return nil, err
	}

	var h *tabHandle
	select {
	case h = <-e.pool:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if h == nil || h.ctx.Err() != nil {
		if h != nil && h.cancel != nil {
			h.cancel()
		}
		e.mu.Lock()
		browserCtx := e.browserCtx
		e.mu.Unlock()
		tabCtx, tabCancel := chromedp.NewContext(browserCtx)
		h = &tabHandle{ctx: tabCtx, cancel: tabCancel}
	}

	// Bind the caller deadline to this render only, without killing the
	// underlying pooled tab when the deadline expires.
	runCtx := h.ctx
	if dl, ok := ctx.Deadline(); ok {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithDeadline(h.ctx, dl)
		defer cancel()
	}

	out, err := renderInTab(runCtx, html)

	// If the pooled tab itself died (browser crash, target closed), discard the
	// handle so the next checkout materialises a fresh one. Otherwise keep it.
	if h.ctx.Err() != nil {
		h.cancel()
		e.pool <- nil
	} else {
		e.pool <- h
	}
	return out, err
}
