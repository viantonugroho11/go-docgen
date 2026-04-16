package csv

import (
	"bytes"
	texttmpl "text/template"

	"github.com/viantonugroho11/go-docgen/internal/strfmt"
)

func Build(tmpl string, data any) ([][]string, error) {
	var rows [][]string

	t := texttmpl.New("csv").Funcs(texttmpl.FuncMap{
		"row": func(values ...any) string {
			r := make([]string, len(values))
			for i, v := range values {
				r[i] = strfmt.FormatAny(v)
			}
			rows = append(rows, r)
			return ""
		},
	})

	t, err := t.Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_ = t.Execute(&buf, data)
	return rows, nil
}
