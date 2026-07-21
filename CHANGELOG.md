# Changelog

All notable changes to go-docgen are documented here.

---

## [Unreleased] — Performance & Concurrency Optimization

### Summary

This release focuses exclusively on internal performance improvements. **No public API changes**; all existing code continues to compile and behave identically.

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
