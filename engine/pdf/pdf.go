package pdf

import (
	"context"
	"time"
)

type config struct {
	timeout time.Duration
}

type engine struct {
	cfg config
}

func New(timeout time.Duration) Engine {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &engine{cfg: config{timeout: timeout}}
}

func (e *engine) Render(ctx context.Context, html string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.timeout)
	defer cancel()

	pdf, err := chromium(ctx, html)
	if err == nil {
		return pdf, nil
	}

	return light(ctx, html)
}
