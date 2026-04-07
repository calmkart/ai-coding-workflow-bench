package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	cb := NewCircuitBreaker(Options{
		MaxFailures: 3,
		Timeout:     time.Second,
	})

	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Fatalf("expected closed, got %v", cb.State())
	}

	_ = errors.New("test")
}
