package csv

import (
	"bytes"
	"sync"
	texttmpl "text/template"

	"github.com/viantonugroho11/go-docgen/internal/strfmt"
)

var (
	csvTmplCache sync.Map // map[string]*texttmpl.Template
	csvExecBuf   = sync.Pool{New: func() any { return new(bytes.Buffer) }}
)

func Build(src string, data any) ([][]string, error) {
	base, err := cachedCSVTemplate(src)
	if err != nil {
		return nil, err
	}

	var rows [][]string

	t, err := base.Clone()
	if err != nil {
		return nil, err
	}
	t.Funcs(texttmpl.FuncMap{
		"row": func(values ...any) string {
			r := make([]string, len(values))
			for i, v := range values {
				r[i] = strfmt.FormatAny(v)
			}
			rows = append(rows, r)
			return ""
		},
	})

	buf := csvExecBuf.Get().(*bytes.Buffer)
	buf.Reset()
	_ = t.Execute(buf, data)
	csvExecBuf.Put(buf)

	return rows, nil
}

func cachedCSVTemplate(src string) (*texttmpl.Template, error) {
	if v, ok := csvTmplCache.Load(src); ok {
		return v.(*texttmpl.Template), nil
	}
	t := texttmpl.New("csv").Funcs(texttmpl.FuncMap{
		"row": func(...any) string { return "" }, // placeholder; replaced per-call via Clone+Funcs
	})
	t, err := t.Parse(src)
	if err != nil {
		return nil, err
	}
	actual, _ := csvTmplCache.LoadOrStore(src, t)
	return actual.(*texttmpl.Template), nil
}
