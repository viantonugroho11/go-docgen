package docgen

import (
	"context"
	"testing"
)

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
