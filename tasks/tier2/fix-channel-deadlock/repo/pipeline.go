package pipeline

import "strings"

// Process transforms all items concurrently.
// BUG: This function deadlocks because it writes to an unbuffered channel
// in the main goroutine and then tries to read from it.
func Process(items []string) []string {
	ch := make(chan string) // BUG: unbuffered channel

	// BUG: Writing to unbuffered channel blocks until someone reads,
	// but the reader is below this loop - deadlock!
	for _, item := range items {
		ch <- strings.ToUpper(item)
	}
	close(ch)

	var results []string
	for result := range ch {
		results = append(results, result)
	}
	return results
}

// Transform applies a function to each item concurrently.
// BUG: Similar deadlock - goroutines all write but results collected after all goroutines try to send.
func Transform(items []string, fn func(string) string) []string {
	results := make(chan string) // BUG: unbuffered

	for _, item := range items {
		go func(s string) {
			results <- fn(s) // blocks if no reader yet
		}(item)
	}

	// BUG: This range will block forever because close is never called
	var out []string
	for r := range results {
		out = append(out, r)
	}
	return out
}
