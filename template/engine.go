package template

type Engine interface {
	Render(tmpl string, data any) (string, error)
}
