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
	Timeout         time.Duration
	PDFRenderMode   PDFRenderMode
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

func defaultConfig() Config {
	return Config{Timeout: 10 * time.Second}
}
