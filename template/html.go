package template

import (
	"bytes"
	"html/template"
)

type htmlEngine struct{}

func NewHTML() Engine { return &htmlEngine{} }

func (e *htmlEngine) Render(tmpl string, data any) (string, error) {
	t, err := template.New("html").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	return buf.String(), err
}
