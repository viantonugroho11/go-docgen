package excel

import (
	"bytes"
	texttmpl "text/template"

	"github.com/viantonugroho11/go-docgen/internal/strfmt"
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

	var buf bytes.Buffer
	_ = t.Execute(&buf, data)
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
		r[i] = strfmt.FormatAny(v)
	}
	b.cur.Rows = append(b.cur.Rows, r)
	return ""
}
