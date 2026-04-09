package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

func makeTrendRun(id, tag, taskID string, l1 bool, l4p, l4t int, corr float64, wallSecs float64, startedAt time.Time) *store.Run {
	finished := startedAt.Add(time.Duration(wallSecs) * time.Second)
	return &store.Run{
		ID:               id,
		Tag:              tag,
		Workflow:         "vanilla",
		TaskID:           taskID,
		Tier:             1,
		TaskType:         "http-server",
		RunNumber:        1,
		Status:           "completed",
		StartedAt:        startedAt,
		FinishedAt:       &finished,
		L1Build:          ptrBool(l1),
		L4E2EPassed:      ptrInt(l4p),
		L4E2ETotal:       ptrInt(l4t),
		CorrectnessScore: ptrFloat(corr),
		WallTimeSecs:     ptrFloat(wallSecs),
	}
}

func TestGenerateTrend_Basic(t *testing.T) {
	t1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)

	tagRunSets := []TagRunSet{
		{
			Tag: "v1",
			Runs: []*store.Run{
				makeTrendRun("v1-1", "v1", "tier1/task-a", true, 5, 5, 0.82, 150, t1),
				makeTrendRun("v1-2", "v1", "tier1/task-b", true, 3, 5, 0.60, 150, t1),
			},
		},
		{
			Tag: "v2",
			Runs: []*store.Run{
				makeTrendRun("v2-1", "v2", "tier1/task-a", true, 5, 5, 0.88, 135, t2),
				makeTrendRun("v2-2", "v2", "tier1/task-b", true, 5, 5, 0.85, 135, t2),
			},
		},
		{
			Tag: "v3",
			Runs: []*store.Run{
				makeTrendRun("v3-1", "v3", "tier1/task-a", true, 5, 5, 0.95, 110, t3),
				makeTrendRun("v3-2", "v3", "tier1/task-b", true, 5, 5, 0.90, 110, t3),
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateTrend(&buf, tagRunSets)
	if err != nil {
		t.Fatalf("GenerateTrend: %v", err)
	}

	output := buf.String()

	// Check header.
	if !strings.Contains(output, "# Trend Report") {
		t.Error("expected '# Trend Report' header")
	}

	// Check all tags appear.
	for _, tag := range []string{"v1", "v2", "v3"} {
		if !strings.Contains(output, tag) {
			t.Errorf("expected tag %q in output", tag)
		}
	}

	// Check dates appear.
	if !strings.Contains(output, "2026-04-01") {
		t.Error("expected date 2026-04-01")
	}
	if !strings.Contains(output, "2026-04-09") {
		t.Error("expected date 2026-04-09")
	}

	// Check trend summary line.
	if !strings.Contains(output, "Trend:") {
		t.Error("expected 'Trend:' summary line")
	}

	// v1: pass rate = 50% (1/2 fully pass L4), v3: 100% -> delta +50.0%
	if !strings.Contains(output, "Pass Rate +50.0%") {
		t.Errorf("expected 'Pass Rate +50.0%%' in trend, got:\n%s", output)
	}
}

func TestGenerateTrend_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateTrend(&buf, nil)
	if err == nil {
		t.Error("expected error for nil tagRunSets")
	}
}

func TestGenerateTrend_SingleTag(t *testing.T) {
	t1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	tagRunSets := []TagRunSet{
		{
			Tag: "v1",
			Runs: []*store.Run{
				makeTrendRun("v1-1", "v1", "tier1/task-a", true, 5, 5, 0.90, 120, t1),
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateTrend(&buf, tagRunSets)
	if err != nil {
		t.Fatalf("GenerateTrend single tag: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "v1") {
		t.Error("expected tag v1 in output")
	}
	// Single tag should not have trend summary.
	if strings.Contains(output, "Trend:") {
		t.Error("expected no trend summary for single tag")
	}
}

func TestGenerateTrend_NoWallTime(t *testing.T) {
	t1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)

	// Runs without wall time data.
	r1 := &store.Run{
		ID: "nw-1", Tag: "v1", Workflow: "vanilla",
		TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
		RunNumber: 1, Status: "completed", StartedAt: t1,
		L1Build:          ptrBool(true),
		L4E2EPassed:      ptrInt(5),
		L4E2ETotal:       ptrInt(5),
		CorrectnessScore: ptrFloat(0.80),
	}
	r2 := &store.Run{
		ID: "nw-2", Tag: "v2", Workflow: "vanilla",
		TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
		RunNumber: 1, Status: "completed", StartedAt: t2,
		L1Build:          ptrBool(true),
		L4E2EPassed:      ptrInt(5),
		L4E2ETotal:       ptrInt(5),
		CorrectnessScore: ptrFloat(0.95),
	}

	tagRunSets := []TagRunSet{
		{Tag: "v1", Runs: []*store.Run{r1}},
		{Tag: "v2", Runs: []*store.Run{r2}},
	}

	var buf bytes.Buffer
	err := GenerateTrend(&buf, tagRunSets)
	if err != nil {
		t.Fatalf("GenerateTrend no wall time: %v", err)
	}

	output := buf.String()
	// Wall time delta should be "-" when no wall time data.
	if !strings.Contains(output, "Wall Time -") {
		t.Errorf("expected 'Wall Time -' for missing wall time, got:\n%s", output)
	}
}

func TestGenerateTrendHTML_Basic(t *testing.T) {
	t1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)

	tagRunSets := []TagRunSet{
		{
			Tag: "v1",
			Runs: []*store.Run{
				makeTrendRun("v1-1", "v1", "tier1/task-a", true, 5, 5, 0.80, 150, t1),
			},
		},
		{
			Tag: "v2",
			Runs: []*store.Run{
				makeTrendRun("v2-1", "v2", "tier1/task-a", true, 5, 5, 0.95, 110, t2),
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateTrendHTML(&buf, tagRunSets)
	if err != nil {
		t.Fatalf("GenerateTrendHTML: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<html") {
		t.Error("expected HTML output")
	}
	if !strings.Contains(output, "Trend Report") {
		t.Error("expected 'Trend Report' in HTML output")
	}
	if !strings.Contains(output, "v1") {
		t.Error("expected tag v1 in HTML output")
	}
}

func TestBuildTrendData_Deltas(t *testing.T) {
	t1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)

	tagRunSets := []TagRunSet{
		{
			Tag: "v1",
			Runs: []*store.Run{
				makeTrendRun("v1-1", "v1", "tier1/task-a", true, 5, 5, 1.0, 200, t1),
				makeTrendRun("v1-2", "v1", "tier1/task-b", false, 0, 5, 0.0, 200, t1),
			},
		},
		{
			Tag: "v2",
			Runs: []*store.Run{
				makeTrendRun("v2-1", "v2", "tier1/task-a", true, 5, 5, 1.0, 100, t2),
				makeTrendRun("v2-2", "v2", "tier1/task-b", true, 5, 5, 1.0, 100, t2),
			},
		},
	}

	data := BuildTrendData(tagRunSets)

	// v1: 1/2 pass = 50%, v2: 2/2 pass = 100% -> delta +50.0%
	if data.DeltaPassRate != "+50.0%" {
		t.Errorf("expected DeltaPassRate=+50.0%%, got %q", data.DeltaPassRate)
	}

	// v1 avg correctness = 0.50, v2 avg correctness = 1.00 -> delta +0.50
	if data.DeltaCorrectness != "+0.50" {
		t.Errorf("expected DeltaCorrectness=+0.50, got %q", data.DeltaCorrectness)
	}

	// Wall time: 200 -> 100 = -50%
	if data.DeltaWallTime != "-50%" {
		t.Errorf("expected DeltaWallTime=-50%%, got %q", data.DeltaWallTime)
	}
}
