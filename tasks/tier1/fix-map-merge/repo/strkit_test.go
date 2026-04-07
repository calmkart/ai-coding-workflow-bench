package strkit

import "testing"

func TestMergeMapsBasic(t *testing.T) {
	m1 := map[string]string{"a": "1"}
	m2 := map[string]string{"b": "2"}
	got := MergeMaps(m1, m2)
	if got["a"] != "1" || got["b"] != "2" {
		t.Fatalf("MergeMaps = %v, want {a:1, b:2}", got)
	}
}

func TestMergeMapsOverride(t *testing.T) {
	m1 := map[string]string{"a": "1"}
	m2 := map[string]string{"a": "2"}
	got := MergeMaps(m1, m2)
	if got["a"] != "2" {
		t.Fatalf("expected a=2, got %s", got["a"])
	}
}

func TestKeys(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}
	keys := Keys(m)
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}
