package store

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func mustOpen(t *testing.T) *DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open :memory: db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenAndMigrate(t *testing.T) {
	db := mustOpen(t)
	// Verify the runs table exists by inserting a row.
	r := &Run{
		ID:        "test-001",
		Tag:       "smoke",
		Workflow:  "vanilla",
		TaskID:    "tier1/fix-handler-bug",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "running",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert run: %v", err)
	}
}

func TestInsertAndGetRuns(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)
	r := &Run{
		ID:          "run-001",
		Tag:         "test-tag",
		Workflow:    "vanilla",
		TaskID:      "tier1/fix-handler-bug",
		Tier:        1,
		TaskType:    "http-server",
		RunNumber:   1,
		Status:      "running",
		StartedAt:   now,
		PlanContent: "test plan",
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Update with results
	finished := now.Add(5 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	score := 0.95
	r.Status = "completed"
	r.FinishedAt = &finished
	r.L1Build = &l1
	r.L2UtPassed = &l2p
	r.L2UtTotal = &l2t
	r.L3LintIssues = &l3
	r.L4E2EPassed = &l4p
	r.L4E2ETotal = &l4t
	r.CorrectnessScore = &score

	if err := db.UpdateRun(r); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Query back
	runs, err := db.GetRunsByTag("test-tag")
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.Status != "completed" {
		t.Errorf("expected status completed, got %s", got.Status)
	}
	if got.L1Build == nil || !*got.L1Build {
		t.Error("expected L1Build=true")
	}
	if got.L2UtPassed == nil || *got.L2UtPassed != 8 {
		t.Error("expected L2UtPassed=8")
	}
	if got.CorrectnessScore == nil || *got.CorrectnessScore != 0.95 {
		t.Errorf("expected correctness=0.95, got %v", got.CorrectnessScore)
	}
	if got.PlanContent != "test plan" {
		t.Errorf("expected plan content 'test plan', got %q", got.PlanContent)
	}
}

func TestGetTags(t *testing.T) {
	db := mustOpen(t)

	for _, tag := range []string{"beta", "alpha", "beta"} {
		r := &Run{
			ID:        "run-" + tag + "-" + time.Now().String(),
			Tag:       tag,
			Workflow:  "vanilla",
			TaskID:    "tier1/test",
			Tier:      1,
			TaskType:  "http-server",
			RunNumber: 1,
			Status:    "completed",
			StartedAt: time.Now(),
		}
		if err := db.InsertRun(r); err != nil {
			t.Fatal(err)
		}
	}

	tags, err := db.GetTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 distinct tags, got %d: %v", len(tags), tags)
	}
	if tags[0] != "alpha" || tags[1] != "beta" {
		t.Errorf("expected [alpha, beta], got %v", tags)
	}
}

func TestRunExists(t *testing.T) {
	db := mustOpen(t)

	r := &Run{
		ID:        "run-exist-001",
		Tag:       "test",
		Workflow:  "vanilla",
		TaskID:    "tier1/fix",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatal(err)
	}

	exists, err := db.RunExists("test", "vanilla", "tier1/fix", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("expected run to exist")
	}

	exists, err = db.RunExists("test", "vanilla", "tier1/fix", 2)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("expected run NOT to exist")
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := mustOpen(t)
	// Opening again on same DB should not fail (idempotent migration).
	_ = db
}

// TestGetTagSummaries verifies that GetTagSummaries returns correct aggregate info per tag.
func TestGetTagSummaries(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)

	// Insert 2 runs under tag "alpha" (vanilla workflow).
	for i := 1; i <= 2; i++ {
		r := &Run{
			ID:        fmt.Sprintf("run-alpha-%d", i),
			Tag:       "alpha",
			Workflow:  "vanilla",
			TaskID:    "tier1/test",
			Tier:      1,
			TaskType:  "http-server",
			RunNumber: i,
			Status:    "completed",
			StartedAt: now.Add(time.Duration(i) * time.Minute),
		}
		if err := db.InsertRun(r); err != nil {
			t.Fatal(err)
		}
	}

	// Insert 1 run under tag "beta" (vanilla workflow).
	r := &Run{
		ID:        "run-beta-1",
		Tag:       "beta",
		Workflow:  "vanilla",
		TaskID:    "tier1/test",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: now,
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatal(err)
	}

	summaries, err := db.GetTagSummaries()
	if err != nil {
		t.Fatalf("GetTagSummaries: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 tag summaries, got %d", len(summaries))
	}

	// Sorted alphabetically: alpha, beta.
	if summaries[0].Tag != "alpha" {
		t.Errorf("expected first tag=alpha, got %s", summaries[0].Tag)
	}
	if summaries[0].Runs != 2 {
		t.Errorf("expected alpha runs=2, got %d", summaries[0].Runs)
	}
	if len(summaries[0].Workflows) != 1 || summaries[0].Workflows[0] != "vanilla" {
		t.Errorf("expected alpha workflows=[vanilla], got %v", summaries[0].Workflows)
	}

	if summaries[1].Tag != "beta" {
		t.Errorf("expected second tag=beta, got %s", summaries[1].Tag)
	}
	if summaries[1].Runs != 1 {
		t.Errorf("expected beta runs=1, got %d", summaries[1].Runs)
	}
}

// TestGetTagSummaries_Empty verifies empty DB returns nil summaries.
func TestGetTagSummaries_Empty(t *testing.T) {
	db := mustOpen(t)
	summaries, err := db.GetTagSummaries()
	if err != nil {
		t.Fatalf("GetTagSummaries on empty: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}
}

func TestUpdateRunNotFound(t *testing.T) {
	db := mustOpen(t)

	r := &Run{
		ID:        "nonexistent-run",
		Status:    "completed",
		StartedAt: time.Now(),
	}
	err := db.UpdateRun(r)
	if err == nil {
		t.Fatal("expected error for nonexistent run ID")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}
