package docgen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew_AppliesTimeoutOption(t *testing.T) {
	g, ok := New(WithTimeout(2 * time.Second)).(*generator)
	if !ok {
		t.Fatal("expected concrete *generator type")
	}
	if g.cfg.Timeout != 2*time.Second {
		t.Fatalf("timeout = %v, want %v", g.cfg.Timeout, 2*time.Second)
	}
}

func TestNew_AppliesPDFRenderModeOption(t *testing.T) {
	g, ok := New(WithPDFRenderMode(PDFRenderLight)).(*generator)
	if !ok {
		t.Fatal("expected concrete *generator type")
	}
	if g.cfg.PDFRenderMode != PDFRenderLight {
		t.Fatalf("PDFRenderMode = %v, want PDFRenderLight", g.cfg.PDFRenderMode)
	}
}

func TestGenerator_CSV(t *testing.T) {
	gen := New()

	tmpl := `{{row "name" "age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
		},
	}

	out, err := gen.CSV(context.Background(), tmpl, data)
	if err != nil {
		t.Fatalf("CSV() error = %v", err)
	}
	if !strings.Contains(string(out), "Alice,30") {
		t.Fatalf("CSV() output = %q", string(out))
	}
}

func TestGenerator_CSVFromFile(t *testing.T) {
	gen := New()
	dir := t.TempDir()
	file := filepath.Join(dir, "people.csv.tmpl")
	tmpl := `{{row "name"}}{{range .People}}{{row .Name}}{{end}}`

	if err := os.WriteFile(file, []byte(tmpl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := gen.CSVFromFile(context.Background(), file, map[string]any{
		"People": []map[string]any{{"Name": "Bob"}},
	})
	if err != nil {
		t.Fatalf("CSVFromFile() error = %v", err)
	}
	if !strings.Contains(string(out), "Bob") {
		t.Fatalf("CSVFromFile() output = %q", string(out))
	}
}

func TestGenerator_Excel(t *testing.T) {
	gen := New()
	tmpl := `{{sheet "Users"}}{{row "name"}}{{range .Users}}{{row .Name}}{{end}}`

	out, err := gen.Excel(context.Background(), tmpl, map[string]any{
		"Users": []map[string]any{{"Name": "Alice"}},
	})
	if err != nil {
		t.Fatalf("Excel() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Excel() returned empty bytes")
	}
}

func TestGenerator_PDF_LightMode(t *testing.T) {
	gen := New(WithPDFRenderMode(PDFRenderLight))
	out, err := gen.PDF(context.Background(), `<p>x</p>`, nil)
	if err != nil {
		t.Fatalf("PDF() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("PDF() returned empty bytes")
	}
}

func TestGenerator_PDFFromFile(t *testing.T) {
	gen := New()
	dir := t.TempDir()
	file := filepath.Join(dir, "hello.html.tmpl")
	tmpl := `<h1>Hello {{.Name}}</h1>`

	if err := os.WriteFile(file, []byte(tmpl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := gen.PDFFromFile(context.Background(), file, map[string]any{"Name": "Alice"})
	if err != nil {
		t.Fatalf("PDFFromFile() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("PDFFromFile() returned empty bytes")
	}
}

func TestGenerator_CSVFromFile_NotFound(t *testing.T) {
	gen := New()
	_, err := gen.CSVFromFile(context.Background(), "/not/exist.tmpl", nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
