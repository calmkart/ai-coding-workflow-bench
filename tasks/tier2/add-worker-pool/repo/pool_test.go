package workerpool

import "testing"

func TestNewPool(t *testing.T) {
	p := NewPool(4)
	if p == nil {
		t.Fatal("expected non-nil pool")
	}
}

func TestPanicOnZeroWorkers(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewPool(0)
}
