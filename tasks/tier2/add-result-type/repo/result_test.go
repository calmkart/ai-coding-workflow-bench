package result

import (
	"errors"
	"testing"
)

func TestOk(t *testing.T) {
	r := Ok(42)
	if !r.IsOk() {
		t.Fatal("expected ok")
	}
	if r.Value != 42 {
		t.Fatalf("expected 42, got %d", r.Value)
	}
}

func TestFail(t *testing.T) {
	r := Fail[int](errors.New("boom"))
	if !r.IsErr() {
		t.Fatal("expected err")
	}
}
