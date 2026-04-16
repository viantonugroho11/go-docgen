package pdf

import (
	"context"
	"time"
)

// EngineConfig configures the PDF engine.
type EngineConfig struct {
	Timeout time.Duration
	// Mode defaults to RenderModeAuto when zero.
	Mode RenderMode
}

type config struct {
	timeout time.Duration
	mode    RenderMode
}

type engine struct {
	cfg config
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
	return &engine{cfg: config{timeout: timeout, mode: mode}}
}

func (e *engine) Render(ctx context.Context, html string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.timeout)
	defer cancel()

	switch e.cfg.mode {
	case RenderModeChromium:
		return chromium(ctx, html)
	case RenderModeLight:
		return light(ctx, html)
	default:
		pdfBytes, err := chromium(ctx, html)
		if err == nil {
			return pdfBytes, nil
		}
		return light(ctx, html)
	}
}
