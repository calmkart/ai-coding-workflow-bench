package cachebox

import "testing"

func TestSimpleCache(t *testing.T) {
	c := NewSimpleCache[string, int]()
	c.Set("a", 1)
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}
}
