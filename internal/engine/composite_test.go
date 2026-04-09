package engine

import (
	"math"
	"testing"
)

func floatApprox(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func ptrF(f float64) *float64 { return &f }

func TestComputeCompositeScore_AllDimensions(t *testing.T) {
	// All four dimensions available.
	// correctness=0.8, efficiency=0.6, quality=0.9, stability=1.0
	// = 0.40*0.8 + 0.25*0.6 + 0.25*0.9 + 0.10*1.0
	// = 0.32 + 0.15 + 0.225 + 0.10 = 0.795
	got := ComputeCompositeScore(0.8, ptrF(0.6), ptrF(0.9), ptrF(1.0))
	if !floatApprox(got, 0.795) {
		t.Errorf("all dims: got %.4f, want 0.795", got)
	}
}

func TestComputeCompositeScore_OnlyCorrectness(t *testing.T) {
	// Only correctness available -> weight redistributed to correctness only = 1.0 * correctness.
	got := ComputeCompositeScore(0.85, nil, nil, nil)
	if !floatApprox(got, 0.85) {
		t.Errorf("only correctness: got %.4f, want 0.85", got)
	}
}

func TestComputeCompositeScore_CorrectnessAndEfficiency(t *testing.T) {
	// Correctness=0.40, Efficiency=0.25. Available weight = 0.65.
	// Redistributed: correctness = 0.40/0.65, efficiency = 0.25/0.65
	// Result = (0.40/0.65)*1.0 + (0.25/0.65)*0.5
	// = 0.6154 + 0.1923 = 0.8077
	got := ComputeCompositeScore(1.0, ptrF(0.5), nil, nil)
	expected := (0.40/0.65)*1.0 + (0.25/0.65)*0.5
	if !floatApprox(got, expected) {
		t.Errorf("corr+eff: got %.4f, want %.4f", got, expected)
	}
}

func TestComputeCompositeScore_CorrectnessAndQuality(t *testing.T) {
	// Correctness=0.40, Quality=0.25. Available weight = 0.65.
	got := ComputeCompositeScore(0.9, nil, ptrF(0.7), nil)
	expected := (0.40/0.65)*0.9 + (0.25/0.65)*0.7
	if !floatApprox(got, expected) {
		t.Errorf("corr+quality: got %.4f, want %.4f", got, expected)
	}
}

func TestComputeCompositeScore_AllPerfect(t *testing.T) {
	got := ComputeCompositeScore(1.0, ptrF(1.0), ptrF(1.0), ptrF(1.0))
	if !floatApprox(got, 1.0) {
		t.Errorf("all perfect: got %.4f, want 1.0", got)
	}
}

func TestComputeCompositeScore_AllZero(t *testing.T) {
	got := ComputeCompositeScore(0.0, ptrF(0.0), ptrF(0.0), ptrF(0.0))
	if !floatApprox(got, 0.0) {
		t.Errorf("all zero: got %.4f, want 0.0", got)
	}
}

func TestComputeCompositeScore_ThreeDimensions(t *testing.T) {
	// Without stability (weight 0.10). Available weight = 0.90.
	// correctness=1.0, efficiency=0.8, quality=0.6
	// = (0.40/0.90)*1.0 + (0.25/0.90)*0.8 + (0.25/0.90)*0.6
	got := ComputeCompositeScore(1.0, ptrF(0.8), ptrF(0.6), nil)
	expected := (0.40/0.90)*1.0 + (0.25/0.90)*0.8 + (0.25/0.90)*0.6
	if !floatApprox(got, expected) {
		t.Errorf("three dims: got %.4f, want %.4f", got, expected)
	}
}

// TestComputeCompositeScore_BackwardsCompat verifies that the old formula
// (0.60*correctness + 0.40*efficiency) is a reasonable approximation of the
// new formula when only correctness and efficiency are available.
// This is NOT an exact match since weights changed.
func TestComputeCompositeScore_OldFormulaCompare(t *testing.T) {
	// Old: 0.60*1.0 + 0.40*0.5 = 0.80
	// New: (0.40/0.65)*1.0 + (0.25/0.65)*0.5 = 0.6154 + 0.1923 = 0.8077
	got := ComputeCompositeScore(1.0, ptrF(0.5), nil, nil)
	// Just verify it's in a reasonable range (not exactly matching old formula).
	if got < 0.7 || got > 0.9 {
		t.Errorf("old formula compare: got %.4f, expected in [0.7, 0.9] range", got)
	}
}
