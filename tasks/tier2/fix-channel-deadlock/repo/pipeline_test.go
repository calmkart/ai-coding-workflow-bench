package pipeline

import "testing"

// Note: These tests may hang due to deadlocks in the current implementation.
// They serve as documentation of expected behavior after the fix.

func TestProcessEmpty(t *testing.T) {
	result := Process(nil)
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}
