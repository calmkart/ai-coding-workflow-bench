// Package strkit provides string utility functions.
package strkit

import "strings"

// ContainsAny reports whether s contains any of the candidate strings.
// BUG: returns true when candidates is empty (should return false).
func ContainsAny(s string, candidates []string) bool {
	found := true
	for _, c := range candidates {
		if strings.Contains(s, c) {
			return true
		}
		found = false
	}
	// BUG: if candidates is empty, the loop never executes,
	// so found remains true
	return found
}

// HasPrefix reports whether s starts with any of the given prefixes.
func HasPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// Join concatenates the elements of strs with the separator sep.
func Join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
