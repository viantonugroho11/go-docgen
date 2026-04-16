package pdf

import (
	"context"
	"testing"
	"time"
)

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	e, ok := New(0).(*engine)
	if !ok {
		t.Fatal("expected concrete *engine type")
	}
	if e.cfg.timeout != 10*time.Second {
		t.Fatalf("timeout = %v, want %v", e.cfg.timeout, 10*time.Second)
	}
}

func TestRender_FallbackProducesBytes(t *testing.T) {
	e := New(200 * time.Millisecond)

	out, err := e.Render(context.Background(), "<h1>Hello</h1>")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
}
