package docgen_test

import (
	"context"
	"fmt"

	"github.com/viantonugroho11/go-docgen"
)

func ExampleGenerator_CSV() {
	gen := docgen.New()
	tmpl := `{{row "name" "age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
		},
	}

	out, err := gen.CSV(context.Background(), tmpl, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(out) > 0)
	// Output: true
}

func ExampleGenerator_Excel() {
	gen := docgen.New()
	tmpl := `{{sheet "Users"}}{{row "name"}}{{range .Users}}{{row .Name}}{{end}}`
	data := map[string]any{
		"Users": []map[string]any{
			{"Name": "Bob"},
		},
	}

	out, err := gen.Excel(context.Background(), tmpl, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(out) > 0)
	// Output: true
}
