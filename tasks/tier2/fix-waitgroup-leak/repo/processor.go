package processor

import (
	"fmt"
	"strings"
	"sync"
)

// ProcessAll processes all items concurrently and returns results.
// BUG: If any goroutine panics, wg.Done() is never called and Wait() blocks forever.
func ProcessAll(items []string) ([]string, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []string
	var firstErr error

	for _, item := range items {
		wg.Add(1)
		go func(s string) {
			// BUG: wg.Done() is not deferred - if processItem panics, Done is never called
			result := processItem(s)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			wg.Done() // BUG: unreachable if processItem panics
		}(item)
	}

	wg.Wait()
	return results, firstErr
}

// processItem processes a single item. Panics on "panic" input.
func processItem(s string) string {
	if s == "panic" {
		panic("unexpected input: " + s)
	}
	return strings.ToUpper(s)
}

// ProcessBatch is similar but with error collection.
// BUG: Same WaitGroup leak issue.
func ProcessBatch(items []string, fn func(string) (string, error)) ([]string, []error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var results []string
	var errs []error

	for _, item := range items {
		wg.Add(1)
		go func(s string) {
			// BUG: not deferred
			result, err := fn(s)
			mu.Lock()
			if err != nil {
				errs = append(errs, fmt.Errorf("item %q: %w", s, err))
			} else {
				results = append(results, result)
			}
			mu.Unlock()
			wg.Done()
		}(item)
	}

	wg.Wait()
	return results, errs
}
