package cachebox

import "testing"

func TestSimpleCacheBasic(t *testing.T) {
	c := NewSimpleCache[string, int]()
	c.Set("a", 1)

	val, ok := c.Get("a")
	if !ok || val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}
