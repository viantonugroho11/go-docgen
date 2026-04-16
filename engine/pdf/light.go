package pdf

import (
	"bytes"
	"context"

	"github.com/jung-kurt/gofpdf"
)

func light(ctx context.Context, html string) ([]byte, error) {
	_ = ctx

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 10, html, "", "", false)

	var buf bytes.Buffer
	_ = pdf.Output(&buf)
	return buf.Bytes(), nil
}
