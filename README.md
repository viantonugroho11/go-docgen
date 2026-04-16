# go-docgen

`go-docgen` is a lightweight Go library to generate documents from templates:

- PDF from HTML templates
- CSV from text templates
- Excel (XLSX) from text templates

It is designed for backend/reporting use cases where template + data in, bytes out.

## Installation

```bash
go get github.com/viantonugroho11/go-docgen
```

## Why use go-docgen

- Unified API for `PDF`, `CSV`, and `Excel`
- Works with in-memory templates and file-based templates
- Small public surface area (`New`, `To*Template`, `To*FromFile`)
- Extensible internals split by engine (`engine/csv`, `engine/excel`, `engine/pdf`)

## Core API

Create exporter:

```go
exp := godocgen.New()
```

Optional config:

```go
exp := godocgen.New(godocgen.WithTimeout(5 * time.Second))
```

Main methods:

- `ToPDFTemplate(ctx, tmpl, data)`
- `ToPDFFromFile(ctx, path, data)`
- `ToCSVTemplate(ctx, tmpl, data)`
- `ToCSVFromFile(ctx, path, data)`
- `ToExcelTemplate(ctx, tmpl, data)`
- `ToExcelFromFile(ctx, path, data)`

## Usage Examples

### CSV Example

```go
csvTpl := `{{row "Name" "Age"}}{{range .People}}{{row .Name .Age}}{{end}}`
csvData := map[string]any{
	"People": []map[string]any{
		{"Name": "Alice", "Age": 30},
		{"Name": "Bob", "Age": 25},
	},
}

csvBytes, err := exp.ToCSVTemplate(context.Background(), csvTpl, csvData)
if err != nil {
	panic(err)
}
_ = os.WriteFile("people.csv", csvBytes, 0o644)
```

### Excel Example

```go
excelTpl := `
{{sheet "Users"}}
{{row "Name" "Role"}}
{{range .Users}}{{row .Name .Role}}{{end}}
`
excelData := map[string]any{
	"Users": []map[string]any{
		{"Name": "Alice", "Role": "Admin"},
		{"Name": "Bob", "Role": "Viewer"},
	},
}

excelBytes, err := exp.ToExcelTemplate(context.Background(), excelTpl, excelData)
if err != nil {
	panic(err)
}
_ = os.WriteFile("users.xlsx", excelBytes, 0o644)
```

### PDF Example

```go
pdfTpl := `<h1>Hello {{.Name}}</h1><p>Welcome to go-docgen.</p>`
pdfBytes, err := exp.ToPDFTemplate(context.Background(), pdfTpl, map[string]any{"Name": "Alice"})
if err != nil {
	panic(err)
}
_ = os.WriteFile("hello.pdf", pdfBytes, 0o644)
```

### File-based Template Example

```go
bytes, err := exp.ToCSVFromFile(context.Background(), "templates/report.csv.tmpl", data)
if err != nil {
	panic(err)
}
```

## Template Helpers

### CSV Helpers

- `row ...any`: append one CSV row

```gotemplate
{{row "col1" "col2"}}
{{range .Items}}
{{row .Name .Value}}
{{end}}
```

### Excel Helpers

- `sheet name`: create/select active sheet
- `row ...any`: append one row to active sheet

```gotemplate
{{sheet "Summary"}}
{{row "Metric" "Value"}}
{{row "Users" .TotalUsers}}
```

## Error Handling Notes

- Invalid template syntax returns template parse errors.
- Missing template files return file read errors.
- PDF rendering tries Chromium first, then falls back to a lightweight renderer.

## Testing

Run unit tests:

```bash
go test ./...
```

## Benchmark

Benchmark tests are included for CSV and Excel export flows.

Run all benchmarks:

```bash
go test -run='^$' -bench . -benchmem ./...
```

Run exporter-specific benchmarks only:

```bash
go test -run='^$' -bench BenchmarkExporter -benchmem .
```

Notes:

- PDF benchmark is intentionally omitted because it depends on external browser/runtime conditions and can produce unstable numbers across machines.

### Latest Benchmark Result (Sample)

Environment:

- `goos`: `darwin`
- `goarch`: `arm64`
- `cpu`: `Apple M2`

Command:

```bash
go test -run='^$' -bench . -benchmem ./...
```

Result summary:

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkExporter_ToCSVTemplate` | 89745 | 44758 | 1379 |
| `BenchmarkExporter_ToExcelTemplate` | 1464491 | 811963 | 7979 |
| `engine/csv.BenchmarkBuild` | 6172 | 5683 | 97 |
| `engine/csv.BenchmarkGenerate` | 3129 | 5040 | 3 |
| `engine/excel.BenchmarkBuild` | 7343 | 6267 | 112 |
| `engine/excel.BenchmarkGenerate` | 1288893 | 762705 | 6587 |

These numbers are machine-dependent. Use them as a baseline and compare against your own environment when optimizing.