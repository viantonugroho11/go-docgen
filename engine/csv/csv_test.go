package csv

import (
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	tmpl := `{{row "name" "age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
			{"Name": "Bob", "Age": 25},
		},
	}

	rows, err := Build(tmpl, data)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("Build() rows length = %d, want %d", len(rows), 3)
	}
	if rows[1][0] != "Alice" || rows[2][0] != "Bob" {
		t.Fatalf("Build() unexpected rows = %#v", rows)
	}
}

func TestEngineGenerate(t *testing.T) {
	e := New()

	out, err := e.Generate([][]string{
		{"name", "age"},
		{"Alice", "30"},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "name,age") {
		t.Fatalf("Generate() output = %q, expected CSV header", got)
	}
}

func TestBuild_InvalidTemplate(t *testing.T) {
	if _, err := Build(`{{row "a"`, nil); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
