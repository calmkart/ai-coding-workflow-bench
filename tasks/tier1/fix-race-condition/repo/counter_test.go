package workerpool

import "testing"

func TestCounterInc(t *testing.T) {
	c := NewCounter()
	c.Inc()
	c.Inc()
	c.Inc()

	if c.Get() != 3 {
		t.Fatalf("expected 3, got %d", c.Get())
	}
}

func TestCounterReset(t *testing.T) {
	c := NewCounter()
	c.Inc()
	c.Inc()
	c.Reset()

	if c.Get() != 0 {
		t.Fatalf("expected 0 after reset, got %d", c.Get())
	}
}
