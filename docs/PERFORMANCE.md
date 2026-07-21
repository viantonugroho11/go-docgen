# go-docgen — Performance Reference

This document provides detailed performance measurements, architecture notes, and a comparison of HTML-to-PDF tools that are commonly evaluated alongside go-docgen.

---

## Table of Contents

1. [Benchmark Environment](#benchmark-environment)
2. [CSV & Excel — Before / After Optimization](#csv--excel--before--after-optimization)
3. [PDF Engine Comparison](#pdf-engine-comparison)
4. [HTML-to-PDF Tool Landscape](#html-to-pdf-tool-landscape)
5. [Concurrency & Browser Lifecycle](#concurrency--browser-lifecycle)
6. [Memory Analysis](#memory-analysis)
7. [Choosing a PDF Backend](#choosing-a-pdf-backend)
8. [Reproducing the Numbers](#reproducing-the-numbers)

---

## Benchmark Environment

All figures in this document were measured on:

| Field | Value |
|-------|-------|
| OS | `darwin` / `arm64` |
| CPU | Apple M2 |
| Go | `go1.25.0` |
| Date | 2026-07-21 |

Numbers are **machine-specific**. Re-run the benchmarks on your target environment before making deployment decisions. Linux/x86-64 servers typically yield different absolute numbers (Chromium startup is slower on Linux containers without GPU acceleration), but the **ratios** between backends are generally consistent.

---

## CSV & Excel — Before / After Optimization

### What changed

Both `csv.Build()` and `excel.Build()` previously called `text/template.Parse()` on every invocation. Go's template parser allocates heavily (AST nodes, string copies, intern tables). The fix caches parsed templates by source string in a `sync.Map` and clones the cached tree per call.

Additionally, `bytes.Buffer` instances used for template execution and serialization output are now pooled via `sync.Pool`, reducing GC pressure under concurrent load.

### CSV results

```
go test -bench=. -benchmem -run=^$ ./engine/csv/
```

| Benchmark | Before | After | Δ time | Δ memory | Δ allocs |
|-----------|-------:|------:|-------:|---------:|---------:|
| `BenchmarkBuild` | 7 175 ns/op | 2 746 ns/op | **−62 %** | **−61 %** | **−50 %** |
| `BenchmarkGenerate` | 4 074 ns/op | 3 067 ns/op | −25 % | −1 % | **−33 %** |

`Build` is the hot path in every `CSV()` call. A −62 % reduction means roughly 2.6× more `CSV()` calls per second at the same CPU budget, with half the GC pressure.

### Excel results

```
go test -bench=. -benchmem -run=^$ ./engine/excel/
```

| Benchmark | Before | After | Δ time | Δ memory | Δ allocs |
|-----------|-------:|------:|-------:|---------:|---------:|
| `BenchmarkBuild` | 8 369 ns/op | 3 873 ns/op | **−54 %** | **−62 %** | **−52 %** |
| `BenchmarkGenerate` | 1 312 723 ns/op | 1 419 550 ns/op | ~same | ~same | ~same |

`Generate` is dominated by `excelize` internal allocations (6 856 allocs/op) — the pooled output buffer is a rounding error there. The excelize library itself calls `SetSheetRow` and manages XML encoding internally; further gains require replacing or patching that dependency.

---

## PDF Engine Comparison

### go-docgen internal PDF backends

```
go test -bench=BenchmarkRender_Light -benchmem -run=^$ ./engine/pdf/
```

| Mode | ns/op | B/op | allocs/op | Notes |
|------|------:|-----:|----------:|-------|
| `PDFRenderLight` (medium HTML) | 477 286 | 3 732 543 | 713 | gofpdf `MultiCell` — no HTML layout |
| `PDFRenderLight` (small HTML) | 163 148 | 1 234 848 | 198 | Same |
| `PDFRenderChromium` | ~750–800 ms | ~1.9 MB | ~2 400 | machine-dependent; Chromium cold start |

`PDFRenderLight` allocations are dominated by gofpdf's font metrics and PDF serialization internals, not by go-docgen. The `sync.Pool` on the output buffer reduces marginal pressure but does not affect the gofpdf internal cost.

### `pdfcompare` tool — wall-clock comparison

> Captured 2026-04-16, Apple M2, `go1.23.4`, `runs=3 warmup=1`, same ~80-row HTML table fixture.

| Backend | Median | Mean | p95 | ns/op | B/op (Go heap) | allocs/op (Go heap) |
|---------|-------:|-----:|----:|------:|---------------:|--------------------:|
| `chromedp` (`engine/pdf.RenderChromeDP`) | 776 ms | 776 ms | 796 ms | 775 702 638 | 1 916 728 | 2 341 |
| go-docgen `PDFRenderAuto` | 766 ms | 761 ms | 768 ms | 760 544 444 | 1 955 781 | 2 401 |
| go-docgen `PDFRenderChromium` | 798 ms | 793 ms | 820 ms | 793 411 819 | 1 953 800 | 2 407 |
| go-docgen `PDFRenderLight` | 1 ms | 1 ms | 1 ms | 848 375 | 4 983 501 | 974 |
| Chrome CLI `--print-to-pdf` | 2.055 s | 2.111 s | 2.237 s | 2 111 353 583 | 16 882 | 55 |
| `wkhtmltopdf` CLI (subprocess) | 403 ms | 407 ms | 419 ms | 407 107 486 | 21 245 | 109 |
| `gofpdf` `MultiCell` only | 1 ms | 1 ms | 1 ms | 983 749 | 4 952 437 | 930 |

> **B/op and allocs/op for subprocess tools** (Chrome CLI, wkhtmltopdf) count only Go heap activity around the `exec.Command` call — not the native process memory. Chromium and wkhtmltopdf use 80–300 MB of native RSS that does not appear here.

Run this yourself:

```bash
go run -C cmd/pdfcompare . -runs 10 -warmup 2
```

---

## HTML-to-PDF Tool Landscape

### Tool comparison matrix

| Tool | HTML/CSS fidelity | JS support | Maintenance | Go integration | Typical latency (cold) | Typical latency (warm/reuse) |
|------|------------------|-----------|-------------|---------------|----------------------:|-----------------------------:|
| **go-docgen `PDFRenderChromium`** | Full (Blink) | Yes | Active (Chromium) | Native — no subprocess | ~500–800 ms | ~50–150 ms (browser reused) |
| **go-docgen `PDFRenderLight`** | None — text dump | No | Active | Native | < 1 ms | < 1 ms |
| **wkhtmltopdf** | Partial (WebKit-based, frozen 2023) | Limited | **Abandoned** (no releases since 0.12.6, 2020) | subprocess only | ~200–500 ms | subprocess per call |
| **WeasyPrint** | Good (CSS Paged Media) | No | Active (Python) | subprocess only | ~300–800 ms | subprocess per call |
| **Chrome CLI** (`--headless --print-to-pdf`) | Full (Blink) | Yes | Active | subprocess only | ~1.5–3 s | subprocess per call (new process each time) |
| **puppeteer / playwright** | Full (Blink) | Yes | Active | subprocess / gRPC | ~200–600 ms | browser reuse possible |
| **gotenberg** | Full (Chromium) | Yes | Active | HTTP API | network + ~200–500 ms | server keeps browser alive |
| **rod** (Go CDP library) | Full (Blink) | Yes | Active | Native Go | same as chromedp | browser reuse possible |

### Key trade-offs

#### wkhtmltopdf

**Do not use for new projects.**

- Based on Qt WebKit, which was forked from WebKit in 2010 and has not received security or feature updates since 2020.
- No support for CSS Grid, CSS Flexbox, CSS custom properties, modern `@media` queries, or ES6+ JavaScript.
- Fast because it is lightweight, not because it is efficient — it simply implements less.
- Ships as a large static binary (~80 MB) with Qt dependencies that conflict with container hardening (e.g. Alpine, `--no-install-recommends`).
- go-docgen does not wrap wkhtmltopdf; the `pdfcompare` tool measures it as an external reference point only.

#### WeasyPrint

- Implements CSS Paged Media (`@page`, margin boxes, named pages) better than Chromium.
- No JavaScript support — pure HTML/CSS document rendering.
- Python subprocess from Go means process-spawn cost on every call unless you run it as a long-lived service.
- Excellent choice for publishing-style PDFs (books, invoices with strict page layout).
- go-docgen does not wrap WeasyPrint.

#### Chrome CLI (`--headless --print-to-pdf`)

- Spawns a full Chrome process per invocation, including writing a user-data-dir.
- Slowest option (~2 s+ on most systems) because of process isolation overhead.
- No connection reuse — every call is a cold start.
- Use only for one-off conversions, not in a service.

#### go-docgen `PDFRenderChromium` (chromedp)

- Runs Chromium via Chrome DevTools Protocol (CDP) over a WebSocket.
- go-docgen keeps **one** browser process alive for the `Generator`'s lifetime; subsequent renders open tabs, not new processes.
- Full HTML/CSS/JS support (whatever the installed Chromium version supports).
- Requires Chromium (or Google Chrome) to be installed on the host. In containers, use `chromium-browser` or the official `ghcr.io/puppeteer/puppeteer` base image.
- First render in a process pays ~100–500 ms browser startup; subsequent renders pay ~10–50 ms tab open + actual render time.

#### go-docgen `PDFRenderLight`

- Calls `gofpdf.MultiCell` on the raw HTML string — there is **no HTML or CSS parsing**.
- HTML tags appear as literal text in the PDF.
- Use only when: (a) the template produces plain text or very simple content, and (b) you need sub-millisecond generation without Chromium as a dependency.
- Not a substitute for real HTML rendering.

---

## Concurrency & Browser Lifecycle

### Architecture (after optimization)

```
┌─────────────────────────────────────────────────────┐
│  docgen.Generator  (one per service / pool entry)   │
│                                                     │
│  ┌───────────────────────────────────────────────┐  │
│  │  engine/pdf.engine                            │  │
│  │                                               │  │
│  │  allocCtx ──► Chromium process (1 per engine) │  │
│  │  browserCtx                                   │  │
│  │  sem (chan struct{}, cap = MaxConcurrency)     │  │
│  │                                               │  │
│  │  Render() call 1 ──► tab 1 (chromedp context) │  │
│  │  Render() call 2 ──► tab 2                    │  │
│  │  Render() call N ──► blocks at semaphore       │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Lifecycle states

| State | Description |
|-------|-------------|
| **Uninitialised** | `New()` called, no browser spawned yet. First `PDF()` call triggers `ensureBrowser()`. |
| **Running** | Browser process alive. All `PDF()` calls open tabs against the shared browser. |
| **Crashed** | Browser context's `Done()` channel is closed. Next `PDF()` call detects this, tears down the dead context, and reinitialises. |
| **Closed** | `Close()` called explicitly. Browser and allocator contexts are cancelled. Next `PDF()` call starts a new browser. |

### Semaphore behaviour under load

```
MaxConcurrency = 4

t=0ms   req1 ──► acquire sem (slots: 3 free)
t=0ms   req2 ──► acquire sem (slots: 2 free)
t=0ms   req3 ──► acquire sem (slots: 1 free)
t=0ms   req4 ──► acquire sem (slots: 0 free)
t=0ms   req5 ──► BLOCKS at sem (waiting)
t=0ms   req6 ──► BLOCKS at sem (waiting)

t=150ms req1 completes ──► release sem
t=150ms req5 ──► acquire sem (unblocked)

If req5's ctx deadline fires while waiting:
        req5 ──► returns ctx.Err() immediately (no Chrome tab opened)
```

### Recommended `MaxConcurrency` settings

| Deployment | Recommended value | Rationale |
|------------|------------------|-----------|
| Single-core container (256–512 MB) | 1–2 | Avoid OOM from parallel Chrome tabs |
| Standard pod (2 CPU, 1 GB) | 2–4 | Balance throughput and memory |
| Batch processor (8+ CPU, 4+ GB) | 8–16 | Maximize throughput; Chrome tabs ~50–80 MB each |
| Default (zero) | `runtime.GOMAXPROCS(0)` | Equals logical CPUs; safe starting point |

---

## Memory Analysis

### Go heap (per render call)

These are `runtime.MemStats` deltas, not total process RSS.

| Path | B/op | allocs/op | Notes |
|------|-----:|----------:|-------|
| `csv.Build()` (after) | 2 225 | 46 | Template clone + row accumulation |
| `excel.Build()` (after) | 2 354 | 51 | Template clone + sheet accumulation |
| `csv.Generate()` | 4 993 | 2 | CSV encoding; pooled buffer |
| `excel.Generate()` | ~769 KB | 6 856 | excelize XML serialisation dominates |
| `light()` PDF | ~3.7 MB | 713 | gofpdf font metrics + PDF serialisation |
| Chromium path (Go heap only) | ~1.9 MB | ~2 400 | CDP message encoding/decoding |

### Native process memory (RSS)

Go heap numbers do not include native memory inside the Chromium process.

| Component | Approximate RSS |
|-----------|----------------|
| Chromium browser process (idle) | 80–150 MB |
| Each active Chrome tab | 30–80 MB (HTML-dependent) |
| wkhtmltopdf process (per invocation) | 60–120 MB |
| WeasyPrint process (per invocation) | 50–100 MB |

For `PDFMaxConcurrency = N`, budget approximately:

```
Chrome RSS ≈ 100 MB (base) + (N × 60 MB per active tab)
```

Example: `MaxConcurrency = 4` → ~340 MB Chrome RSS peak. Pod memory limit should be at least `340 MB + Go heap + OS overhead`.

---

## Choosing a PDF Backend

```
Need full CSS/JS rendering?
  └─ Yes → PDFRenderChromium (or PDFRenderAuto)
       ├─ Chromium available on host? Yes → renders correctly
       └─ Chromium not available?    → falls back to light (PDFRenderAuto)
                                       or returns error (PDFRenderChromium)

  └─ No, only plain text needed?
       └─ PDFRenderLight
            ├─ Sub-millisecond, zero browser dependency
            └─ WARNING: HTML tags appear as literal text in output
```

### `PDFRenderAuto` vs `PDFRenderChromium`

Use `PDFRenderAuto` (default) when:
- Chromium may or may not be installed (mixed environments).
- A degraded text-only PDF is acceptable when Chromium is unavailable.

Use `PDFRenderChromium` when:
- Correct layout is mandatory — you want an error, not a broken fallback.
- You control the deployment and guarantee Chromium is present.

---

## Reproducing the Numbers

### CSV / Excel benchmarks

```bash
# Run all benchmarks (no Chromium required)
go test -bench=. -benchmem -run=^$ ./engine/csv/ ./engine/excel/

# Count=5 for stability
go test -bench=. -benchmem -run=^$ -count=5 ./engine/csv/ ./engine/excel/
```

### PDF light path benchmarks

```bash
go test -bench=BenchmarkRender_Light -benchmem -run=^$ ./engine/pdf/
```

### PDF Chromium benchmarks (requires installed Chromium)

```bash
# Engine-level
go test -bench='BenchmarkRender_Chromium|BenchmarkRenderChromeDP' \
        -benchmem -run=^$ -benchtime=3x ./engine/pdf/

# Generator-level (includes template rendering overhead)
go test -bench=BenchmarkGenerator_PDF_Chromium \
        -benchmem -run=^$ -benchtime=3x .
```

### Full suite (skip slow Chromium benches)

```bash
go test -short -bench=. -benchmem -run=^$ ./...
```

### Wall-clock comparison with external tools

```bash
# Requires: Chromium on PATH, optionally wkhtmltopdf on PATH
go run -C cmd/pdfcompare . -runs 10 -warmup 3

# Custom HTML fixture
go run -C cmd/pdfcompare . -runs 10 -warmup 3 -html /path/to/your.html
```

`pdfcompare` is a separate Go module under `cmd/pdfcompare/` so its dependencies (wkhtmltopdf subprocess wrappers, etc.) do not affect the main module's `go.sum`.
