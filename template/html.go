package template

import (
	"bytes"
	"html/template"
	"sync"
)

type htmlEngine struct{}

func NewHTML() Engine { return &htmlEngine{} }

var (
	parsedHTML  sync.Map // tmpl string -> *template.Template
	htmlBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

func (e *htmlEngine) Render(tmpl string, data any) (string, error) {
	t, err := parsedHTMLTemplate(tmpl)
	if err != nil {
		return "", err
	}
	buf := htmlBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer htmlBufPool.Put(buf)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func parsedHTMLTemplate(tmpl string) (*template.Template, error) {
	if v, ok := parsedHTML.Load(tmpl); ok {
		return v.(*template.Template), nil
	}
	parsed, err := template.New("html").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	actual, _ := parsedHTML.LoadOrStore(tmpl, parsed)
	return actual.(*template.Template), nil
}
