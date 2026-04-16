package godocgen

import (
	"context"

	"github.com/viantonugroho11/go-docgen/engine/csv"
	"github.com/viantonugroho11/go-docgen/engine/excel"
	"github.com/viantonugroho11/go-docgen/engine/pdf"
	"github.com/viantonugroho11/go-docgen/internal/loader"
	"github.com/viantonugroho11/go-docgen/template"
)

type Exporter interface {
	ToPDFTemplate(ctx context.Context, tmpl string, data any) ([]byte, error)
	ToPDFFromFile(ctx context.Context, path string, data any) ([]byte, error)

	ToCSVTemplate(ctx context.Context, tmpl string, data any) ([]byte, error)
	ToCSVFromFile(ctx context.Context, path string, data any) ([]byte, error)

	ToExcelTemplate(ctx context.Context, tmpl string, data any) ([]byte, error)
	ToExcelFromFile(ctx context.Context, path string, data any) ([]byte, error)
}

type exporter struct {
	cfg Config

	htmlTmpl template.Engine
	textTmpl template.Engine

	pdfEngine   pdf.Engine
	csvEngine   csv.Engine
	excelEngine excel.Engine
}

func New(opts ...Option) Exporter {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	return &exporter{
		cfg:         cfg,
		htmlTmpl:    template.NewHTML(),
		textTmpl:    template.NewText(),
		pdfEngine:   pdf.New(cfg.Timeout),
		csvEngine:   csv.New(),
		excelEngine: excel.New(),
	}
}

func (e *exporter) ToPDFTemplate(ctx context.Context, tmpl string, data any) ([]byte, error) {
	html, err := e.htmlTmpl.Render(tmpl, data)
	if err != nil {
		return nil, err
	}
	return e.pdfEngine.Render(ctx, html)
}

func (e *exporter) ToPDFFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return e.ToPDFTemplate(ctx, tmpl, data)
}

func (e *exporter) ToCSVTemplate(ctx context.Context, tmpl string, data any) ([]byte, error) {
	rows, err := csv.Build(tmpl, data)
	if err != nil {
		return nil, err
	}
	return e.csvEngine.Generate(rows)
}

func (e *exporter) ToCSVFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return e.ToCSVTemplate(ctx, tmpl, data)
}

func (e *exporter) ToExcelTemplate(ctx context.Context, tmpl string, data any) ([]byte, error) {
	sheets, err := excel.Build(tmpl, data)
	if err != nil {
		return nil, err
	}
	return e.excelEngine.Generate(sheets)
}

func (e *exporter) ToExcelFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return e.ToExcelTemplate(ctx, tmpl, data)
}
