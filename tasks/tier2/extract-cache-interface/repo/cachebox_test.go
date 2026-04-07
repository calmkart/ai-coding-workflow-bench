package cachebox

import "testing"

func TestMapCacheBasic(t *testing.T) {
	c := NewMapCache[string, int]()
	c.Set("a", 1)

	val, ok := c.Get("a")
	if !ok || val != 1 {
		t.Fatalf("expected 1, got %d (ok=%v)", val, ok)
	}

	c.Delete("a")
	_, ok = c.Get("a")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestMapCacheLen(t *testing.T) {
	c := NewMapCache[string, string]()
	c.Set("a", "1")
	c.Set("b", "2")

	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}
