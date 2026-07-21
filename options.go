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
	Timeout            time.Duration
	PDFRenderMode      PDFRenderMode
	PDFMaxConcurrency  int
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

func defaultConfig() Config {
	return Config{Timeout: 10 * time.Second}
}
