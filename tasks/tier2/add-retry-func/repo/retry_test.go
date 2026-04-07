package retry

import (
	"testing"
	"time"
)

func TestRetrySuccess(t *testing.T) {
	calls := 0
	err := Retry(func() error {
		calls++
		return nil
	}, 3, time.Millisecond)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryPanicOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	Retry(func() error { return nil }, 0, time.Millisecond)
}
