package godocgen

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew_AppliesTimeoutOption(t *testing.T) {
	exp, ok := New(WithTimeout(2 * time.Second)).(*exporter)
	if !ok {
		t.Fatal("expected concrete *exporter type")
	}
	if exp.cfg.Timeout != 2*time.Second {
		t.Fatalf("timeout = %v, want %v", exp.cfg.Timeout, 2*time.Second)
	}
}

func TestToCSVTemplate(t *testing.T) {
	exp := New()

	tmpl := `{{row "name" "age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
		},
	}

	out, err := exp.ToCSVTemplate(context.Background(), tmpl, data)
	if err != nil {
		t.Fatalf("ToCSVTemplate() error = %v", err)
	}
	if !strings.Contains(string(out), "Alice,30") {
		t.Fatalf("ToCSVTemplate() output = %q", string(out))
	}
}

func TestToCSVFromFile(t *testing.T) {
	exp := New()
	dir := t.TempDir()
	file := filepath.Join(dir, "people.csv.tmpl")
	tmpl := `{{row "name"}}{{range .People}}{{row .Name}}{{end}}`

	if err := os.WriteFile(file, []byte(tmpl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := exp.ToCSVFromFile(context.Background(), file, map[string]any{
		"People": []map[string]any{{"Name": "Bob"}},
	})
	if err != nil {
		t.Fatalf("ToCSVFromFile() error = %v", err)
	}
	if !strings.Contains(string(out), "Bob") {
		t.Fatalf("ToCSVFromFile() output = %q", string(out))
	}
}

func TestToExcelTemplate(t *testing.T) {
	exp := New()
	tmpl := `{{sheet "Users"}}{{row "name"}}{{range .Users}}{{row .Name}}{{end}}`

	out, err := exp.ToExcelTemplate(context.Background(), tmpl, map[string]any{
		"Users": []map[string]any{{"Name": "Alice"}},
	})
	if err != nil {
		t.Fatalf("ToExcelTemplate() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ToExcelTemplate() returned empty bytes")
	}
}

func TestToPDFFromFile(t *testing.T) {
	exp := New()
	dir := t.TempDir()
	file := filepath.Join(dir, "hello.html.tmpl")
	tmpl := `<h1>Hello {{.Name}}</h1>`

	if err := os.WriteFile(file, []byte(tmpl), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	out, err := exp.ToPDFFromFile(context.Background(), file, map[string]any{"Name": "Alice"})
	if err != nil {
		t.Fatalf("ToPDFFromFile() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("ToPDFFromFile() returned empty bytes")
	}
}

func TestFromFile_NotFound(t *testing.T) {
	exp := New()
	_, err := exp.ToCSVFromFile(context.Background(), "/not/exist.tmpl", nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
