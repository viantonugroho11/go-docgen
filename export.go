// Package docgen generates PDF, CSV, and Excel documents from templates.
package docgen

import (
	"context"

	"github.com/viantonugroho11/go-docgen/engine/csv"
	"github.com/viantonugroho11/go-docgen/engine/excel"
	"github.com/viantonugroho11/go-docgen/engine/pdf"
	"github.com/viantonugroho11/go-docgen/internal/loader"
	"github.com/viantonugroho11/go-docgen/template"
)

// Generator turns templates plus data into document bytes (PDF, CSV, or XLSX).
type Generator interface {
	PDF(ctx context.Context, template string, data any) ([]byte, error)
	PDFFromFile(ctx context.Context, path string, data any) ([]byte, error)

	CSV(ctx context.Context, template string, data any) ([]byte, error)
	CSVFromFile(ctx context.Context, path string, data any) ([]byte, error)

	Excel(ctx context.Context, template string, data any) ([]byte, error)
	ExcelFromFile(ctx context.Context, path string, data any) ([]byte, error)
}

type generator struct {
	cfg Config

	htmlTmpl template.Engine

	pdfEngine   pdf.Engine
	csvEngine   csv.Engine
	excelEngine excel.Engine
}

// New returns a Generator with optional configuration (see WithTimeout).
func New(opts ...Option) Generator {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	return &generator{
		cfg:         cfg,
		htmlTmpl:    template.NewHTML(),
		pdfEngine:   pdf.New(cfg.Timeout),
		csvEngine:   csv.New(),
		excelEngine: excel.New(),
	}
}

func (g *generator) PDF(ctx context.Context, template string, data any) ([]byte, error) {
	html, err := g.htmlTmpl.Render(template, data)
	if err != nil {
		return nil, err
	}
	return g.pdfEngine.Render(ctx, html)
}

func (g *generator) PDFFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return g.PDF(ctx, tmpl, data)
}

func (g *generator) CSV(ctx context.Context, template string, data any) ([]byte, error) {
	rows, err := csv.Build(template, data)
	if err != nil {
		return nil, err
	}
	return g.csvEngine.Generate(rows)
}

func (g *generator) CSVFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return g.CSV(ctx, tmpl, data)
}

func (g *generator) Excel(ctx context.Context, template string, data any) ([]byte, error) {
	sheets, err := excel.Build(template, data)
	if err != nil {
		return nil, err
	}
	return g.excelEngine.Generate(sheets)
}

func (g *generator) ExcelFromFile(ctx context.Context, path string, data any) ([]byte, error) {
	tmpl, err := loader.Load(path)
	if err != nil {
		return nil, err
	}
	return g.Excel(ctx, tmpl, data)
}
