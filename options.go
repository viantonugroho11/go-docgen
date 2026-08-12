package docgen

import (
	"time"

	"github.com/viantonugroho11/go-docgen/engine/pdf"
)

// PDFRenderMode selects the PDF backend for Generator.PDF / PDFFromFile.
type PDFRenderMode uint8

const (
	// PDFRenderAuto tries Chromium first, then the lightweight path (same as zero value).
	PDFRenderAuto PDFRenderMode = PDFRenderMode(pdf.RenderModeAuto)
	// PDFRenderChromium uses only headless Chromium; no fallback.
	PDFRenderChromium PDFRenderMode = PDFRenderMode(pdf.RenderModeChromium)
	// PDFRenderLight uses only gofpdf text layout (not full HTML rendering).
	PDFRenderLight PDFRenderMode = PDFRenderMode(pdf.RenderModeLight)
)

type Config struct {
	Timeout           time.Duration
	PDFRenderMode     PDFRenderMode
	PDFMaxConcurrency int
	PDFChromePath     string
	PDFExtraFlags     []string
	PDFCacheSize      int
	PDFPrewarm        bool
}

type Option func(*Config)

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.Timeout = timeout
	}
}

// WithPDFRenderMode fixes the PDF pipeline at generator construction time.
func WithPDFRenderMode(mode PDFRenderMode) Option {
	return func(cfg *Config) {
		cfg.PDFRenderMode = mode
	}
}

// WithPDFMaxConcurrency sets the maximum number of concurrent Chromium tab renders.
// Defaults to runtime.GOMAXPROCS(0) when zero or negative.
// Tune higher for throughput-heavy workloads; lower to cap memory usage per pod.
// Has no effect when PDFRenderMode is PDFRenderLight.
func WithPDFMaxConcurrency(n int) Option {
	return func(cfg *Config) {
		cfg.PDFMaxConcurrency = n
	}
}

// WithPDFChromePath overrides the Chromium executable used by the PDF engine.
//
// The default (empty string) makes chromedp search PATH for the standard
// Chrome / Chromium binaries (~200 MB installed). Point this at a
// chrome-headless-shell binary (~80 MB) to shrink container images:
//
//	// Install once (per image build):
//	//   npx @puppeteer/browsers install chrome-headless-shell@stable
//	// Then:
//	gen := docgen.New(
//	    docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
//	    docgen.WithPDFChromePath("/opt/chrome-headless-shell/chrome-headless-shell"),
//	)
//
// chrome-headless-shell is the Chrome team's Blink-based rendering shell; PDF
// output is byte-equivalent to full Chrome's --print-to-pdf. Has no effect
// when PDFRenderMode is PDFRenderLight.
func WithPDFChromePath(path string) Option {
	return func(cfg *Config) {
		cfg.PDFChromePath = path
	}
}

// WithPDFExtraFlags appends extra Chromium command-line flags. Each entry is
// either "flag" (bool true) or "flag=value". Applied after chromedp defaults.
// Has no effect when PDFRenderMode is PDFRenderLight.
//
//	docgen.WithPDFExtraFlags(
//	    "font-render-hinting=none",
//	    "hide-scrollbars",
//	)
func WithPDFExtraFlags(flags ...string) Option {
	return func(cfg *Config) {
		cfg.PDFExtraFlags = append(cfg.PDFExtraFlags, flags...)
	}
}

// WithPDFCacheSize enables an in-memory LRU cache of rendered PDFs keyed by
// sha256(html). A hit returns instantly, skipping the browser round-trip and
// PrintToPDF entirely — effective latency for repeated inputs drops to ~0 ms.
// Zero (default) disables the cache. Recommended: 64-512 for invoice/statement
// workloads with high content repetition.
func WithPDFCacheSize(n int) Option {
	return func(cfg *Config) {
		cfg.PDFCacheSize = n
	}
}

// WithPDFPrewarm launches Chromium and materialises every pooled tab in a
// background goroutine at Generator construction time, so the first PDF()
// call does not pay the ~200-500 ms cold-start cost. Has no effect when
// PDFRenderMode is PDFRenderLight.
func WithPDFPrewarm(enable bool) Option {
	return func(cfg *Config) {
		cfg.PDFPrewarm = enable
	}
}

func defaultConfig() Config {
	return Config{Timeout: 10 * time.Second}
}
