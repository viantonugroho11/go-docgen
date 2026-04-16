package csv

import (
	"bytes"
	"encoding/csv"
)

type Engine interface {
	Generate(rows [][]string) ([]byte, error)
}

type engine struct{}

func New() Engine { return &engine{} }

func (e *engine) Generate(rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, r := range rows {
		_ = w.Write(r)
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}
