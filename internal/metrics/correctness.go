// Package metrics provides benchmark score calculations.
package metrics

import "math"

// CorrectnessInput holds the inputs for correctness score calculation.
type CorrectnessInput struct {
	L1Build             bool
	L2Passed            int
	L2Total             int
	L3Issues            int
	L4Passed            int
	L4Total             int
	CriticalVTFailCount int
}

// CalculateCorrectness computes the correctness score from L1-L4 results.
//
// Formula:
//
//	if L1 == FAIL: correctness = 0.0
//	else:
//	  l2_score = ut_passed / ut_total (0.0-1.0)
//	  l3_score = max(0, 1.0 - lint_issues * 0.05) (each issue deducts 5%, min 0)
//	  l4_score = e2e_passed / e2e_total (0.0-1.0)
//	  correctness = 0.20 * l2_score + 0.10 * l3_score + 0.70 * l4_score
//	  correctness = max(0, correctness - 0.1 * critical_vt_fail_count)
//
// @implements REQ-CORRECTNESS (correctness score calculation per spec 4.2.A)
func CalculateCorrectness(input CorrectnessInput) float64 {
	if !input.L1Build {
		return 0.0
	}

	var l2Score float64
	if input.L2Total > 0 {
		l2Score = float64(input.L2Passed) / float64(input.L2Total)
	} else {
		l2Score = 1.0 // No tests = no failures = perfect.
	}

	l3Score := math.Max(0, 1.0-float64(input.L3Issues)*0.05)

	var l4Score float64
	if input.L4Total > 0 {
		l4Score = float64(input.L4Passed) / float64(input.L4Total)
	} else {
		l4Score = 1.0 // No E2E tests = no failures.
	}

	correctness := 0.20*l2Score + 0.10*l3Score + 0.70*l4Score

	// Critical VT deduction.
	if input.CriticalVTFailCount > 0 {
		correctness = math.Max(0, correctness-0.1*float64(input.CriticalVTFailCount))
	}

	return correctness
}
