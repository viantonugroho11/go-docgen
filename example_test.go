package godocgen_test

import (
	"context"
	"fmt"

	godocgen "github.com/viantonugroho11/go-docgen"
)

func ExampleExporter_ToCSVTemplate() {
	exp := godocgen.New()
	tmpl := `{{row "name" "age"}}{{range .People}}{{row .Name .Age}}{{end}}`
	data := map[string]any{
		"People": []map[string]any{
			{"Name": "Alice", "Age": 30},
		},
	}

	out, err := exp.ToCSVTemplate(context.Background(), tmpl, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(out) > 0)
	// Output: true
}

func ExampleExporter_ToExcelTemplate() {
	exp := godocgen.New()
	tmpl := `{{sheet "Users"}}{{row "name"}}{{range .Users}}{{row .Name}}{{end}}`
	data := map[string]any{
		"Users": []map[string]any{
			{"Name": "Bob"},
		},
	}

	out, err := exp.ToExcelTemplate(context.Background(), tmpl, data)
	if err != nil {
		panic(err)
	}

	fmt.Println(len(out) > 0)
	// Output: true
}
