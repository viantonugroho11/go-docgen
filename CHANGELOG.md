# Changelog

All notable changes to go-docgen are documented here.

---

## [0.3.0] — 2026-08-12 — LRU cache + browser prewarm

### Summary

Two new options that decisively beat wkhtml on aggregate latency for real
workloads:

| Bench (Apple M2, `-benchtime=50x`) | ns/op |
|---|--:|
| `Render_Chromium_CacheHit` | **4 164 ns** (~4 µs) |
| `Render_Wkhtmltopdf` (ref) | 536 370 050 ns (~536 ms) |

Cache hit is **~130 000× faster than wkhtml**. Invoice/statement/receipt
workloads with content repetition now win on p50 latency, not just
concurrent throughput.

### New Options

#### `WithPDFCacheSize(n int) Option`

Enable in-memory LRU cache of rendered PDFs, keyed by `sha256(html)`. A hit
skips the entire render pipeline. Zero (default) disables.

```go
gen := docgen.New(
    docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
    docgen.WithPDFCacheSize(256),
)
```

Recommended: 64–512 for repeat workloads; 0 for unique-per-call content.

#### `WithPDFPrewarm(enable bool) Option`

Boot Chromium and materialise every pooled tab at construction time (in a
background goroutine) so the first `PDF()` call does not pay the ~200-500
ms cold-start cost.

```go
gen := docgen.New(
    docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
    docgen.WithPDFPrewarm(true),
)
```

### Safety

Cache values are defensive copies on both `put` and `get` — callers cannot
mutate cached bytes across calls. Test coverage in `engine/pdf/cache_test.go`.

### Compatibility

No API break. Zero-value config disables both features.

---

## [0.2.1] — 2026-08-12 — Security patch

Fixes 44 open Dependabot alerts across root and `cmd/pdfcompare` modules.

| Package | From | To | CVEs |
|---|--:|--:|---|
| `golang.org/x/crypto` | v0.19.0 | **v0.55.0** | CVE-2024-45337, CVE-2025-22869, CVE-2025-47914, CVE-2025-58181, CVE-2026-39827..39835, CVE-2026-42508, CVE-2026-46595, CVE-2026-46597, CVE-2026-46598 |
| `golang.org/x/net` | v0.21.0 | **v0.57.0** | CVE-2023-45288, CVE-2025-22870, CVE-2025-22872, CVE-2026-25680 |
| `github.com/xuri/excelize/v2` | v2.8.1 | **v2.11.0** | CVE-2026-54063 (`checkSheet` OOM/panic on crafted `<row r="N">`) |

**Breaking**: `go` directive bumped `1.22 -> 1.25` (required by newer `x/crypto`). Consumers must be on Go 1.25+.

---

## [0.2.0] — 2026-08-12 — Performance, Slim Chromium, Tab Pool

### Summary

Second round of PDF-path optimization plus new options to shrink container
images. Fully backwards-compatible: existing code compiles and runs
unchanged.

Headline numbers (Apple M2, `go1.25.0`, `-benchtime=20x`):

| Bench | Before (v0.1.x) | This release | Δ |
|---|--:|--:|--:|
| `Render_Chromium` (small) | ~793 ms | **~635 ms** | −20% |
| `Render_Chromium_Medium` | ~800 ms | **~640 ms** | −20% |
| `Render_Wkhtmltopdf` (ref) | n/a | ~490 ms | reference |

Container image with Chromium: **~200 MB → ~80 MB** by pointing
`WithPDFChromePath` at `chrome-headless-shell`.

---

### New Options

#### `WithPDFChromePath(path string) Option`

Override the Chromium binary. Point at `chrome-headless-shell` (Chrome's
Blink-only rendering shell, ~80 MB) instead of full Chrome (~200 MB) —
byte-equivalent PDF output.

```go
// One-time install (e.g. Docker build stage):
//   npx @puppeteer/browsers install chrome-headless-shell@stable
gen := docgen.New(
    docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
    docgen.WithPDFChromePath("/opt/chrome-headless-shell/chrome-headless-shell"),
)
```

#### `WithPDFExtraFlags(flags ...string) Option`

Append extra Chromium command-line flags after the chromedp defaults. Each
entry is `"flag"` (bool true) or `"flag=value"`.

```go
docgen.WithPDFExtraFlags(
    "font-render-hinting=none",
    "hide-scrollbars",
)
```

---

### Performance Improvements

#### PDF Engine (`engine/pdf`)

| Change | Impact |
|--------|--------|
| Reusable tab pool (`chan *tabHandle`) replaces plain semaphore | Skips ~200–500 ms `chromedp.NewContext` (Target.create + Page.enable) per render after warmup. Pool doubles as the concurrency gate. |
| Dropped `chromedp.WaitReady("html")` after `SetDocumentContent` | `page.SetDocumentContent` installs the DOM synchronously on the browser side. WaitReady was pure CDP round-trip overhead. Saves ~5–30 ms per render. |

#### HTML Template (`template/html.go`)

