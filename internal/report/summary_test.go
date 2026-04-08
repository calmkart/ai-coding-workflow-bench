package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

func TestGenerateSummary(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	cs := 0.95

	runs := []*store.Run{
		{
			ID: "test-001", Tag: "smoke", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Benchmark Report: smoke") {
		t.Error("expected report header")
	}
	if !strings.Contains(output, "tier1/fix-handler-bug") {
		t.Error("expected task ID in report")
	}
	if !strings.Contains(output, "100.0%") {
		t.Error("expected 100% pass rate")
	}
	if !strings.Contains(output, "PASS") {
		t.Error("expected PASS in L1 column")
	}
	if !strings.Contains(output, "8/8") {
		t.Error("expected 8/8 in L2 column")
	}
	if !strings.Contains(output, "5/5") {
		t.Error("expected 5/5 in L4 column")
	}
}

func TestGenerateSummaryEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateSummary(&buf, nil)
	if err == nil {
		t.Error("expected error for empty runs")
	}
}

// TestGenerateSummary_DateIncludesTime verifies that the report date includes hours and minutes (item 4).
func TestGenerateSummary_DateIncludesTime(t *testing.T) {
	now := time.Date(2026, 3, 25, 14, 30, 0, 0, time.Local)
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95

	runs := []*store.Run{
		{
			ID: "test-date", Tag: "date-test", Workflow: "vanilla",
			TaskID: "tier1/fix", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// Should contain time portion, not just date.
	if !strings.Contains(output, "14:30") {
		t.Errorf("expected date to include time '14:30', got:\n%s", output)
	}
}

// TestGenerateSummary_RunsPerTaskVaries verifies that RunsPerTask shows a range when different (item 5).
func TestGenerateSummary_RunsPerTaskVaries(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-a1", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-a2", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 2, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-a3", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 3, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-b1", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// task-a has 3 runs, task-b has 1 run -> "1-3"
	if !strings.Contains(output, "Runs/task: 1-3") {
		t.Errorf("expected 'Runs/task: 1-3' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_WorkflowFromData verifies workflow is read from actual data (item 6).
func TestGenerateSummary_WorkflowFromData(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-wf1", Tag: "wf-test", Workflow: "custom-wf",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workflow: custom-wf") {
		t.Errorf("expected 'Workflow: custom-wf' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_MultipleWorkflows verifies "multiple" is shown when mixed workflows (item 6).
func TestGenerateSummary_MultipleWorkflows(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-mw1", Tag: "mw-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-mw2", Tag: "mw-test", Workflow: "custom",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workflow: multiple") {
		t.Errorf("expected 'Workflow: multiple' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_L1FailMarked verifies L1=FAIL rows are marked with **FAIL** and !! (item 8).
func TestGenerateSummary_L1FailMarked(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1Fail := false
	cs := 0.0

	runs := []*store.Run{
		{
			ID: "run-fail-mark", Tag: "fail-mark-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Fail, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "**FAIL**") {
		t.Errorf("expected **FAIL** markdown bold in L1 column, got:\n%s", output)
	}
	if !strings.Contains(output, "!!") {
		t.Errorf("expected '!!' warning marker in correctness column for L1=FAIL, got:\n%s", output)
	}
}
