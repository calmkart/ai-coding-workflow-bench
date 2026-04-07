package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
// TODO: Implement Allow, Wait.
type RateLimiter struct {
	mu       sync.Mutex
	rate     float64 // tokens per second
	burst    int     // max tokens
	tokens   float64
	lastTime time.Time
}

// NewRateLimiter creates a new rate limiter.
// rate: tokens generated per second
// burst: maximum bucket capacity
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:     rate,
		burst:    burst,
		tokens:   float64(burst), // start full
		lastTime: time.Now(),
	}
}
