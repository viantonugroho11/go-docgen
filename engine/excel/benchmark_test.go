package excel

import "testing"

func benchmarkSheets() []Sheet {
	rows := make([][]string, 0, 101)
	rows = append(rows, []string{"Name", "Role"})
	for i := 0; i < 100; i++ {
		rows = append(rows, []string{"User", "Member"})
	}

	return []Sheet{
		{
			Name: "Users",
			Rows: rows,
		},
	}
}

func BenchmarkBuild(b *testing.B) {
	tmpl := `{{sheet "Users"}}{{row "Name" "Role"}}{{range .Users}}{{row .Name .Role}}{{end}}`
	data := map[string]any{
		"Users": []map[string]any{
			{"Name": "Alice", "Role": "Admin"},
			{"Name": "Bob", "Role": "Viewer"},
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
	sheets := benchmarkSheets()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Generate(sheets); err != nil {
			b.Fatalf("Generate() error = %v", err)
		}
	}
}
