# go-docgen

Fast, batteries-included Go library for generating documents from templates.

- **PDF** from HTML templates (headless Chromium via chromedp, or lightweight `gofpdf` fallback)
- **CSV** from text templates
- **Excel (XLSX)** from text templates

Designed for backend/reporting use cases: template + data in, bytes out. Persistent browser + tab pool + result cache — steady-state PDF cache hit is ~4 µs.

**Requires Go 1.25+.**

---

## Table of Contents

1. [Install](#install)
2. [Quick start](#quick-start)
3. [Configuration options](#configuration-options)
4. [Template helpers](#template-helpers)
5. [PDF backends](#pdf-backends)
6. [Performance & benchmarks](#performance--benchmarks)
7. [Deployment: shrink your container image](#deployment-shrink-your-container-image)
8. [pdfcompare CLI (side-by-side backend comparison)](#pdfcompare-cli-side-by-side-backend-comparison)
9. [Testing](#testing)
10. [Contributing](#contributing)

---

## Install

```bash
go get github.com/viantonugroho11/go-docgen
```

Chromium (or Chrome) must be installed on the host for the PDF Chromium path. For the smallest footprint, use [`chrome-headless-shell`](#deployment-shrink-your-container-image) (~80 MB) instead of full Chrome (~200 MB).

---

## Quick start

```go
package main

import (
	"context"
	"os"

	"github.com/viantonugroho11/go-docgen"
)

func main() {
	gen := docgen.New()

	pdf, err := gen.PDF(context.Background(),
		`<h1>Hello {{.Name}}</h1><p>Welcome.</p>`,
		map[string]any{"Name": "Alice"},
	)
	if err != nil {
		panic(err)
	}
	_ = os.WriteFile("hello.pdf", pdf, 0o644)
}
```

Production-tuned generator (cache + prewarm + slim binary):

```go
gen := docgen.New(
	docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
	docgen.WithPDFMaxConcurrency(8),
	docgen.WithPDFCacheSize(256),
	docgen.WithPDFPrewarm(true),
	docgen.WithPDFChromePath("/opt/chrome-headless-shell/chrome-headless-shell"),
	docgen.WithTimeout(15 * time.Second),
)
```

### Generator API

| Method | Input | Output |
|---|---|---|
| `PDF(ctx, template, data)` | HTML string | PDF bytes |
| `PDFFromFile(ctx, path, data)` | HTML file path | PDF bytes |
| `CSV(ctx, template, data)` | text template | CSV bytes |
| `CSVFromFile(ctx, path, data)` | text template file | CSV bytes |
| `Excel(ctx, template, data)` | text template | XLSX bytes |
| `ExcelFromFile(ctx, path, data)` | text template file | XLSX bytes |

---

## Configuration options

All options passed to `docgen.New(opts...)`. Every option is optional; zero-value config yields sensible defaults.

| Option | Type | Default | What it does |
|---|---|---|---|
| `WithTimeout(d)` | `time.Duration` | `10s` | Per-render deadline. |
| `WithPDFRenderMode(m)` | `PDFRenderMode` | `PDFRenderAuto` | Selects PDF backend at construction time. See [PDF backends](#pdf-backends). |
| `WithPDFMaxConcurrency(n)` | `int` | `GOMAXPROCS` | Sizes the tab pool. Also caps concurrent Chromium renders per generator. |
| `WithPDFChromePath(path)` | `string` | *(chromedp searches PATH)* | Path to a Chromium-family binary. Point at `chrome-headless-shell` for slim images. |
| `WithPDFExtraFlags(flags...)` | `[]string` | none | Extra Chromium argv flags. Each entry is `"flag"` (bool true) or `"flag=value"`. |
| `WithPDFCacheSize(n)` | `int` | `0` (disabled) | LRU cache of rendered PDFs keyed by `sha256(html)`. Cache hit ≈ 4 µs. |
| `WithPDFPrewarm(enable)` | `bool` | `false` | Boot Chromium + pool tabs in a background goroutine at `New()`. Eliminates first-render cold start. |

---

## Template helpers

Templates use Go's [`text/template`](https://pkg.go.dev/text/template) (CSV / Excel) or [`html/template`](https://pkg.go.dev/html/template) (PDF).

### CSV

- `row ...any` — append one row.

```gotemplate
{{row "Name" "Age"}}
{{range .People}}{{row .Name .Age}}{{end}}
```

### Excel

- `sheet name` — create/select the active sheet.
- `row ...any` — append one row to the active sheet.

```gotemplate
{{sheet "Summary"}}
{{row "Metric" "Value"}}
{{row "Users" .TotalUsers}}
{{sheet "Detail"}}
{{range .Rows}}{{row .Name .Value}}{{end}}
```

### PDF

Any valid HTML/CSS. Standard `html/template` action syntax for interpolation:

```gotemplate
<h1>Invoice #{{.Number}}</h1>
<table>
  {{range .LineItems}}
  <tr><td>{{.SKU}}</td><td>{{.Qty}}</td><td>{{.Total}}</td></tr>
  {{end}}
</table>
```

---

## PDF backends

Set once at `docgen.New` via `WithPDFRenderMode`. Not per-call.

| Mode | Behaviour | Fidelity | Typical latency (warm) |
|---|---|---|---|
| `PDFRenderAuto` *(default)* | Try Chromium; on failure fall back to `gofpdf` text path. | Full HTML/CSS when Chromium succeeds; text dump otherwise. | ~500-700 ms Chromium; < 1 ms fallback |
| `PDFRenderChromium` | Chromium only. Errors propagate. | Full HTML/CSS. | ~500-700 ms |
| `PDFRenderLight` | `gofpdf.MultiCell` only. HTML tags render as literal text. | **Not** an HTML engine. | Sub-millisecond |

### PDF pipeline internals (v0.3.0)

```
Render(html)
  │
  ├─ cache.get(sha256(html))  ── hit ─► return cached bytes (≈ 4 µs)
  │        miss ▼
  │
  ├─ ensureBrowser()  (persistent Chromium; started once, reused)
  │
  ├─ pool.checkout()  ── nil slot → materialise tab
  │                    live tab   → reuse (skips ~200-500 ms Target.create)
  │
  ├─ SetDocumentContent(html) + PrintToPDF()
  │
  ├─ pool.return()    (tab kept alive for next call)
  │
  └─ cache.put(bytes) ── evict LRU if full
```

`Close()` on the returned `Generator` (via `io.Closer` type assertion) shuts the browser down deterministically. After `Close`, the next render restarts it.

---

## Performance & benchmarks

### Headline numbers (Apple M2, `go1.25.0`, `-benchtime=50x`)

| Bench | ns/op | Notes |
|---|--:|---|
| `Render_Chromium_CacheHit` | **~4 µs** | Cache hit; no browser round-trip |
| `Render_Light_Small` | ~195 µs | `gofpdf` text path |
| `Render_Light` (medium) | ~541 µs | `gofpdf` text path |
| `Wkhtmltopdf` (ref) | ~490 ms | External subprocess, reference |
| `Render_Chromium` (small) | ~635 ms | Chromium miss, warm browser |
| `Render_Chromium_Medium` | ~640 ms | Chromium miss, warm browser |

### Real-world aggregate (invoice/statement workload, 80% cache hit rate)

| Backend | Effective avg latency |
|---|--:|
| **go-docgen (`WithPDFCacheSize(256)`)** | **~127 ms** |
| wkhtml (no cache; fresh subprocess each call) | ~490 ms |

Above **~15% cache hit rate**, go-docgen wins aggregate latency vs wkhtml. Real-world hit rates for structured docs (invoice, receipt, statement) are typically 40-90%.

### Run the benchmarks

```bash
# Fast set (skip chromedp)
go test -short -run='^$' -bench . -benchmem ./...

# Include Chromium benches (needs Chrome on PATH)
go test -run='^$' -bench . -benchmem ./engine/pdf/

# Wkhtml reference (needs wkhtmltopdf on PATH)
go test -run='^$' -bench=Wkhtml -benchmem ./engine/pdf/
```

---

## Deployment: shrink your container image

Full Chrome/Chromium is ~200 MB installed. Chrome team ships **`chrome-headless-shell`** — the Blink-only rendering shell used by Puppeteer — at ~80 MB with byte-equivalent PDF output.

### Dockerfile snippet

```dockerfile
# Build stage: fetch chrome-headless-shell
FROM node:20-slim AS chrome
RUN npx --yes @puppeteer/browsers install chrome-headless-shell@stable \
    && mv chrome-headless-shell /opt/chrome-headless-shell

# Runtime stage
FROM gcr.io/distroless/base-debian12
COPY --from=chrome /opt/chrome-headless-shell /opt/chrome-headless-shell
COPY app /app
ENV CHROME_PATH=/opt/chrome-headless-shell/chrome-headless-shell
ENTRYPOINT ["/app"]
```

```go
gen := docgen.New(
	docgen.WithPDFChromePath(os.Getenv("CHROME_PATH")),
)
```

Result: container image ~120 MB smaller with no rendering-quality regression.

### Alpine alternative

```dockerfile
FROM alpine:3.20
RUN apk add --no-cache chromium
ENV CHROME_PATH=/usr/bin/chromium
```

~120 MB total for chromium package. Slight font-rendering diffs vs Chrome; usually acceptable.

---

## pdfcompare CLI (side-by-side backend comparison)

`cmd/pdfcompare` is a **separate Go module** (own `go.mod`) so its optional deps stay out of the library's dependency graph. It measures wall-clock time and Go-heap allocations across every PDF backend for the same fixture.

From repo root:

```bash
go run -C cmd/pdfcompare . -runs 15 -warmup 3
```

Or against your own HTML:

```bash
go run -C cmd/pdfcompare . -html ./sample.html -runs 10
```

Flags: `-runs N`, `-warmup N`, `-html PATH`, `-nomem` (skip alloc pass). Set `WKHTMLTOPDF_PATH` to point at a non-`PATH` wkhtml binary.

Each backend prints two lines: median/mean/p95 wall time, then `ns/op` + `B/op` + `allocs/op` (Go heap only; native Chrome/wkhtml RSS not counted).

---

## Testing

```bash
# Full suite (short mode: skips slow Chromium benches)
go test ./... -short

# Chromium PDF tests included
go test ./engine/pdf/... -run TestRender
```

CI-friendly:

```bash
go test -short -race ./...
```

---

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md). Please:

- Add tests for changes to template parsing or output formats.
- Run `go test ./... -short` before PR.
- For perf-sensitive changes, include before/after `go test -bench` output in the PR description.
- Keep template assets and rendering logic separated from consumer apps.

## License

See [`LICENSE`](LICENSE).

## Changelog

Per-version release notes in [`CHANGELOG.md`](CHANGELOG.md). Highlights:

- **v0.3.0** — LRU cache + browser prewarm. Cache hit ~4 µs; beats wkhtml on aggregate latency above ~15% hit rate.
- **v0.2.1** — Security patch (x/crypto, x/net, excelize CVEs). Requires Go 1.25+.
- **v0.2.0** — Tab pool, `chrome-headless-shell` support (`WithPDFChromePath`), HTML template Clone dropped, wkhtml reference benches.
- **v0.1.1** — Persistent browser, bounded concurrency, buffer pools, CSV/Excel template cache.
