package metrics

import (
	"math"
	"testing"
)

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		output   int
		inputP   float64
		outputP  float64
		wantCost float64
	}{
		{
			name: "default pricing 1M each",
			input: 1_000_000, output: 1_000_000,
			inputP: 3.0, outputP: 15.0,
			wantCost: 18.0,
		},
		{
			name: "zero tokens",
			input: 0, output: 0,
			inputP: 3.0, outputP: 15.0,
			wantCost: 0.0,
		},
		{
			name: "small token counts",
			input: 10000, output: 5000,
			inputP: 3.0, outputP: 15.0,
			wantCost: 0.105, // 0.03 + 0.075
		},
		{
			name: "custom pricing",
			input: 1_000_000, output: 1_000_000,
			inputP: 1.0, outputP: 5.0,
			wantCost: 6.0,
		},
		{
			name: "only input tokens",
			input: 500_000, output: 0,
			inputP: 3.0, outputP: 15.0,
			wantCost: 1.5,
		},
		{
			name: "only output tokens",
			input: 0, output: 500_000,
			inputP: 3.0, outputP: 15.0,
			wantCost: 7.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateCost(tt.input, tt.output, tt.inputP, tt.outputP)
			if math.Abs(got-tt.wantCost) > 0.001 {
				t.Errorf("EstimateCost(%d, %d, %.1f, %.1f) = %v, want %v",
					tt.input, tt.output, tt.inputP, tt.outputP, got, tt.wantCost)
			}
		})
	}
}
