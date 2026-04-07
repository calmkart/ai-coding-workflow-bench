package semaphore

import "testing"

func TestNewSemaphore(t *testing.T) {
	s := NewSemaphore(3)
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestPanicOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewSemaphore(0)
}
