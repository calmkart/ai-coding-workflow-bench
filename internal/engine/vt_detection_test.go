package engine

import (
	"testing"

	"github.com/calmkart/ai-coding-workflow-bench/internal/config"
)

func TestCountCriticalVTFailures_NoCritical(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "medium", Detection: "e2e test case"},
		{ID: "VT-2", Severity: "high", Detection: "go build"},
	}
	result := &VerifyResult{L1Build: false, L4Passed: 0, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 critical failures (no critical VTs), got %d", got)
	}
}

func TestCountCriticalVTFailures_CriticalBuildFail(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "go build"},
	}
	result := &VerifyResult{L1Build: false, L4Passed: 0, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 critical failure (build fail), got %d", got)
	}
}

func TestCountCriticalVTFailures_CriticalBuildPass(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "go build"},
	}
	result := &VerifyResult{L1Build: true, L4Passed: 5, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 critical failures (build pass), got %d", got)
	}
}

func TestCountCriticalVTFailures_CriticalE2EFail(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-LEAK-01", Severity: "critical", Detection: "e2e test case"},
	}
	result := &VerifyResult{L1Build: true, L4Passed: 3, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 critical failure (e2e fail), got %d", got)
	}
}

func TestCountCriticalVTFailures_CriticalE2EPass(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-LEAK-01", Severity: "critical", Detection: "e2e test case"},
	}
	result := &VerifyResult{L1Build: true, L4Passed: 5, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 critical failures (all e2e pass), got %d", got)
	}
}

func TestCountCriticalVTFailures_CriticalE2ENoTests(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-LEAK-01", Severity: "critical", Detection: "e2e test case"},
	}
	result := &VerifyResult{L1Build: true, L4Passed: 0, L4Total: 0}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 critical failures (no e2e tests), got %d", got)
	}
}

func TestCountCriticalVTFailures_MultipleCritical(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "go build"},
		{ID: "VT-2", Severity: "critical", Detection: "e2e test case"},
		{ID: "VT-3", Severity: "medium", Detection: "e2e test case"},
	}
	result := &VerifyResult{L1Build: false, L4Passed: 3, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	if got != 2 {
		t.Errorf("expected 2 critical failures (build fail + e2e fail), got %d", got)
	}
}

func TestCountCriticalVTFailures_EmptyVTs(t *testing.T) {
	result := &VerifyResult{L1Build: true, L4Passed: 5, L4Total: 5}
	got := countCriticalVTFailures(nil, result)
	if got != 0 {
		t.Errorf("expected 0 critical failures (no VTs), got %d", got)
	}
}

func TestCountCriticalVTFailures_UnknownDetection(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "some-unknown-tool"},
	}
	result := &VerifyResult{L1Build: false, L4Passed: 0, L4Total: 5}
	got := countCriticalVTFailures(vts, result)
	// Unknown detection methods are not mapped, so should not count.
	if got != 0 {
		t.Errorf("expected 0 critical failures (unknown detection), got %d", got)
	}
}

func TestCountCriticalVTFailures_CaseInsensitive(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "Critical", Detection: "Go Build"},
	}
	result := &VerifyResult{L1Build: false}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 (case insensitive match), got %d", got)
	}
}

// P9: Tests for expanded VT detection type mappings.

func TestCountCriticalVTFailures_BuildAlias(t *testing.T) {
	// "build" is an alias for "go build".
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "build"},
	}
	result := &VerifyResult{L1Build: false}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 (build alias), got %d", got)
	}
}

func TestCountCriticalVTFailures_UnitTestFail(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "unit test"},
	}
	result := &VerifyResult{L1Build: true, L2Passed: 5, L2Total: 8}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 (unit test fail), got %d", got)
	}
}

func TestCountCriticalVTFailures_UnitTestPass(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "ut"},
	}
	result := &VerifyResult{L1Build: true, L2Passed: 8, L2Total: 8}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 (unit test pass), got %d", got)
	}
}

func TestCountCriticalVTFailures_GoTestAlias(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "go test"},
	}
	result := &VerifyResult{L1Build: true, L2Passed: 0, L2Total: 3}
	got := countCriticalVTFailures(vts, result)
	if got != 1 {
		t.Errorf("expected 1 (go test fail), got %d", got)
	}
}

func TestCountCriticalVTFailures_LintIssues(t *testing.T) {
	detections := []string{"go vet", "staticcheck", "lint", "errcheck", "bodyclose"}
	for _, det := range detections {
		t.Run(det, func(t *testing.T) {
			vts := []config.VerificationTarget{
				{ID: "VT-1", Severity: "critical", Detection: det},
			}
			result := &VerifyResult{L1Build: true, L3Issues: 3}
			got := countCriticalVTFailures(vts, result)
			if got != 1 {
				t.Errorf("expected 1 (lint issues for %q), got %d", det, got)
			}
		})
	}
}

func TestCountCriticalVTFailures_LintNoIssues(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "go vet"},
	}
	result := &VerifyResult{L1Build: true, L3Issues: 0}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 (no lint issues), got %d", got)
	}
}

func TestCountCriticalVTFailures_RaceDetector(t *testing.T) {
	detections := []string{"race detector", "go test -race", "goleak"}
	for _, det := range detections {
		t.Run(det, func(t *testing.T) {
			vts := []config.VerificationTarget{
				{ID: "VT-1", Severity: "critical", Detection: det},
			}
			result := &VerifyResult{L1Build: true, L2Passed: 5, L2Total: 8}
			got := countCriticalVTFailures(vts, result)
			if got != 1 {
				t.Errorf("expected 1 (race detector fail for %q), got %d", det, got)
			}
		})
	}
}

func TestCountCriticalVTFailures_RaceDetectorPass(t *testing.T) {
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "race detector"},
	}
	result := &VerifyResult{L1Build: true, L2Passed: 8, L2Total: 8}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 (race detector pass), got %d", got)
	}
}

func TestCountCriticalVTFailures_E2EAliases(t *testing.T) {
	aliases := []string{"e2e test", "e2e"}
	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			vts := []config.VerificationTarget{
				{ID: "VT-1", Severity: "critical", Detection: alias},
			}
			result := &VerifyResult{L1Build: true, L4Passed: 3, L4Total: 5}
			got := countCriticalVTFailures(vts, result)
			if got != 1 {
				t.Errorf("expected 1 (e2e fail for %q), got %d", alias, got)
			}
		})
	}
}

func TestCountCriticalVTFailures_UnitTestNoTests(t *testing.T) {
	// No unit tests exist (L2Total=0), should not count as failure.
	vts := []config.VerificationTarget{
		{ID: "VT-1", Severity: "critical", Detection: "unit test"},
	}
	result := &VerifyResult{L1Build: true, L2Passed: 0, L2Total: 0}
	got := countCriticalVTFailures(vts, result)
	if got != 0 {
		t.Errorf("expected 0 (no unit tests), got %d", got)
	}
}
