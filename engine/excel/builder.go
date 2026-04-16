package excel

import (
	"bytes"
	"fmt"
	texttmpl "text/template"
)

type Sheet struct {
	Name string
	Rows [][]string
}

func Build(tmpl string, data any) ([]Sheet, error) {
	b := &builder{}

	t := texttmpl.New("excel").Funcs(texttmpl.FuncMap{
		"sheet": b.sheet,
		"row":   b.row,
	})

	t, err := t.Parse(tmpl)
	if err != nil {
		return nil, err
	}

	_ = t.Execute(&bytes.Buffer{}, data)
	return b.sheets, nil
}

type builder struct {
	sheets []Sheet
	cur    *Sheet
}

func (b *builder) sheet(name string) string {
	s := Sheet{Name: name}
	b.sheets = append(b.sheets, s)
	b.cur = &b.sheets[len(b.sheets)-1]
	return ""
}

func (b *builder) row(values ...any) string {
	r := make([]string, len(values))
	for i, v := range values {
		r[i] = fmt.Sprintf("%v", v)
	}
	b.cur.Rows = append(b.cur.Rows, r)
	return ""
}
