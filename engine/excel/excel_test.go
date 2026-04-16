package excel

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestBuild(t *testing.T) {
	tmpl := `{{sheet "Users"}}{{row "name" "role"}}{{range .Users}}{{row .Name .Role}}{{end}}`
	data := map[string]any{
		"Users": []map[string]any{
			{"Name": "Alice", "Role": "Admin"},
			{"Name": "Bob", "Role": "Viewer"},
		},
	}

	sheets, err := Build(tmpl, data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(sheets) != 1 {
		t.Fatalf("Build() sheets length = %d, want 1", len(sheets))
	}
	if sheets[0].Name != "Users" {
		t.Fatalf("Build() sheet name = %q, want %q", sheets[0].Name, "Users")
	}
	if len(sheets[0].Rows) != 3 {
		t.Fatalf("Build() rows length = %d, want 3", len(sheets[0].Rows))
	}
}

func TestEngineGenerate(t *testing.T) {
	e := New()
	bytesOut, err := e.Generate([]Sheet{
		{
			Name: "Summary",
			Rows: [][]string{
				{"metric", "value"},
				{"users", "10"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(bytesOut))
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}

	cellA1, err := f.GetCellValue("Summary", "A1")
	if err != nil {
		t.Fatalf("GetCellValue() error = %v", err)
	}
	if cellA1 != "metric" {
		t.Fatalf("A1 = %q, want %q", cellA1, "metric")
	}
}

func TestBuild_InvalidTemplate(t *testing.T) {
	if _, err := Build(`{{sheet "Users"`, nil); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
