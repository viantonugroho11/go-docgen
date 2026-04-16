package template

import (
	"bytes"
	"html/template"
	"sync"
)

type htmlEngine struct{}

func NewHTML() Engine { return &htmlEngine{} }

var parsedHTML sync.Map // tmpl string -> *template.Template

func (e *htmlEngine) Render(tmpl string, data any) (string, error) {
	base, err := parsedHTMLTemplate(tmpl)
	if err != nil {
		return "", err
	}
	t, err := base.Clone()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}

func parsedHTMLTemplate(tmpl string) (*template.Template, error) {
	if v, ok := parsedHTML.Load(tmpl); ok {
		return v.(*template.Template), nil
	}
	parsed, err := template.New("html").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	parsedHTML.Store(tmpl, parsed)
	return parsed, nil
}
