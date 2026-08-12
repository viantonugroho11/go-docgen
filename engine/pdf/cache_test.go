package pdf

import (
	"bytes"
	"testing"
)

func TestCache_HitAndEvict(t *testing.T) {
	c := newCache(2)

	c.put("a", []byte{1})
	c.put("b", []byte{2})
	if got, ok := c.get("a"); !ok || !bytes.Equal(got, []byte{1}) {
		t.Fatalf(`get("a") = %v, %v; want [1], true`, got, ok)
	}

	// Access order after get: a (MRU), b (LRU).
	c.put("c", []byte{3}) // should evict b
	if _, ok := c.get("b"); ok {
		t.Fatal(`get("b") returned hit after eviction`)
	}
	if got, ok := c.get("c"); !ok || !bytes.Equal(got, []byte{3}) {
		t.Fatalf(`get("c") = %v, %v; want [3], true`, got, ok)
	}
}

func TestCache_Disabled(t *testing.T) {
	c := newCache(0)
	c.put("a", []byte{1})
	if _, ok := c.get("a"); ok {
		t.Fatal("disabled cache returned a hit")
	}
}

func TestCache_ValueDefensiveCopy(t *testing.T) {
	c := newCache(1)
	original := []byte{1, 2, 3}
	c.put("k", original)
	original[0] = 99 // mutate after put

	got, ok := c.get("k")
	if !ok {
		t.Fatal("miss")
	}
	if got[0] == 99 {
		t.Fatal("cache did not take a defensive copy of the input")
	}
}
