package retry

// Package retry provides a simple retry function.
// TODO: Implement Retry with exponential backoff.

import (
	"fmt"
	"time"
)

// Retry executes fn up to maxAttempts times with exponential backoff.
// baseDelay is the initial delay between retries.
// Returns nil on success, or the last error if all attempts fail.
//
// BUG: This is a stub - needs actual implementation.
func Retry(fn func() error, maxAttempts int, baseDelay time.Duration) error {
	if maxAttempts < 1 {
		panic("maxAttempts must be >= 1")
	}
	// TODO: implement retry with exponential backoff
	err := fn()
	if err != nil {
		return fmt.Errorf("failed: %w", err)
	}
	return nil
}
