package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.tmpl")
	want := "hello template"

	if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %q, want %q", got, want)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	if _, err := Load("/path/that/does/not/exist"); err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
