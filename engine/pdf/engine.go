package pdf

import "context"

type Engine interface {
	Render(ctx context.Context, html string) ([]byte, error)
}
