package pdf

// RenderMode selects how HTML is turned into PDF bytes.
type RenderMode uint8

const (
	// RenderModeAuto tries headless Chromium first, then the lightweight text-based path if Chromium fails.
	RenderModeAuto RenderMode = iota
	// RenderModeChromium uses only headless Chromium (chromedp). Errors are returned if printing fails.
	RenderModeChromium
	// RenderModeLight uses only gofpdf MultiCell on the HTML string (not a real HTML/CSS layout engine).
	RenderModeLight
)
