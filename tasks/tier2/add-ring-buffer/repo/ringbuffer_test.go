package ringbuffer

import "testing"

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int](5)
	if rb == nil {
		t.Fatal("expected non-nil")
	}
}

func TestPanicOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewRingBuffer[int](0)
}
