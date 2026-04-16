package csv

import (
	"bytes"
	"fmt"
	texttmpl "text/template"
)

func Build(tmpl string, data any) ([][]string, error) {
	var rows [][]string

	t := texttmpl.New("csv").Funcs(texttmpl.FuncMap{
		"row": func(values ...any) string {
			r := make([]string, len(values))
			for i, v := range values {
				r[i] = fmt.Sprintf("%v", v)
			}
			rows = append(rows, r)
			return ""
		},
	})

	t, err := t.Parse(tmpl)
	if err != nil {
		return nil, err
	}

	_ = t.Execute(&bytes.Buffer{}, data)
	return rows, nil
}
