// Command pdfcompare measures wall-clock time for several HTML→PDF backends
// on the same fixture (go-docgen, Chrome headless CLI, wkhtmltopdf CLI, optional libwkhtmltox in-process, gofpdf).
//
// Usage (from this directory): go run . [-runs N] [-warmup N] [-html path] [-nomem]
// wkhtmltopdf is invoked as a subprocess when the binary is on PATH (or WKHTMLTOPDF_PATH).
// In-process libwkhtmltox (adrg/go-wkhtmltopdf): CGO_ENABLED=1 go run -tags libwkhtmltox . (same flags).
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	docgen "github.com/viantonugroho11/go-docgen"
	"github.com/viantonugroho11/go-docgen/engine/pdf"
	"github.com/jung-kurt/gofpdf"
)

func main() {
	runs := flag.Int("runs", 10, "timed iterations per backend (after warmup)")
	warmup := flag.Int("warmup", 2, "warmup iterations (discarded)")
	htmlPath := flag.String("html", "", "path to HTML file (default: embedded report fixture)")
	noMem := flag.Bool("nomem", false, "skip second pass (saves time); no B/op or allocs/op columns")
	flag.Parse()

	defer wkHTMLTeardown()

	html, err := loadHTML(*htmlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "html: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "pdfcompare-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	htmlFile := filepath.Join(tmpDir, "fixture.html")
	if err := os.WriteFile(htmlFile, []byte(html), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write fixture: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("PDF compare — same HTML input, wall time + Go runtime allocation shape")
	fmt.Printf("runs=%d warmup=%d nomem=%v go=%s os=%s/%s\n", *runs, *warmup, *noMem, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Println("Line 1: median/mean/p95 wall time per iteration.")
	fmt.Println("Line 2: ns/op = mean wall nanoseconds; B/op & allocs/op = (TotalAlloc, Mallocs delta) / runs over a fresh batch (Go heap only — not Chrome/wkhtml native heaps).")
	fmt.Println("For library micro-benchmarks (CSV/Excel) see README \"Latest benchmark result\".")
	fmt.Println("If chromedp SKIP but go-docgen Auto is very fast, Auto used gofpdf fallback, not Chromium.")
	fmt.Println()

	rows := []struct {
		name string
		fn   func(context.Context) error
	}{
		{
			name: "chromedp only (engine/pdf.RenderChromeDP, no fallback)",
			fn: func(ctx context.Context) error {
				_, err := pdf.RenderChromeDP(ctx, html)
				return err
			},
		},
		{
			name: "go-docgen Generator.PDF (PDFRenderAuto)",
			fn: func(ctx context.Context) error {
				gen := docgen.New(
					docgen.WithTimeout(3*time.Minute),
					docgen.WithPDFRenderMode(docgen.PDFRenderAuto),
				)
				_, err := gen.PDF(ctx, html, nil)
				return err
			},
		},
		{
			name: "go-docgen Generator.PDF (PDFRenderChromium)",
			fn: func(ctx context.Context) error {
				gen := docgen.New(
					docgen.WithTimeout(3*time.Minute),
					docgen.WithPDFRenderMode(docgen.PDFRenderChromium),
				)
				_, err := gen.PDF(ctx, html, nil)
				return err
			},
		},
		{
			name: "go-docgen Generator.PDF (PDFRenderLight)",
			fn: func(ctx context.Context) error {
				gen := docgen.New(
					docgen.WithTimeout(3*time.Minute),
					docgen.WithPDFRenderMode(docgen.PDFRenderLight),
				)
				_, err := gen.PDF(ctx, html, nil)
				return err
			},
		},
		{
			name: "Chrome/Chromium CLI (--headless --print-to-pdf)",
			fn: func(ctx context.Context) error {
				chrome := findChrome()
				if chrome == "" {
					return fmt.Errorf("chrome binary not found (set CHROME_PATH or install Chrome/Chromium)")
				}
				out := filepath.Join(tmpDir, fmt.Sprintf("chrome-%d.pdf", time.Now().UnixNano()))
				return runChromePrintPDF(ctx, chrome, htmlFile, out)
			},
		},
		{
			name: "wkhtmltopdf CLI (subprocess)",
			fn: func(ctx context.Context) error {
				wk := findWkHTMLTopdf()
				if wk == "" {
					return fmt.Errorf("wkhtmltopdf not found (install package or set WKHTMLTOPDF_PATH)")
				}
				out := filepath.Join(tmpDir, fmt.Sprintf("wkhtml-cli-%d.pdf", time.Now().UnixNano()))
				return runWkHTMLTopdfCLI(ctx, wk, htmlFile, out)
			},
		},
		{
			name: "libwkhtmltox (github.com/adrg/go-wkhtmltopdf, CGO, in-process)",
			fn: func(ctx context.Context) error {
				_ = ctx
				return runWkHTMLViaLib(htmlFile)
			},
		},
		{
			name: "gofpdf MultiCell (same idea as engine/pdf light fallback)",
			fn: func(ctx context.Context) error {
				_ = ctx
				pdf := gofpdf.New("P", "mm", "A4", "")
				pdf.AddPage()
				pdf.SetFont("Arial", "", 12)
				pdf.MultiCell(0, 10, html, "", "", false)
				var buf bytes.Buffer
				return pdf.Output(&buf)
			},
		},
	}

	for _, row := range rows {
		if err := benchOne(ctx, row.name, *warmup, *runs, !*noMem, row.fn); err != nil {
			fmt.Printf("%-58s  SKIP: %v\n", row.name, err)
			continue
		}
	}
}

func loadHTML(path string) (string, error) {
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return fixtureHTML(), nil
}

func fixtureHTML() string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>Bench</title>
<style>
body{font-family:system-ui,sans-serif;margin:24px}
table{border-collapse:collapse;width:100%}
td,th{border:1px solid #ccc;padding:6px;text-align:left}
th{background:#f0f0f0}
</style></head><body>
<h1>Laporan contoh</h1>
<p>Fixture statis untuk perbandingan performa HTML → PDF.</p>
<table><thead><tr><th>#</th><th>SKU</th><th>Qty</th><th>Total</th></tr></thead><tbody>`)
	for i := 0; i < 80; i++ {
		fmt.Fprintf(&b, "<tr><td>%d</td><td>ITEM-%04d</td><td>%d</td><td>Rp %d</td></tr>\n",
			i+1, i, (i%12)+1, 15000*(i+1))
	}
	b.WriteString(`</tbody></table></body></html>`)
	return b.String()
}

func benchOne(ctx context.Context, name string, warmup, runs int, withMem bool, fn func(context.Context) error) error {
	for i := 0; i < warmup; i++ {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	times := make([]time.Duration, 0, runs)
	for i := 0; i < runs; i++ {
		t0 := time.Now()
		if err := fn(ctx); err != nil {
			return err
		}
		times = append(times, time.Since(t0))
	}
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	med := times[len(times)/2]
	p95 := times[(len(times)*95)/100]
	if p95 < med {
		p95 = times[len(times)-1]
	}
	var sum time.Duration
	for _, d := range times {
		sum += d
	}
	mean := sum / time.Duration(len(times))
	fmt.Printf("%-58s  median=%10s  mean=%10s  p95=%10s\n", name, med.Round(time.Millisecond), mean.Round(time.Millisecond), p95.Round(time.Millisecond))
	if !withMem {
		return nil
	}
	bPerOp, allocsPerOp, err := goAllocPerOp(ctx, runs, fn)
	if err != nil {
		fmt.Printf("%-58s  mem batch SKIP: %v\n", "", err)
		return nil
	}
	fmt.Printf("%-58s  ns/op=%d  B/op=%d  allocs/op=%d\n", "", mean.Nanoseconds(), bPerOp, allocsPerOp)
	return nil
}

// goAllocPerOp runs fn runs times after GC and returns average TotalAlloc and Mallocs deltas (Go runtime only).
func goAllocPerOp(ctx context.Context, runs int, fn func(context.Context) error) (bPerOp, allocsPerOp uint64, err error) {
	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	for i := 0; i < runs; i++ {
		if err := fn(ctx); err != nil {
			return 0, 0, err
		}
	}
	runtime.ReadMemStats(&m1)
	dAlloc := m1.TotalAlloc - m0.TotalAlloc
	dMallocs := m1.Mallocs - m0.Mallocs
	if runs <= 0 {
		return 0, 0, fmt.Errorf("runs must be > 0")
	}
	ur := uint64(runs)
	return dAlloc / ur, dMallocs / ur, nil
}

func findWkHTMLTopdf() string {
	if p := os.Getenv("WKHTMLTOPDF_PATH"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	if p, err := exec.LookPath("wkhtmltopdf"); err == nil {
		return p
	}
	return ""
}

func runWkHTMLTopdfCLI(ctx context.Context, wkhtml, htmlPath, outPDF string) error {
	cmd := exec.CommandContext(ctx, wkhtml, "--quiet", htmlPath, outPDF)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func findChrome() string {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	candidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/google-chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	for _, name := range []string{"google-chrome-stable", "google-chrome", "chromium", "chromium-browser"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func runChromePrintPDF(ctx context.Context, chrome, htmlPath, outPDF string) error {
	fileURL := fileURLFromPath(htmlPath)
	cmd := exec.CommandContext(ctx, chrome,
		"--headless=new",
		"--disable-gpu",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-extensions",
		"--print-to-pdf="+outPDF,
		fileURL,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func fileURLFromPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = p
	}
	abs = filepath.ToSlash(abs)
	if runtime.GOOS == "windows" {
		return "file:///" + abs
	}
	return "file://" + abs
}
