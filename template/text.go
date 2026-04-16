package template

import (
	"bytes"
	texttmpl "text/template"
)

type textEngine struct{}

func NewText() Engine { return &textEngine{} }

func (e *textEngine) Render(tmpl string, data any) (string, error) {
	t, err := texttmpl.New("text").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}
