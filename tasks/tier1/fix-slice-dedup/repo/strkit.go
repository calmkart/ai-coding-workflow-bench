// Package strkit provides string and slice utility functions.
package strkit

// Dedup removes duplicate strings from a slice, preserving order.
// BUG: panics on nil or empty slice because it accesses items[0] directly.
func Dedup(items []string) []string {
	seen := map[string]bool{}
	// BUG: accessing items[0] without checking length
	seen[items[0]] = true
	result := []string{items[0]}

	for _, item := range items[1:] {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

// Filter returns a new slice containing only elements for which fn returns true.
func Filter(items []string, fn func(string) bool) []string {
	var result []string
	for _, item := range items {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map applies fn to each element and returns a new slice with the results.
func Map(items []string, fn func(string) string) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}
