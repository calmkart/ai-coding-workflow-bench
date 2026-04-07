package fanout

// FanOut processes input items concurrently using the given number of workers.
// Results are returned in the same order as input.
// TODO: Implement concurrent fan-out pattern.
func FanOut[T any, R any](input []T, workers int, fn func(T) R) []R {
	if workers < 1 {
		panic("workers must be >= 1")
	}
	// BUG: Serial processing - not actually concurrent
	results := make([]R, len(input))
	for i, item := range input {
		results[i] = fn(item)
	}
	return results
}
