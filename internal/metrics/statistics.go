// Package metrics provides benchmark score calculations.
//
// Spec: docs/optimization-roadmap.md P1
package metrics

import "math"

// WilsonCI computes the 95% Wilson Score confidence interval for a pass rate.
// It returns lower and upper bounds of the proportion (0.0-1.0 range).
//
// Wilson Score is preferred over normal approximation because it performs well
// even for small sample sizes and proportions near 0 or 1.
//
// @implements P1 (statistical methods - confidence interval)
func WilsonCI(passed, total int) (lower, upper float64) {
	if total == 0 {
		return 0, 1
	}
	z := 1.96 // 95% CI
	p := float64(passed) / float64(total)
	n := float64(total)
	denom := 1 + z*z/n
	center := (p + z*z/(2*n)) / denom
	spread := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / denom
	return math.Max(0, center-spread), math.Min(1, center+spread)
}

// CIOverlaps returns true if two confidence intervals overlap.
// Non-overlapping CIs suggest the difference is statistically significant.
func CIOverlaps(lower1, upper1, lower2, upper2 float64) bool {
	return lower1 <= upper2 && lower2 <= upper1
}

// FormatCI formats a pass rate with its Wilson Score 95% CI.
// Example output: "95.0% [87.2-98.6]"
func FormatCI(passed, total int) string {
	if total == 0 {
		return "N/A"
	}
	rate := float64(passed) / float64(total) * 100
	lower, upper := WilsonCI(passed, total)
	return formatCIString(rate, lower*100, upper*100)
}

// formatCIString formats rate and CI bounds into a display string.
func formatCIString(rate, lowerPct, upperPct float64) string {
	return formatFloat1(rate) + "% [" + formatFloat1(lowerPct) + "-" + formatFloat1(upperPct) + "]"
}

// formatFloat1 formats a float with 1 decimal place.
func formatFloat1(f float64) string {
	s := math.Round(f*10) / 10
	// Use Sprintf to get consistent formatting.
	return floatToStr(s)
}

func floatToStr(f float64) string {
	// Avoid importing fmt in this helper for test simplicity;
	// we'll just use a minimal approach.
	if f == 0 {
		return "0.0"
	}
	sign := ""
	if f < 0 {
		sign = "-"
		f = -f
	}
	whole := int(f)
	frac := int(math.Round((f - float64(whole)) * 10))
	if frac >= 10 {
		whole++
		frac = 0
	}
	return sign + itoa(whole) + "." + itoa(frac)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// IsLowSampleSize returns true when the number of runs K is below the
// minimum threshold for meaningful statistical inference.
func IsLowSampleSize(k int) bool {
	return k < 5
}
