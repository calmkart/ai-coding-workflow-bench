package strkit

import (
	"strings"
	"testing"
)

func TestDedupWithDuplicates(t *testing.T) {
	got := Dedup([]string{"a", "b", "a", "c", "b"})
	if len(got) != 3 {
		t.Fatalf("expected 3 items, got %d: %v", len(got), got)
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]string{"hello", "world", "hi"}, func(s string) bool {
		return strings.HasPrefix(s, "h")
	})
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
}

func TestMap(t *testing.T) {
	got := Map([]string{"a", "b"}, strings.ToUpper)
	if got[0] != "A" || got[1] != "B" {
		t.Fatalf("expected [A B], got %v", got)
	}
}
