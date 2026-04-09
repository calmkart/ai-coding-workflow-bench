package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// TestGenerateSummary_StabilityShownWhenMultipleRuns verifies that the Stability
// section appears when RunsPerTask > 1.
func TestGenerateSummary_StabilityShownWhenMultipleRuns(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1Pass := true
	l1Fail := false
	csPass := 1.0
	csFail := 0.0
	l4p5 := 5
	l4t5 := 5
	l4p0 := 0
	l4t0 := 5

	runs := []*store.Run{
		{
			ID: "stab-1a", Tag: "stab-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Pass, L4E2EPassed: &l4p5, L4E2ETotal: &l4t5, CorrectnessScore: &csPass,
		},
		{
			ID: "stab-1b", Tag: "stab-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 2, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Pass, L4E2EPassed: &l4p5, L4E2ETotal: &l4t5, CorrectnessScore: &csPass,
		},
		{
			ID: "stab-1c", Tag: "stab-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 3, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Fail, L4E2EPassed: &l4p0, L4E2ETotal: &l4t0, CorrectnessScore: &csFail,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// Stability section should appear.
	if !strings.Contains(output, "## Stability") {
		t.Errorf("expected '## Stability' section, got:\n%s", output)
	}
	// task-a: 2/3 passed (67%)
	if !strings.Contains(output, "2/3") {
		t.Errorf("expected '2/3' stability for task-a, got:\n%s", output)
	}
	if !strings.Contains(output, "67%") {
		t.Errorf("expected '67%%' stability, got:\n%s", output)
	}
}

// TestGenerateSummary_StabilityHiddenWhenSingleRun verifies that the Stability
// section is omitted when RunsPerTask == 1.
func TestGenerateSummary_StabilityHiddenWhenSingleRun(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "stab-single-1", Tag: "stab-single", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "stab-single-2", Tag: "stab-single", Workflow: "vanilla",
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
	if strings.Contains(output, "## Stability") {
		t.Errorf("expected no Stability section for single runs, got:\n%s", output)
	}
}

// TestBuildSummaryData_StabilityData verifies the StabilityData field is correctly populated.
func TestBuildSummaryData_StabilityData(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	l1Fail := false
	cs := 1.0
	csFail := 0.0
	l4p := 5
	l4t := 5
	l4p0 := 0

	runs := []*store.Run{
		{
			ID: "sd-1a", Tag: "sd-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
		{
			ID: "sd-1b", Tag: "sd-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 2, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Fail, L4E2EPassed: &l4p0, L4E2ETotal: &l4t, CorrectnessScore: &csFail,
		},
		{
			ID: "sd-2a", Tag: "sd-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
		{
			ID: "sd-2b", Tag: "sd-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 2, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
	}

	data := buildSummaryData(runs)

	if len(data.StabilityData) != 2 {
		t.Fatalf("expected 2 stability entries, got %d", len(data.StabilityData))
	}

	// task-a: 1/2 passed
	sa := data.StabilityData[0]
	if sa.TaskID != "tier1/task-a" {
		t.Errorf("expected task-a first, got %s", sa.TaskID)
	}
	if sa.PassCount != 1 || sa.TotalRuns != 2 {
		t.Errorf("task-a: expected 1/2, got %d/%d", sa.PassCount, sa.TotalRuns)
	}
	if sa.Stability != "1/2 (50%)" {
		t.Errorf("task-a stability: expected '1/2 (50%%)', got %q", sa.Stability)
	}

	// task-b: 2/2 passed
	sb := data.StabilityData[1]
	if sb.TaskID != "tier1/task-b" {
		t.Errorf("expected task-b second, got %s", sb.TaskID)
	}
	if sb.PassCount != 2 || sb.TotalRuns != 2 {
		t.Errorf("task-b: expected 2/2, got %d/%d", sb.PassCount, sb.TotalRuns)
	}
	if sb.Stability != "2/2 (100%)" {
		t.Errorf("task-b stability: expected '2/2 (100%%)', got %q", sb.Stability)
	}
}
