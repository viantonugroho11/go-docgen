package template

import "testing"

func TestHTMLEngineRender(t *testing.T) {
	engine := NewHTML()

	got, err := engine.Render("<h1>Hello {{.Name}}</h1>", map[string]string{"Name": "Alice"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "<h1>Hello Alice</h1>"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestTextEngineRender(t *testing.T) {
	engine := NewText()

	got, err := engine.Render("Hello {{.Name}}", map[string]string{"Name": "Bob"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	want := "Hello Bob"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestHTMLEngineRender_InvalidTemplate(t *testing.T) {
	engine := NewHTML()
	if _, err := engine.Render("{{.Name", map[string]string{"Name": "Alice"}); err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
