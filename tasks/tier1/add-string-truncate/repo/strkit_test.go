package strkit

import "testing"

func TestReverse(t *testing.T) {
	got := Reverse("hello")
	if got != "olleh" {
		t.Fatalf("Reverse(hello) = %q, want olleh", got)
	}
}

func TestCountWords(t *testing.T) {
	if CountWords("hello world") != 2 {
		t.Fatal("expected 2 words")
	}
}
