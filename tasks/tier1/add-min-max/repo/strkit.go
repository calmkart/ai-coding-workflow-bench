// Package strkit provides utility functions.
package strkit

// Abs returns the absolute value of an integer.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Clamp restricts value to be within [min, max].
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// NOTE: Min and Max functions are missing - agent should add them
// Signatures:
//   func Min[T cmp.Ordered](a, b T) T
//   func Max[T cmp.Ordered](a, b T) T
