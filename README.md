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

### PDF vs other backends (local comparison)

`cmd/pdfcompare` is a **separate Go module** (with its own `go.mod`) so the **libwkhtmltox** dependency (CGO, via `github.com/adrg/go-wkhtmltopdf`) does not enter the main module's import graph.

From the **repository root**:

```bash
go run -C cmd/pdfcompare .
```

Or:

```bash
cd cmd/pdfcompare && go run .
```

Common flags: `-runs 15 -warmup 3 -html /path/to/file.html`.

**libwkhtmltox (not the `wkhtmltopdf` binary):** calls go through Go import + CGO. You need `wkhtmltox` *headers* and *library* that match the CPU architecture (for example arm64 vs amd64). Example:

```bash
cd cmd/pdfcompare
CGO_ENABLED=1 go run -tags libwkhtmltox .
```

Without that build tag, the libwkhtmltox path is skipped with a message; Chromium and Chrome CLI paths are still benchmarked. The `engine/pdf.RenderChromeDP` function remains available to benchmark chromedp alone (without gofpdf fallback).

#### Sample PDF results (median wall-clock, one machine)

Same environment as the CSV/Excel table below (`darwin` / `arm64` / Apple M2). Command: `go run -C cmd/pdfcompare . -runs 1 -warmup 0` on the bundled fixture (~80-row HTML table). **This is illustrative only; rerun on your machine.**

| Backend | Median (sample) | Brief notes |
| --- | ---: | --- |
| chromedp (`RenderChromeDP`) | ~1.6 s | CDP + `PrintToPDF`, Chromium process |
| go-docgen `Generator.PDF` | ~0.9–1.6 s | HTML template + chromedp (variance across runs) |
| Chrome CLI `--print-to-pdf` | ~2.0 s | Spawns a new Chrome process per iteration |
| libwkhtmltox (`adrg/go-wkhtmltopdf`) | ~0.4–0.5 s | WebKit/Qt in-process, no subprocess |
| gofpdf `MultiCell` (text fallback) | ~2 ms | Not HTML layout; baseline only |

#### Why libwkhtmltox / wkhtml is often faster than Chromium in this benchmark

- **Smaller stack:** `wkhtmltox` uses an embedded WebKit/Qt render engine in one process; you do not need to run a full-featured Chrome browser.
- **Less protocol overhead:** chromedp talks to Chromium over DevTools (WebSocket/CDP), navigates a `data:` URL, then `PrintToPDF` — more steps and synchronization than libwkhtmltox's native conversion path.
- **Chrome CLI** typically spawns a new process and user data directory per invocation, similar to cold-start cost; that often loses to a single library converting HTML straight into a PDF buffer.
- **gofpdf** almost always "wins" on the numbers because it is **not** an HTML renderer: it writes raw text to the PDF instead of parsing DOM/CSS.

Warning: wkhtmltopdf / libwkhtmltox is **no longer maintained** upstream; Chromium is more modern for CSS and web features. The numbers above help explain **performance cost**, not a single-product recommendation.

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