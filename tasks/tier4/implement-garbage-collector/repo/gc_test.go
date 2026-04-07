package gc

import "testing"

func TestNewGC(t *testing.T) {
	store := NewStore()
	gc := NewGarbageCollector(store)
	if gc == nil {
		t.Fatal("gc should not be nil")
	}
}
