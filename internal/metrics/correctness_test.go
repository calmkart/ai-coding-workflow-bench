package metrics

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestCalculateCorrectness(t *testing.T) {
	tests := []struct {
		name  string
		input CorrectnessInput
		want  float64
	}{
		{
			name:  "L1 fail -> 0",
			input: CorrectnessInput{L1Build: false},
			want:  0.0,
		},
		{
			name: "all pass",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 8, L2Total: 8,
				L3Issues: 0, L4Passed: 5, L4Total: 5,
			},
			// 0.20*1.0 + 0.10*1.0 + 0.70*1.0 = 1.0
			want: 1.0,
		},
		{
			name: "partial L2 and L4",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 6, L2Total: 8,
				L3Issues: 0, L4Passed: 3, L4Total: 5,
			},
			// 0.20*(6/8) + 0.10*1.0 + 0.70*(3/5) = 0.20*0.75 + 0.10 + 0.70*0.6 = 0.15 + 0.10 + 0.42 = 0.67
			want: 0.67,
		},
		{
			name: "lint issues",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 8, L2Total: 8,
				L3Issues: 4, L4Passed: 5, L4Total: 5,
			},
			// 0.20*1.0 + 0.10*(1.0-4*0.05) + 0.70*1.0 = 0.20 + 0.10*0.80 + 0.70 = 0.20 + 0.08 + 0.70 = 0.98
			want: 0.98,
		},
		{
			name: "20+ lint issues caps at 0",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 8, L2Total: 8,
				L3Issues: 25, L4Passed: 5, L4Total: 5,
			},
			// 0.20*1.0 + 0.10*0.0 + 0.70*1.0 = 0.90
			want: 0.90,
		},
		{
			name: "zero totals treated as 1.0",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 0, L2Total: 0,
				L3Issues: 0, L4Passed: 0, L4Total: 0,
			},
			// 0.20*1.0 + 0.10*1.0 + 0.70*1.0 = 1.0
			want: 1.0,
		},
		{
			name: "critical VT deduction",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 8, L2Total: 8,
				L3Issues: 0, L4Passed: 5, L4Total: 5,
				CriticalVTFailCount: 2,
			},
			// 1.0 - 0.2 = 0.8
			want: 0.8,
		},
		{
			name: "critical VT deduction with floor at 0",
			input: CorrectnessInput{
				L1Build: true, L2Passed: 2, L2Total: 8,
				L3Issues: 10, L4Passed: 1, L4Total: 5,
				CriticalVTFailCount: 5,
			},
			// base: 0.20*(2/8) + 0.10*(1-10*0.05) + 0.70*(1/5) = 0.05 + 0.05 + 0.14 = 0.24
			// deduction: 0.24 - 0.5 = -0.26 -> clamped to 0
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCorrectness(tt.input)
			if !approxEqual(got, tt.want) {
				t.Errorf("got %.4f, want %.4f", got, tt.want)
			}
		})
	}
}
