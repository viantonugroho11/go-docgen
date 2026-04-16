package pdf

import (
	"context"
	"testing"
	"time"
)

func TestNew_UsesDefaultTimeoutWhenZero(t *testing.T) {
	e, ok := New(EngineConfig{}).(*engine)
	if !ok {
		t.Fatal("expected concrete *engine type")
	}
	if e.cfg.timeout != 10*time.Second {
		t.Fatalf("timeout = %v, want %v", e.cfg.timeout, 10*time.Second)
	}
	if e.cfg.mode != RenderModeAuto {
		t.Fatalf("mode = %v, want RenderModeAuto", e.cfg.mode)
	}
}

func TestRender_FallbackProducesBytes(t *testing.T) {
	e := New(EngineConfig{Timeout: 200 * time.Millisecond, Mode: RenderModeAuto})

	out, err := e.Render(context.Background(), "<h1>Hello</h1>")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
}

func TestRender_ModeLightSkipsChromium(t *testing.T) {
	e := New(EngineConfig{Timeout: time.Nanosecond, Mode: RenderModeLight})

	out, err := e.Render(context.Background(), "<h1>Hello</h1>")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if len(out) == 0 {
		t.Fatal("Render() returned empty bytes")
	}
}
