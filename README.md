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
- Small public surface area (`docgen.New`, `Generator` methods: `PDF`, `CSV`, `Excel`, and `*FromFile` variants); PDF backend is selectable at construction time (`WithPDFRenderMode`)
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

Fix the PDF pipeline (default is Chromium then lightweight fallback):

```go
import (
	"time"

	"github.com/viantonugroho11/go-docgen"
)

gen := docgen.New(
	docgen.WithTimeout(5 * time.Second),
	docgen.WithPDFRenderMode(docgen.PDFRenderChromium), // or PDFRenderAuto, PDFRenderLight
)
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

## PDF engine and performance

`WithPDFRenderMode` is evaluated when you call `docgen.New` (not per `PDF` call):

| Mode | Behavior | Typical latency | Fidelity |
| --- | --- | --- | --- |
| `PDFRenderAuto` (default) | Chromium (chromedp) first; on failure, `gofpdf` text path | Dominated by Chromium when it succeeds (~0.5–2 s cold per run, machine-dependent) | Full HTML/CSS when Chromium succeeds |
| `PDFRenderChromium` | Chromium only; errors propagate | Same as Auto when Chromium succeeds | Full HTML/CSS |
| `PDFRenderLight` | `gofpdf` `MultiCell` only; no browser | Sub-millisecond to low milliseconds for small HTML strings | **Not** HTML layout—tags show as text; use only when that is acceptable |

For **apples-to-apples** wall times against wkhtmltopdf, Chrome CLI, and the light path, use `cmd/pdfcompare` (see below). **`Generator` micro-benchmarks** (CSV, Excel, PDF light, PDF Chromium) live in `export_benchmark_test.go` at the module root; lower-level PDF benches live in `engine/pdf`. Chromedp-backed benchmarks are opt-in (skipped under `go test -short`) because they are slow and machine-dependent.

## Error Handling Notes

- Invalid template syntax returns template parse errors.
- Missing template files return file read errors.
- PDF: with `PDFRenderAuto`, Chromium is tried first, then the lightweight path. With `PDFRenderChromium`, only Chromium is used. With `PDFRenderLight`, only the lightweight path is used.

## Testing

Run unit tests:

```bash
go test ./...
```

## Benchmark

Benchmark tests cover:

- **Root package** (`export_benchmark_test.go`): `BenchmarkGenerator_CSV`, `BenchmarkGenerator_Excel`, `BenchmarkGenerator_PDF` ( **`PDFRenderLight`** ), and `BenchmarkGenerator_PDF_Chromium` ( **`PDFRenderChromium`** ; **skipped under `-short`** ).
- **`engine/pdf`**: `BenchmarkRender_Light`, `BenchmarkRender_Light_Small`, plus chromedp benches **skipped under `-short`** (`BenchmarkRender_Chromium`, `BenchmarkRenderChromeDP`).

Run all benchmarks (recommended; skips slow Chromium PDF benches):

```bash
go test -short -run='^$' -bench . -benchmem ./...
```

Run all benchmarks including every chromedp PDF benchmark (slow; needs Chromium):

```bash
go test -run='^$' -bench . -benchmem ./...
```

Run generator-specific benchmarks only (includes PDF light; skips PDF Chromium while `-short` is set):

```bash
go test -short -run='^$' -bench BenchmarkGenerator -benchmem .
```

Measure **`Generator.PDF` + Chromium only** (same HTML/data as `BenchmarkGenerator_PDF`; not compatible with `-short`):

```bash
go test -run='^$' -bench=BenchmarkGenerator_PDF_Chromium -benchmem .
```

Notes:

- For **wall-clock** PDF comparisons (chromedp, Chrome CLI, wkhtmltopdf, gofpdf), use `cmd/pdfcompare` (see below).

### PDF vs other backends (local comparison)

`cmd/pdfcompare` is a **separate Go module** (with its own `go.mod`) so this optional CLI harness stays out of the main module’s dependency graph.

From the **repository root**:

```bash
go run -C cmd/pdfcompare .
```

Or:

```bash
cd cmd/pdfcompare && go run .
```

Common flags: `-runs 15 -warmup 3 -html /path/to/file.html`. Add `-nomem` to skip the extra allocation batch (no second line).

For each backend the tool prints **two lines**: wall-clock **median / mean / p95**, then **`ns/op`** (mean wall time in nanoseconds), **`B/op`** and **`allocs/op`** from `runtime.MemStats` (`TotalAlloc` / `Mallocs` delta, divided by `runs`) over a second batch of iterations. Those values are **Go heap only**; memory inside Chromium or WebKit processes is not counted, so they are comparable in spirit to `go test -benchmem` for the Go side only, not identical to it.

**`wkhtmltopdf`:** `pdfcompare` runs the **`wkhtmltopdf` subprocess** when the binary is on `PATH` (or set **`WKHTMLTOPDF_PATH`**). That compares against the WebKit/Qt-style stack shipped with typical wkhtml installers.

#### Sample PDF results (captured run)

Captured **2026-04-16** on `darwin` / `arm64` / Apple M2, `go1.23.4`. Command from repository root:

```bash
go run -C cmd/pdfcompare . -runs 3 -warmup 1
```

Bundled HTML fixture (~80-row table). **Rerun on your machine**; wall times shift with cache and load. **`wkhtmltopdf`** was on `PATH` as `/usr/local/bin/wkhtmltopdf`.

| Backend | median | mean | p95 | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | ---: | ---: | ---: |
| chromedp (`RenderChromeDP`) | 776ms | 776ms | 796ms | 775702638 | 1916728 | 2341 |
| go-docgen `PDFRenderAuto` | 766ms | 761ms | 768ms | 760544444 | 1955781 | 2401 |
| go-docgen `PDFRenderChromium` | 798ms | 793ms | 820ms | 793411819 | 1953800 | 2407 |
| go-docgen `PDFRenderLight` | 1ms | 1ms | 1ms | 848375 | 4983501 | 974 |
| Chrome CLI `--print-to-pdf` | 2.055s | 2.111s | 2.237s | 2111353583 | 16882 | 55 |
| `wkhtmltopdf` CLI (subprocess) | 403ms | 407ms | 419ms | 407107486 | 21245 | 109 |
| gofpdf `MultiCell` only | 1ms | 1ms | 1ms | 983749 | 4952437 | 930 |

`ns/op` is mean wall time per iteration (nanoseconds). `B/op` and `allocs/op` are Go `runtime.MemStats` averages over a second batch (see tool banner text). Generator rows include template + `docgen.New` each iteration in this harness, so `B/op` can exceed the bare `gofpdf` row even when wall time is tiny for Light mode.

<details>
<summary>Raw `pdfcompare` output (same run)</summary>

```
PDF compare — same HTML input, wall time + Go runtime allocation shape
runs=3 warmup=1 nomem=false go=go1.23.4 os=darwin/arm64
Line 1: median/mean/p95 wall time per iteration.
Line 2: ns/op = mean wall nanoseconds; B/op & allocs/op = (TotalAlloc, Mallocs delta) / runs over a fresh batch (Go heap only — not Chrome/wkhtml native heaps).
For library micro-benchmarks (CSV/Excel/PDF generator + engine/pdf) see README "Latest benchmark result".
If chromedp SKIP but go-docgen Auto is very fast, Auto used gofpdf fallback, not Chromium.

