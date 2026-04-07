package strkit

import "testing"

func TestContainsAnyMatch(t *testing.T) {
	if !ContainsAny("hello world", []string{"world", "foo"}) {
		t.Fatal("expected true for matching candidate")
	}
}

func TestContainsAnyNoMatch(t *testing.T) {
	if ContainsAny("hello world", []string{"foo", "bar"}) {
		t.Fatal("expected false for non-matching candidates")
	}
}

func TestHasPrefix(t *testing.T) {
	if !HasPrefix("hello", []string{"he", "wo"}) {
		t.Fatal("expected true")
	}
}

func TestJoin(t *testing.T) {
	got := Join([]string{"a", "b", "c"}, ",")
	if got != "a,b,c" {
		t.Fatalf("Join = %q, want a,b,c", got)
	}
}
