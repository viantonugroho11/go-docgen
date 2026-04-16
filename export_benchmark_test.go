package docgen

import (
	"context"
	"testing"
	"time"
)

const benchmarkGeneratorPDFHTML = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Bench</title></head><body>` +
	`<h1>Report</h1><table>` +
	`{{range .People}}<tr><td>{{.Name}}</td><td>{{.Age}}</td></tr>{{end}}` +
	`</table></body></html>`

func benchmarkCSVData() map[string]any {
	people := make([]map[string]any, 0, 100)
	for i := 0; i < 100; i++ {
		people = append(people, map[string]any{
			"Name": "User",
			"Age":  i + 20,
		})
	}
	return map[string]any{"People": people}
}

func benchmarkExcelData() map[string]any {
	users := make([]map[string]any, 0, 100)
	for i := 0; i < 100; i++ {
		users = append(users, map[string]any{
			"Name": "User",
			"Role": "Member",
		})
	}
	return map[string]any{"Users": users}
}

func BenchmarkGenerator_PDF(b *testing.B) {
	gen := New(WithPDFRenderMode(PDFRenderLight))
	ctx := context.Background()
	data := benchmarkCSVData()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.PDF(ctx, benchmarkGeneratorPDFHTML, data); err != nil {
			b.Fatalf("PDF() error = %v", err)
		}
	}
}

// BenchmarkGenerator_PDF_Chromium measures headless Chromium + PrintToPDF via Generator.PDF.
// Skipped under go test -short so CI and quick runs stay fast.
func BenchmarkGenerator_PDF_Chromium(b *testing.B) {
	if testing.Short() {
		b.Skip("omit chromedp in -short; run: go test -bench=BenchmarkGenerator_PDF_Chromium -benchmem -run=^$ .")
	}
	gen := New(
		WithPDFRenderMode(PDFRenderChromium),
		WithTimeout(3*time.Minute),
	)
	ctx := context.Background()
	data := benchmarkCSVData()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.PDF(ctx, benchmarkGeneratorPDFHTML, data); err != nil {
			b.Fatalf("PDF() error = %v", err)
		}
	}
}

func BenchmarkGenerator_CSV(b *testing.B) {
	gen := New()
	ctx := context.Background()
	tmpl := `{{row "Name" "Age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := benchmarkCSVData()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.CSV(ctx, tmpl, data); err != nil {
			b.Fatalf("CSV() error = %v", err)
		}
	}
}

func BenchmarkGenerator_Excel(b *testing.B) {
	gen := New()
	ctx := context.Background()
	tmpl := `{{sheet "Users"}}{{row "Name" "Role"}}{{range .Users}}{{row .Name .Role}}{{end}}`
	data := benchmarkExcelData()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.Excel(ctx, tmpl, data); err != nil {
			b.Fatalf("Excel() error = %v", err)
		}
	}
}
