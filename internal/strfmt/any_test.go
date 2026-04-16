package strfmt

import "testing"

func TestFormatAny(t *testing.T) {
	tests := []struct {
		in   any
		want string
	}{
		{"x", "x"},
		{42, "42"},
		{int64(9), "9"},
		{true, "true"},
		{1.5, "1.5"},
	}
	for _, tt := range tests {
		if got := FormatAny(tt.in); got != tt.want {
			t.Fatalf("FormatAny(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