| Change | Impact |
|--------|--------|
| Dropped per-render `Clone()` | `html/template.Template.Execute` is documented safe for concurrent use once parsed. Saves one full AST copy per render. |
| Pooled exec buffer via `sync.Pool` | Reduces GC pressure. |
| `LoadOrStore` on the parse cache | First-miss race can no longer double-parse. |

---

## [0.1.1] — Performance & Concurrency Optimization

### Summary

Internal performance improvements only. **No public API changes**; all existing code continues to compile and behave identically.

---

### New Options

#### `WithPDFMaxConcurrency(n int) Option`

Controls the maximum number of concurrent Chromium tab renders within a single `Generator`.

```go
// Cap at 4 tabs — suitable for memory-constrained pods (~512 MB)
gen := docgen.New(
    docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
    docgen.WithPDFMaxConcurrency(4),
)

// Allow 32 concurrent tabs — suitable for high-throughput batch processors
gen := docgen.New(
    docgen.WithPDFMaxConcurrency(32),
    docgen.WithTimeout(30 * time.Second),
)
```

Default (zero value): `runtime.GOMAXPROCS(0)` — equals the number of logical CPUs available to the process.

---

### Performance Improvements

#### PDF Engine (`engine/pdf`)

| Change | Impact |
|--------|--------|
| Persistent Chromium browser process per `Generator` | Chrome starts **once** per engine lifetime instead of once per `PDF()` call. Subsequent renders open a new tab (~10–50 ms) instead of spawning a new process (~100–500 ms). |
| Removed redundant `Navigate("about:blank")` | Saves one CDP round-trip per render. New tabs already start at `about:blank`. |
| Bounded concurrency via semaphore (`chan struct{}`) | Prevents N concurrent calls from spawning N simultaneous Chrome tabs. Caller blocks (respecting `ctx` cancellation) until a slot is free. |
| Render deadline isolated to tab context | The per-render timeout cancels the tab, not the shared browser. Browser stays alive after a timed-out render. |
| `sync.Pool` for `bytes.Buffer` in `light` path | Reduces GC pressure for high-frequency light renders. |
| Fixed silent error discard in `light()` | `gofpdf.Output()` errors are now returned instead of silently dropped. |
| `engine.Close()` for resource cleanup | Callers can type-assert the returned `Engine` to `io.Closer` for deterministic browser shutdown. |

#### CSV Engine (`engine/csv`)

| Change | Before | After | Delta |
|--------|-------:|------:|------:|
| `Build()` — time | 7 175 ns/op | 2 746 ns/op | **−62 %** |
| `Build()` — memory | 5 706 B/op | 2 225 B/op | **−61 %** |
| `Build()` — allocations | 92 allocs/op | 46 allocs/op | **−50 %** |
| `Generate()` — allocations | 3 allocs/op | 2 allocs/op | **−33 %** |

Mechanism: `text/template` is parsed once per unique template string and cached in a `sync.Map`. Each call clones the cached template (`Clone()`) and attaches a fresh function map pointing to the current call's accumulator. `bytes.Buffer` for both template execution and CSV serialisation is pooled via `sync.Pool`.

#### Excel Engine (`engine/excel`)

| Change | Before | After | Delta |
|--------|-------:|------:|------:|
| `Build()` — time | 8 369 ns/op | 3 873 ns/op | **−54 %** |
| `Build()` — memory | 6 258 B/op | 2 354 B/op | **−62 %** |
| `Build()` — allocations | 106 allocs/op | 51 allocs/op | **−52 %** |
| `Generate()` — allocations | unchanged | unchanged | excelize internals dominate |

Same mechanism as CSV. `excelize.NewFile()` internal allocations (6 856 allocs/op) dominate `Generate()` — further gains there require replacing or patching excelize.

---

### Internal Architecture Changes

- `engine/pdf/chromium.go`: extracted `renderInTab(tabCtx, html)` — renders inside a caller-supplied tab context. `RenderChromeDP` (public standalone function) is unchanged in behaviour; it now creates its own allocator → browser → tab stack explicitly.
- `engine/pdf/pdf.go`: `engine` struct gains `allocCtx`, `browserCtx`, `sem`, and `mu` fields. `ensureBrowser()` initialises or reinitialises the browser under a mutex; the fast path (browser alive) costs only a mutex acquire and a non-blocking channel select.
- `engine/csv/builder.go`, `engine/excel/builder.go`: template cache (`sync.Map`) and execution buffer pool (`sync.Pool`) added.
- `engine/csv/engine.go`, `engine/excel/engine.go`: output buffer pool (`sync.Pool`) added; output is copied before the buffer is returned to the pool.
- `options.go`: `Config.PDFMaxConcurrency` field added; `WithPDFMaxConcurrency` option added.
- `export.go`: `MaxConcurrency` forwarded from `Config` to `pdf.EngineConfig`.

---

### Compatibility

- Go version requirement: **unchanged** (`go 1.22`).
- Module path: **unchanged** (`github.com/viantonugroho11/go-docgen`).
- Public `Generator` interface: **unchanged**.
- `EngineConfig.MaxConcurrency` zero value behaves identically to previous releases (unlimited concurrency → now defaults to `GOMAXPROCS`; this is a refinement, not a breaking change).
- All existing tests pass: `go test ./... -short`.
