package ratelimiter

import "testing"

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	if rl == nil {
		t.Fatal("expected non-nil")
	}
}
