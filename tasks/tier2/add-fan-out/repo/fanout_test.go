package fanout

import "testing"

func TestFanOutBasic(t *testing.T) {
	input := []int{1, 2, 3}
	results := FanOut(input, 2, func(n int) int { return n * 2 })

	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
}
