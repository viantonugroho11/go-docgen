package pdf

import (
	"context"
	"strings"
	"testing"
	"time"
)

func benchHTMLSmall() string {
	return "<html><body><h1>Bench</h1><p>x</p></body></html>"
}

func benchHTMLMedium() string {
	var b strings.Builder
	b.Grow(8000)
	b.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>Bench</title></head><body><table>")
	for i := 0; i < 80; i++ {
		b.WriteString("<tr><td>")
		b.WriteString(strings.Repeat("a", 20))
		b.WriteString("</td><td>")
		b.WriteString(strings.Repeat("b", 20))
		b.WriteString("</td></tr>")
	}
	b.WriteString("</table></body></html>")
	return b.String()
}

// BenchmarkRender_Light measures the gofpdf-only path (no Chromium).
func BenchmarkRender_Light(b *testing.B) {
	html := benchHTMLMedium()
	e := New(EngineConfig{Mode: RenderModeLight})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Render(ctx, html); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRender_Light_Small is a smaller HTML payload for a tighter loop.
func BenchmarkRender_Light_Small(b *testing.B) {
	html := benchHTMLSmall()
	e := New(EngineConfig{Mode: RenderModeLight})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Render(ctx, html); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRender_Chromium measures chromedp + PrintToPDF (slow; needs Chromium).
// Skipped under go test -short to keep CI and quick runs fast.
func BenchmarkRender_Chromium(b *testing.B) {
	if testing.Short() {
		b.Skip("omit chromedp in -short; run: go test -bench=BenchmarkRender_Chromium -benchmem -run=^$ ./engine/pdf/")
	}
	html := benchHTMLSmall()
	e := New(EngineConfig{
		Mode:    RenderModeChromium,
		Timeout: 3 * time.Minute,
	})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Render(ctx, html); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderChromeDP is the raw chromedp path (same engine as RenderModeChromium, no engine wrapper).
func BenchmarkRenderChromeDP(b *testing.B) {
	if testing.Short() {
		b.Skip("omit chromedp in -short; run: go test -bench=BenchmarkRenderChromeDP -benchmem -run=^$ ./engine/pdf/")
	}
	html := benchHTMLSmall()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := RenderChromeDP(ctx, html); err != nil {
			b.Fatal(err)
		}
	}
}
