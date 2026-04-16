package excel

import (
	"bytes"

	"github.com/xuri/excelize/v2"
)

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

	var buf bytes.Buffer
	err := f.Write(&buf)
	return buf.Bytes(), err
}
