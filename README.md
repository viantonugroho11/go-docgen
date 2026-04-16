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
- Small public surface area (`docgen.New`, `Generator` methods: `PDF`, `CSV`, `Excel`, and `*FromFile` variants)
- Extensible internals split by engine (`engine/csv`, `engine/excel`, `engine/pdf`)

## Core API

The library root package is `docgen` (same module path: `github.com/viantonugroho11/go-docgen`).

Create a generator:

```go
import "github.com/viantonugroho11/go-docgen"

gen := docgen.New()
```

Optional config:

```go
gen := docgen.New(docgen.WithTimeout(5 * time.Second))
```

`Generator` methods:

- `PDF(ctx, template, data)` — HTML template → PDF bytes
- `PDFFromFile(ctx, path, data)` — template file → PDF bytes
- `CSV(ctx, template, data)` — text template → CSV bytes
- `CSVFromFile(ctx, path, data)`
- `Excel(ctx, template, data)` — text template → XLSX bytes
- `ExcelFromFile(ctx, path, data)`

## Usage Examples

Assume:

```go
import (
	"context"
	"os"

	"github.com/viantonugroho11/go-docgen"
)

gen := docgen.New()
```

### CSV Example

```go
csvTpl := `{{row "Name" "Age"}}{{range .People}}{{row .Name .Age}}{{end}}`
csvData := map[string]any{
	"People": []map[string]any{
		{"Name": "Alice", "Age": 30},
		{"Name": "Bob", "Age": 25},
	},
}

csvBytes, err := gen.CSV(context.Background(), csvTpl, csvData)
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

excelBytes, err := gen.Excel(context.Background(), excelTpl, excelData)
if err != nil {
	panic(err)
}
_ = os.WriteFile("users.xlsx", excelBytes, 0o644)
```

### PDF Example

```go
pdfTpl := `<h1>Hello {{.Name}}</h1><p>Welcome to go-docgen.</p>`
pdfBytes, err := gen.PDF(context.Background(), pdfTpl, map[string]any{"Name": "Alice"})
if err != nil {
	panic(err)
}
_ = os.WriteFile("hello.pdf", pdfBytes, 0o644)
```

### File-based Template Example

```go
bytes, err := gen.CSVFromFile(context.Background(), "templates/report.csv.tmpl", data)
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

Run generator-specific benchmarks only:

```bash
go test -run='^$' -bench BenchmarkGenerator -benchmem .
```

Notes:

- PDF benchmark is intentionally omitted because it depends on external browser/runtime conditions and can produce unstable numbers across machines.

### Performance optimizations (in code)

- **HTML / text templates** (`template` package): parsed templates are cached by source string and cloned per render, which speeds up repeated PDF HTML rendering when the template string is reused.
- **CSV / Excel row helpers** (`engine/csv`, `engine/excel`): cell values use `internal/strfmt.FormatAny` with fast paths for common scalar types instead of always using `fmt.Sprintf`, which reduces allocations in hot loops.
- **Excel writer** (`engine/excel`): rows are written with `SetSheetRow` instead of one `SetCellValue` per cell (fewer high-level calls for the same data).
- **CSV / Excel `text/template` parse**: `text/template` requires `Funcs` to be registered **before** `Parse`, and row/sheet helpers close over per-run state, so we **cannot** safely reuse a single parsed template the same way as HTML. Further gains there would need a different API (for example accepting a pre-parsed template or a row sink on `data`).

### Latest benchmark result (sample, after optimizations)

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
| `BenchmarkGenerator_CSV` | 79312 | 44013 | 1197 |
| `BenchmarkGenerator_Excel` | 1407766 | 804018 | 8081 |
| `engine/csv.BenchmarkBuild` | 6552 | 5659 | 91 |
| `engine/csv.BenchmarkGenerate` | 3108 | 5040 | 3 |
| `engine/excel.BenchmarkBuild` | 7266 | 6236 | 106 |
| `engine/excel.BenchmarkGenerate` | 1337422 | 743901 | 6891 |

These numbers are machine-dependent. Use them as a baseline and compare against your own environment when optimizing.