package excel

import (
	"bytes"
	"sync"

	"github.com/xuri/excelize/v2"
)

var excelGenBuf = sync.Pool{New: func() any { return new(bytes.Buffer) }}

type Engine interface {
	Generate(sheets []Sheet) ([]byte, error)
}

type engine struct{}

func New() Engine { return &engine{} }

func (e *engine) Generate(sheets []Sheet) ([]byte, error) {
	f := excelize.NewFile()

	for i, s := range sheets {
		name := s.Name
		if i == 0 {
			f.SetSheetName("Sheet1", name)
		} else {
			f.NewSheet(name)
		}

		for r, row := range s.Rows {
			startCell, err := excelize.CoordinatesToCellName(1, r+1)
			if err != nil {
				return nil, err
			}
			if err := f.SetSheetRow(name, startCell, &row); err != nil {
				return nil, err
			}
		}
	}

	buf := excelGenBuf.Get().(*bytes.Buffer)
	buf.Reset()
	err := f.Write(buf)
	if err != nil {
		excelGenBuf.Put(buf)
		return nil, err
	}
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	excelGenBuf.Put(buf)
	return out, nil
}
