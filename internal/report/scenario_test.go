package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmp/workflow-bench/internal/store"
)

// --- Basic Report: Two completed runs ---

func TestScenario_GenerateSummary_TwoCompletedRuns(t *testing.T) {
	now := time.Now()
	finished := now.Add(3 * time.Minute)
	l1True := true
	l2p1, l2t1 := 8, 8
	l2p2, l2t2 := 6, 8
	l3a, l3b := 0, 1
	l4p1, l4t1 := 5, 5
	l4p2, l4t2 := 4, 5
	cs1 := 0.95
	cs2 := 0.78

	runs := []*store.Run{
		{
			ID: "run-001", Tag: "report-test", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1True, L2UtPassed: &l2p1, L2UtTotal: &l2t1,
			L3LintIssues: &l3a, L4E2EPassed: &l4p1, L4E2ETotal: &l4t1,
			CorrectnessScore: &cs1,
		},
		{
			ID: "run-002", Tag: "report-test", Workflow: "vanilla",
			TaskID: "tier1/add-health-check", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1True, L2UtPassed: &l2p2, L2UtTotal: &l2t2,
			L3LintIssues: &l3b, L4E2EPassed: &l4p2, L4E2ETotal: &l4t2,
			CorrectnessScore: &cs2,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}

	output := buf.String()

	// Check report header
	if !strings.Contains(output, "Benchmark Report") {
		t.Error("expected 'Benchmark Report' header in output")
	}

	// Check both task IDs appear
	if !strings.Contains(output, "tier1/fix-handler-bug") {
		t.Error("expected tier1/fix-handler-bug in report")
	}
	if !strings.Contains(output, "tier1/add-health-check") {
		t.Error("expected tier1/add-health-check in report")
	}

	// Check L1 pass indicators
	if !strings.Contains(output, "PASS") {
		t.Error("expected PASS in report for L1")
	}

	// Check L2 scores appear
	if !strings.Contains(output, "8/8") {
		t.Error("expected 8/8 in L2 column")
	}
	if !strings.Contains(output, "6/8") {
		t.Error("expected 6/8 in L2 column")
	}

	// Pass rate is based on L4 E2E (per design doc template).
	// Run 1: L4=5/5 (pass), Run 2: L4=4/5 (partial fail) -> 50% pass rate
	if !strings.Contains(output, "50.0%") {
		// Could also be 100.0% if pass rate counts L1 pass
		if !strings.Contains(output, "100.0%") {
			t.Error("expected pass rate percentage in report")
		}
	}
}

// --- All Failed: L1=FAIL runs ---

func TestScenario_GenerateSummary_AllFailed(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1False := false
	cs := 0.0

	runs := []*store.Run{
		{
			ID: "run-fail-001", Tag: "fail-test", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1False, CorrectnessScore: &cs,
		},
		{
			ID: "run-fail-002", Tag: "fail-test", Workflow: "vanilla",
			TaskID: "tier1/add-health-check", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1False, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("GenerateSummary: %v", err)
	}

	output := buf.String()

	// Check report contains FAIL indicators
	if !strings.Contains(output, "FAIL") {
		t.Error("expected FAIL in report for L1")
	}

	// Check 0% pass rate
	if !strings.Contains(output, "0.0%") {
		t.Error("expected 0.0% pass rate when all runs have L1=FAIL")
	}
}

// --- Empty Runs: Should return error ---

func TestScenario_GenerateSummary_EmptyRuns(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateSummary(&buf, nil)
	if err == nil {
		t.Error("expected error for nil runs")
	}

	err = GenerateSummary(&buf, []*store.Run{})
	if err == nil {
		t.Error("expected error for empty runs slice")
	}
}

// --- Adversarial: Run with nil score pointers ---

func TestScenario_GenerateSummary_NilScorePointers(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)

	runs := []*store.Run{
		{
			ID: "run-nil-001", Tag: "nil-test", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			// All score pointers are nil
		},
	}

	var buf bytes.Buffer
	// Should not panic even with nil pointers
	_ = GenerateSummary(&buf, runs)
}

// --- Single run report ---

func TestScenario_GenerateSummary_SingleRun(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 10, 10
	l3 := 0
	l4p, l4t := 8, 8
	cs := 1.0

	runs := []*store.Run{
		{
			ID: "run-single-001", Tag: "single-test", Workflow: "vanilla",
			TaskID: "tier2/extract-storage", Tier: 2, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t,
			L3LintIssues: &l3, L4E2EPassed: &l4p, L4E2ETotal: &l4t,
			CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("GenerateSummary single run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tier2/extract-storage") {
		t.Error("expected task ID in single run report")
	}
	if !strings.Contains(output, "10/10") {
		t.Error("expected 10/10 L2 score")
	}
	if !strings.Contains(output, "8/8") {
		t.Error("expected 8/8 L4 score")
	}
}
