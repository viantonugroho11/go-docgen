package csv

import "testing"

func benchmarkRows() [][]string {
	rows := make([][]string, 0, 100)
	rows = append(rows, []string{"Name", "Age"})
	for i := 0; i < 100; i++ {
		rows = append(rows, []string{"User", "30"})
	}
	return rows
}

func BenchmarkBuild(b *testing.B) {
	tmpl := `{{row "Name" "Age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
			{"Name": "Bob", "Age": 25},
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Build(tmpl, data); err != nil {
			b.Fatalf("Build() error = %v", err)
		}
	}
}

func BenchmarkGenerate(b *testing.B) {
	e := New()
	rows := benchmarkRows()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Generate(rows); err != nil {
			b.Fatalf("Generate() error = %v", err)
		}
	}
}
