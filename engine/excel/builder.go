package excel

import (
	"bytes"
	"sync"
	texttmpl "text/template"

	"github.com/viantonugroho11/go-docgen/internal/strfmt"
)

var (
	excelTmplCache sync.Map // map[string]*texttmpl.Template
	excelExecBuf   = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

type Sheet struct {
	Name string
	Rows [][]string
}

func Build(src string, data any) ([]Sheet, error) {
	base, err := cachedExcelTemplate(src)
	if err != nil {
		return nil, err
	}

	b := &builder{}

	t, err := base.Clone()
	if err != nil {
		return nil, err
	}
	t.Funcs(texttmpl.FuncMap{
		"sheet": b.sheet,
		"row":   b.row,
	})

	buf := excelExecBuf.Get().(*bytes.Buffer)
	buf.Reset()
	_ = t.Execute(buf, data)
	excelExecBuf.Put(buf)

	return b.sheets, nil
}

func cachedExcelTemplate(src string) (*texttmpl.Template, error) {
	if v, ok := excelTmplCache.Load(src); ok {
		return v.(*texttmpl.Template), nil
	}
	t := texttmpl.New("excel").Funcs(texttmpl.FuncMap{
		"sheet": func(string) string { return "" },    // placeholder
		"row":   func(...any) string { return "" },    // placeholder
	})
	t, err := t.Parse(src)
	if err != nil {
		return nil, err
	}
	actual, _ := excelTmplCache.LoadOrStore(src, t)
	return actual.(*texttmpl.Template), nil
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
