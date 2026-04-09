package metrics

import (
	"math"
	"testing"
)

func TestWilsonCI_ZeroTotal(t *testing.T) {
	lower, upper := WilsonCI(0, 0)
	if lower != 0.0 || upper != 1.0 {
		t.Errorf("WilsonCI(0,0) = (%.4f, %.4f), want (0.0, 1.0)", lower, upper)
	}
}

func TestWilsonCI_AllPass(t *testing.T) {
	lower, upper := WilsonCI(20, 20)
	// All pass: lower should be high, upper should be 1.0.
	if lower < 0.80 {
		t.Errorf("WilsonCI(20,20) lower = %.4f, expected > 0.80", lower)
	}
	if upper != 1.0 {
		t.Errorf("WilsonCI(20,20) upper = %.4f, expected 1.0", upper)
	}
}

func TestWilsonCI_NonePass(t *testing.T) {
	lower, upper := WilsonCI(0, 20)
	// None pass: lower should be 0.0, upper should be low.
	if lower != 0.0 {
		t.Errorf("WilsonCI(0,20) lower = %.4f, expected 0.0", lower)
	}
	if upper > 0.20 {
		t.Errorf("WilsonCI(0,20) upper = %.4f, expected < 0.20", upper)
	}
}

func TestWilsonCI_SmallSample(t *testing.T) {
	// 2/3 passed: CI should be wide.
	lower, upper := WilsonCI(2, 3)
	if lower > 0.30 {
		t.Errorf("WilsonCI(2,3) lower = %.4f, expected < 0.30", lower)
	}
	if upper < 0.80 {
		t.Errorf("WilsonCI(2,3) upper = %.4f, expected > 0.80", upper)
	}
	// Width should be at least 0.5 for such a small sample.
	width := upper - lower
	if width < 0.40 {
		t.Errorf("WilsonCI(2,3) width = %.4f, expected > 0.40 for small sample", width)
	}
}

func TestWilsonCI_LargeSample(t *testing.T) {
	// 95/100: CI should be tight around 0.95.
	lower, upper := WilsonCI(95, 100)
	if math.Abs(lower-0.886) > 0.02 {
		t.Errorf("WilsonCI(95,100) lower = %.4f, expected ~0.886", lower)
	}
	if math.Abs(upper-0.982) > 0.02 {
		t.Errorf("WilsonCI(95,100) upper = %.4f, expected ~0.982", upper)
	}
}

func TestWilsonCI_BoundsClamp(t *testing.T) {
	// Ensure bounds are always [0, 1].
	lower, upper := WilsonCI(1, 1)
	if lower < 0 || lower > 1 {
		t.Errorf("lower out of [0,1]: %.4f", lower)
	}
	if upper < 0 || upper > 1 {
		t.Errorf("upper out of [0,1]: %.4f", upper)
	}
}

func TestCIOverlaps(t *testing.T) {
	tests := []struct {
		name   string
		l1, u1 float64
		l2, u2 float64
		want   bool
	}{
		{"fully overlapping", 0.3, 0.7, 0.5, 0.9, true},
		{"touching", 0.3, 0.5, 0.5, 0.9, true},
		{"non-overlapping", 0.1, 0.3, 0.5, 0.9, false},
		{"identical", 0.4, 0.6, 0.4, 0.6, true},
		{"first contains second", 0.1, 0.9, 0.3, 0.7, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CIOverlaps(tc.l1, tc.u1, tc.l2, tc.u2)
			if got != tc.want {
				t.Errorf("CIOverlaps(%.1f,%.1f,%.1f,%.1f) = %v, want %v",
					tc.l1, tc.u1, tc.l2, tc.u2, got, tc.want)
			}
		})
	}
}

func TestFormatCI(t *testing.T) {
	tests := []struct {
		passed, total int
		wantContains  string
	}{
		{0, 0, "N/A"},
		{5, 5, "100.0%"},
		{0, 5, "0.0%"},
	}
	for _, tc := range tests {
		result := FormatCI(tc.passed, tc.total)
		if tc.wantContains == "N/A" {
			if result != "N/A" {
				t.Errorf("FormatCI(%d,%d) = %q, want N/A", tc.passed, tc.total, result)
			}
			continue
		}
		// Check it contains the expected rate.
		found := false
		for i := 0; i <= len(result)-len(tc.wantContains); i++ {
			if result[i:i+len(tc.wantContains)] == tc.wantContains {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FormatCI(%d,%d) = %q, want to contain %q", tc.passed, tc.total, result, tc.wantContains)
		}
	}
}

func TestFormatCI_Structure(t *testing.T) {
	// Should be in format "X.X% [Y.Y-Z.Z]"
	result := FormatCI(7, 10)
	// Should contain brackets and a dash.
	hasBracket := false
	hasDash := false
	for _, c := range result {
		if c == '[' {
			hasBracket = true
		}
		if c == '-' {
			hasDash = true
		}
	}
	if !hasBracket || !hasDash {
		t.Errorf("FormatCI(7,10) = %q, expected format 'X.X%% [Y.Y-Z.Z]'", result)
	}
}

func TestIsLowSampleSize(t *testing.T) {
	tests := []struct {
		k    int
		want bool
	}{
		{0, true},
		{1, true},
		{3, true},
		{4, true},
		{5, false},
		{10, false},
	}
	for _, tc := range tests {
		got := IsLowSampleSize(tc.k)
		if got != tc.want {
			t.Errorf("IsLowSampleSize(%d) = %v, want %v", tc.k, got, tc.want)
		}
	}
}

func TestFormatFloat1(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.0, "0.0"},
		{95.0, "95.0"},
		{87.2, "87.2"},
		{98.6, "98.6"},
		{100.0, "100.0"},
	}
	for _, tc := range tests {
		got := formatFloat1(tc.input)
		if got != tc.want {
			t.Errorf("formatFloat1(%.1f) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