chromedp only (engine/pdf.RenderChromeDP, no fallback)      median=     776ms  mean=     776ms  p95=     796ms
                                                            ns/op=775702638  B/op=1916728  allocs/op=2341
go-docgen Generator.PDF (PDFRenderAuto)                     median=     766ms  mean=     761ms  p95=     768ms
                                                            ns/op=760544444  B/op=1955781  allocs/op=2401
go-docgen Generator.PDF (PDFRenderChromium)                 median=     798ms  mean=     793ms  p95=     820ms
                                                            ns/op=793411819  B/op=1953800  allocs/op=2407
go-docgen Generator.PDF (PDFRenderLight)                    median=       1ms  mean=       1ms  p95=       1ms
                                                            ns/op=848375  B/op=4983501  allocs/op=974
Chrome/Chromium CLI (--headless --print-to-pdf)             median=    2.055s  mean=    2.111s  p95=    2.237s
                                                            ns/op=2111353583  B/op=16882  allocs/op=55
wkhtmltopdf CLI (subprocess)                                median=     403ms  mean=     407ms  p95=     419ms
                                                            ns/op=407107486  B/op=21245  allocs/op=109
gofpdf MultiCell (same idea as engine/pdf light fallback)   median=       1ms  mean=       1ms  p95=       1ms
                                                            ns/op=983749  B/op=4952437  allocs/op=930
