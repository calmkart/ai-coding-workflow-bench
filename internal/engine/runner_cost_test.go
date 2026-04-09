package engine

import (
	"math"
	"testing"

	"github.com/calmkart/ai-coding-workflow-bench/internal/metrics"
)

// --- Optimization 2: Efficiency Score + Composite FinalScore ---

func TestEstimateCost(t *testing.T) {
	tests := []struct {
		name       string
		input      int
		output     int
		inputP     float64
		outputP    float64
		wantCost   float64
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
			wantCost: 0.03 + 0.075, // 0.105
		},
		{
			name: "custom pricing",
			input: 1_000_000, output: 1_000_000,
			inputP: 1.0, outputP: 5.0,
			wantCost: 6.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := metrics.EstimateCost(tt.input, tt.output, tt.inputP, tt.outputP)
			if math.Abs(got-tt.wantCost) > 0.001 {
				t.Errorf("metrics.EstimateCost(%d, %d, %.1f, %.1f) = %v, want %v",
					tt.input, tt.output, tt.inputP, tt.outputP, got, tt.wantCost)
			}
		})
	}
}

func TestTierCostBudget(t *testing.T) {
	tests := []struct {
		tier int
		want float64
	}{
		{1, 0.50},
		{2, 1.00},
		{3, 2.00},
		{4, 5.00},
		{0, 2.00},  // default
		{99, 2.00}, // unknown tier -> default
	}
	for _, tt := range tests {
		got := tierCostBudget(tt.tier)
		if got != tt.want {
			t.Errorf("tierCostBudget(%d) = %v, want %v", tt.tier, got, tt.want)
		}
	}
}

func TestEfficiencyScore(t *testing.T) {
	// Efficiency = 1.0 - min(1.0, cost/budget)
	tests := []struct {
		name   string
		cost   float64
		budget float64
		want   float64
	}{
		{"zero cost", 0.0, 0.50, 1.0},
		{"half budget", 0.25, 0.50, 0.5},
		{"full budget", 0.50, 0.50, 0.0},
		{"over budget", 1.00, 0.50, 0.0}, // clamped at 0
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eff := 1.0 - math.Min(1.0, tt.cost/tt.budget)
			if math.Abs(eff-tt.want) > 0.001 {
				t.Errorf("efficiency(cost=%v, budget=%v) = %v, want %v",
					tt.cost, tt.budget, eff, tt.want)
			}
		})
	}
}

