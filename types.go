package docgen

// CSVRow is one logical row for CSV-style data (optional helper type for callers).
type CSVRow []string

// ExcelSheet describes one worksheet when building Excel outside templates.
type ExcelSheet struct {
	Name string
	Rows [][]string
}
