package pdf

import (
	"bytes"
	"context"
	"sync"

	"github.com/jung-kurt/gofpdf"
)

var lightBufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

func light(ctx context.Context, html string) ([]byte, error) {
	_ = ctx

	f := gofpdf.New("P", "mm", "A4", "")
	f.AddPage()
	f.SetFont("Arial", "", 12)
	f.MultiCell(0, 10, html, "", "", false)

	buf := lightBufPool.Get().(*bytes.Buffer)
	buf.Reset()

	if err := f.Output(buf); err != nil {
		lightBufPool.Put(buf)
		return nil, err
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	lightBufPool.Put(buf)
	return out, nil
}
