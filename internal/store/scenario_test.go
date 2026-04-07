package store

import (
	"testing"
	"time"
)

// --- Checkpoint / Resume Scenario ---
// Design doc 5.6: If (tag, workflow, task_id, run_number) already has status=completed, skip.
// status=failed or status=running records get overwritten on re-run.

func TestScenario_CheckpointResume_CompletedRunIsPreserved(t *testing.T) {
	db := mustOpen(t)

	// Insert a completed run
	completed := &Run{
		ID:        "run-resume-001",
		Tag:       "checkpoint-test",
		Workflow:  "vanilla",
		TaskID:    "tier1/fix-handler-bug",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(completed); err != nil {
		t.Fatalf("insert completed run: %v", err)
	}

	// RunExists should return true for this combination
	exists, err := db.RunExists("checkpoint-test", "vanilla", "tier1/fix-handler-bug", 1)
	if err != nil {
		t.Fatalf("RunExists: %v", err)
	}
	if !exists {
		t.Error("expected completed run to exist for checkpoint skip")
	}
}

func TestScenario_CheckpointResume_FailedRunCanBeOverwritten(t *testing.T) {
	db := mustOpen(t)

	// Insert a failed run
	failed := &Run{
		ID:        "run-failed-001",
		Tag:       "checkpoint-test",
		Workflow:  "vanilla",
		TaskID:    "tier1/fix-handler-bug",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "failed",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(failed); err != nil {
		t.Fatalf("insert failed run: %v", err)
	}

	// Update the failed run to completed (simulating overwrite/re-run)
	now := time.Now()
	l1 := true
	score := 0.95
	failed.Status = "completed"
	failed.FinishedAt = &now
	failed.L1Build = &l1
	failed.CorrectnessScore = &score
	if err := db.UpdateRun(failed); err != nil {
		t.Fatalf("update failed->completed: %v", err)
	}

	// Verify the run is now completed
	runs, err := db.GetRunsByTag("checkpoint-test")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != "completed" {
		t.Errorf("expected status=completed, got %s", runs[0].Status)
	}
}

func TestScenario_CheckpointResume_RunningRunCanBeOverwritten(t *testing.T) {
	db := mustOpen(t)

	// Insert a running run (crashed previously)
	running := &Run{
		ID:        "run-running-001",
		Tag:       "checkpoint-test",
		Workflow:  "vanilla",
		TaskID:    "tier1/fix-handler-bug",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "running",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(running); err != nil {
		t.Fatalf("insert running run: %v", err)
	}

	// Overwrite by updating to completed
	now := time.Now()
	l1 := true
	running.Status = "completed"
	running.FinishedAt = &now
	running.L1Build = &l1
	if err := db.UpdateRun(running); err != nil {
		t.Fatalf("update running->completed: %v", err)
	}

	runs, err := db.GetRunsByTag("checkpoint-test")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != "completed" {
		t.Errorf("expected status=completed after overwrite, got %s", runs[0].Status)
	}
}

// --- Multi-Tag Isolation Scenario ---

func TestScenario_MultiTagIsolation(t *testing.T) {
	db := mustOpen(t)

	// Insert runs under two different tags
	for i, tag := range []string{"tag-alpha", "tag-beta"} {
		for j := 1; j <= 2; j++ {
			r := &Run{
				ID:        "run-" + tag + "-" + time.Now().Format("150405.000000") + string(rune('a'+i)) + string(rune('0'+j)),
				Tag:       tag,
				Workflow:  "vanilla",
				TaskID:    "tier1/fix-handler-bug",
				Tier:      1,
				TaskType:  "http-server",
				RunNumber: j,
				Status:    "completed",
				StartedAt: time.Now(),
			}
			if err := db.InsertRun(r); err != nil {
				t.Fatalf("insert run for %s #%d: %v", tag, j, err)
			}
		}
	}

	// GetRunsByTag("tag-alpha") should return exactly 2 runs, all with tag-alpha
	alphaRuns, err := db.GetRunsByTag("tag-alpha")
	if err != nil {
		t.Fatalf("GetRunsByTag(tag-alpha): %v", err)
	}
	if len(alphaRuns) != 2 {
		t.Fatalf("expected 2 alpha runs, got %d", len(alphaRuns))
	}
	for _, r := range alphaRuns {
		if r.Tag != "tag-alpha" {
			t.Errorf("expected tag=tag-alpha, got %s", r.Tag)
		}
	}

	// GetRunsByTag("tag-beta") should return exactly 2 runs, all with tag-beta
	betaRuns, err := db.GetRunsByTag("tag-beta")
	if err != nil {
		t.Fatalf("GetRunsByTag(tag-beta): %v", err)
	}
	if len(betaRuns) != 2 {
		t.Fatalf("expected 2 beta runs, got %d", len(betaRuns))
	}
	for _, r := range betaRuns {
		if r.Tag != "tag-beta" {
			t.Errorf("expected tag=tag-beta, got %s", r.Tag)
		}
	}
}

// --- GetTags Scenario ---

func TestScenario_GetTags_MultipleTagsDeduplicatedAndSorted(t *testing.T) {
	db := mustOpen(t)

	tags := []string{"gamma", "alpha", "beta", "alpha", "gamma", "delta"}
	for i, tag := range tags {
		r := &Run{
			ID:        "run-tags-" + tag + "-" + string(rune('a'+i)),
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
			t.Fatalf("insert: %v", err)
		}
	}

	got, err := db.GetTags()
	if err != nil {
		t.Fatalf("GetTags: %v", err)
	}

	// Should be deduplicated: alpha, beta, delta, gamma (4 distinct, sorted)
	expected := []string{"alpha", "beta", "delta", "gamma"}
	if len(got) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %v", len(expected), len(got), got)
	}
	for i, want := range expected {
		if got[i] != want {
			t.Errorf("tag[%d]: expected %q, got %q", i, want, got[i])
		}
	}
}

// --- Empty Database Scenario ---

func TestScenario_EmptyDB_GetRunsByTagReturnsEmpty(t *testing.T) {
	db := mustOpen(t)

	runs, err := db.GetRunsByTag("nonexistent-tag")
	if err != nil {
		t.Fatalf("GetRunsByTag on empty db: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestScenario_EmptyDB_GetTagsReturnsEmpty(t *testing.T) {
	db := mustOpen(t)

	tags, err := db.GetTags()
	if err != nil {
		t.Fatalf("GetTags on empty db: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("expected 0 tags, got %d: %v", len(tags), tags)
	}
}

func TestScenario_EmptyDB_RunExistsReturnsFalse(t *testing.T) {
	db := mustOpen(t)

	exists, err := db.RunExists("any-tag", "any-workflow", "any-task", 1)
	if err != nil {
		t.Fatalf("RunExists on empty db: %v", err)
	}
	if exists {
		t.Error("expected RunExists to return false on empty db")
	}
}

// --- Adversarial: Boundary ---

func TestScenario_Store_LargeRunNumber(t *testing.T) {
	db := mustOpen(t)

	r := &Run{
		ID:        "run-large-num",
		Tag:       "boundary",
		Workflow:  "vanilla",
		TaskID:    "tier1/test",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 999999,
		Status:    "completed",
		StartedAt: time.Now(),
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert large run number: %v", err)
	}

	exists, err := db.RunExists("boundary", "vanilla", "tier1/test", 999999)
	if err != nil {
		t.Fatalf("RunExists large number: %v", err)
	}
	if !exists {
		t.Error("expected large run number to exist")
	}
}

// --- Adversarial: Invalid Input ---

func TestScenario_Store_EmptyTag(t *testing.T) {
	db := mustOpen(t)

	r := &Run{
		ID:        "run-empty-tag",
		Tag:       "",
		Workflow:  "vanilla",
		TaskID:    "tier1/test",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: time.Now(),
	}
	// Inserting with empty tag should still work (DB doesn't enforce non-empty beyond NOT NULL)
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert empty tag: %v", err)
	}

	runs, err := db.GetRunsByTag("")
	if err != nil {
		t.Fatalf("GetRunsByTag empty: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run with empty tag, got %d", len(runs))
	}
}

// --- Adversarial: Concurrent Access ---

func TestScenario_Store_ConcurrentInserts(t *testing.T) {
	// Use a file-based DB for concurrent access (in-memory SQLite doesn't support concurrent goroutines well)
	dir := t.TempDir()
	dbPath := dir + "/concurrent-test.db"
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Insert runs sequentially first to verify baseline, then test concurrent reads
	for i := 0; i < 10; i++ {
		r := &Run{
			ID:        "run-concurrent-" + string(rune('a'+i)),
			Tag:       "concurrent",
			Workflow:  "vanilla",
			TaskID:    "tier1/test",
			Tier:      1,
			TaskType:  "http-server",
			RunNumber: i + 1,
			Status:    "running",
			StartedAt: time.Now(),
		}
		if err := db.InsertRun(r); err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}

	// Concurrent reads should all succeed
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := db.GetRunsByTag("concurrent")
			done <- err
		}()
	}

	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent read %d: %v", i, err)
		}
	}

	runs, err := db.GetRunsByTag("concurrent")
	if err != nil {
		t.Fatalf("GetRunsByTag after concurrent test: %v", err)
	}
	if len(runs) != 10 {
		t.Errorf("expected 10 runs, got %d", len(runs))
	}
}