```

</details>

#### Why wkhtmltopdf is often faster than Chromium in this benchmark

On the captured run, **`wkhtmltopdf` (~407 ms mean)** sits between **chromedp (~776 ms)** and **Chrome `--print-to-pdf` (~2.11 s)** for the same fixture.

- **Smaller stack:** wkhtml’s WebKit/Qt path is lighter than a full Chrome feature set used with CDP `PrintToPDF`.
- **Less protocol overhead:** chromedp talks to Chromium over DevTools (WebSocket/CDP), then `PrintToPDF` — more steps and synchronization than spawning `wkhtmltopdf` with a `file://` input.
- **Chrome CLI** typically spawns a new process and user data directory per invocation, similar to cold-start cost.
- **gofpdf** almost always "wins" on wall time because it is **not** an HTML renderer: it writes raw text to the PDF instead of parsing DOM/CSS.

Warning: wkhtmltopdf is **no longer maintained** upstream; Chromium is more modern for CSS and web features. The numbers above help explain **performance cost**, not a single-product recommendation.

### Performance optimizations (in code)

- **HTML / text templates** (`template` package): parsed templates are cached by source string and cloned per render, which speeds up repeated PDF HTML rendering when the template string is reused. Choosing `PDFRenderLight` avoids Chromium entirely when plain-text PDF output is enough.
- **Chromium PDF** (`engine/pdf`): HTML is applied with `Page.setDocumentContent` on `about:blank` instead of navigating a `data:` URL built with `url.PathEscape`, avoiding an extra full-size copy of the HTML string on the Go heap each render (Chromium’s own RSS is unchanged).
- **CSV / Excel row helpers** (`engine/csv`, `engine/excel`): cell values use `internal/strfmt.FormatAny` with fast paths for common scalar types instead of always using `fmt.Sprintf`, which reduces allocations in hot loops.
- **Excel writer** (`engine/excel`): rows are written with `SetSheetRow` instead of one `SetCellValue` per cell (fewer high-level calls for the same data).
- **CSV / Excel `text/template` parse**: `text/template` requires `Funcs` to be registered **before** `Parse`, and row/sheet helpers close over per-run state, so we **cannot** safely reuse a single parsed template the same way as HTML. Further gains there would need a different API (for example accepting a pre-parsed template or a row sink on `data`).

### Latest benchmark result (sample)

Environment (single run, **2026-04-16**):

- `goos`: `darwin`
- `goarch`: `arm64`
- `cpu`: `Apple M2`

Command (use **`-short`** so slow chromedp PDF benchmarks are skipped; they still exist for manual runs):

```bash
go test -short -run='^$' -bench . -benchmem ./...
```

Result summary:

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkGenerator_PDF` (`PDFRenderLight`) | 508298 | 2560702 | 2508 |
| `BenchmarkGenerator_CSV` | 78645 | 44015 | 1197 |
| `BenchmarkGenerator_Excel` | 1406753 | 807028 | 8081 |
| `engine/csv.BenchmarkBuild` | 6716 | 5659 | 91 |
| `engine/csv.BenchmarkGenerate` | 3205 | 5040 | 3 |
| `engine/excel.BenchmarkBuild` | 7467 | 6235 | 106 |
| `engine/excel.BenchmarkGenerate` | 1593084 | 750906 | 6891 |
| `engine/pdf.BenchmarkRender_Light` | 466855 | 3731207 | 712 |
| `engine/pdf.BenchmarkRender_Light_Small` | 137595 | 1233915 | 196 |

`BenchmarkGenerator_PDF_Chromium` and the chromedp benches in `engine/pdf` are skipped when `-short` is set. Sample for **`BenchmarkGenerator_PDF_Chromium`** on the same machine (two iterations; wall time dominates):

| Benchmark | ns/op | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| `BenchmarkGenerator_PDF_Chromium` | 713155771 | 988136 | 4436 |

```bash
go test -run='^$' -bench=BenchmarkGenerator_PDF_Chromium -benchtime=2x -benchmem .
```

To measure raw `engine/pdf` chromedp benches locally (slow, needs Chromium):

```bash
go test -run='^$' -bench 'BenchmarkRender_Chromium|BenchmarkRenderChromeDP' -benchtime=3x -benchmem ./engine/pdf/
```

These numbers are machine-dependent. Use them as a baseline and compare against your own environment when optimizing.