// Package strkit provides string utility functions.
package strkit

import "strings"

// Reverse returns the reverse of the input string.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// CountWords returns the number of whitespace-separated words in s.
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// NOTE: Truncate function is missing - agent should add it
// Signature: func Truncate(s string, maxLen int) string
// Behavior: truncate at word boundary, append "..." if truncated
