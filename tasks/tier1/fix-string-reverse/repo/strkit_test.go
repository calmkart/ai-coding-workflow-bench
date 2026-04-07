package strkit

import "testing"

func TestReverseASCII(t *testing.T) {
	got := Reverse("hello")
	if got != "olleh" {
		t.Fatalf("Reverse(hello) = %q, want olleh", got)
	}
}

func TestReverseEmpty(t *testing.T) {
	got := Reverse("")
	if got != "" {
		t.Fatalf("Reverse('') = %q, want empty", got)
	}
}

func TestCountWords(t *testing.T) {
	if CountWords("hello world") != 2 {
		t.Fatal("expected 2 words")
	}
	if CountWords("") != 0 {
		t.Fatal("expected 0 words")
	}
}
