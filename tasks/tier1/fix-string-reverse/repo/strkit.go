// Package strkit provides string utility functions.
package strkit

// Reverse returns the reverse of the input string.
// BUG: reverses by byte instead of by rune, corrupting multi-byte UTF-8 characters.
func Reverse(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// ToUpper returns s with all characters converted to uppercase.
func ToUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// CountWords returns the number of whitespace-separated words in s.
func CountWords(s string) int {
	if s == "" {
		return 0
	}
	count := 0
	inWord := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}
