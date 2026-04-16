package template

import (
	"bytes"
	"sync"
	texttmpl "text/template"
)

type textEngine struct{}

func NewText() Engine { return &textEngine{} }

var parsedText sync.Map // tmpl string -> *texttmpl.Template

func (e *textEngine) Render(tmpl string, data any) (string, error) {
	base, err := parsedTextTemplate(tmpl)
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

func parsedTextTemplate(tmpl string) (*texttmpl.Template, error) {
	if v, ok := parsedText.Load(tmpl); ok {
		return v.(*texttmpl.Template), nil
	}
	parsed, err := texttmpl.New("text").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	parsedText.Store(tmpl, parsed)
	return parsed, nil
}
