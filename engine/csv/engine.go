package csv

import (
	"bytes"
	"encoding/csv"
	"sync"
)

var csvGenBuf = sync.Pool{New: func() any { return new(bytes.Buffer) }}

type Engine interface {
	Generate(rows [][]string) ([]byte, error)
}

type engine struct{}

func New() Engine { return &engine{} }

func (e *engine) Generate(rows [][]string) ([]byte, error) {
	buf := csvGenBuf.Get().(*bytes.Buffer)
	buf.Reset()

	w := csv.NewWriter(buf)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		csvGenBuf.Put(buf)
		return nil, err
	}

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	csvGenBuf.Put(buf)
	return out, nil
}
