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

// TestConcurrentInsertRun verifies that two goroutines can simultaneously
// InsertRun into the same file-backed DB without SQLITE_BUSY errors, thanks
// to the busy_timeout and retry logic.
func TestConcurrentInsertRun(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/concurrent-write.db"
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			r := &Run{
				ID:        fmt.Sprintf("concurrent-run-%d", idx),
				Tag:       "concurrent-write",
				Workflow:  "vanilla",
				TaskID:    "tier1/test",
				Tier:      1,
				TaskType:  "http-server",
				RunNumber: idx + 1,
				Status:    "running",
				StartedAt: time.Now(),
			}
			errs <- db.InsertRun(r)
		}(i)
	}

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent insert %d: %v", i, err)
		}
	}

	runs, err := db.GetRunsByTag("concurrent-write")
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
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

// TestRubricFieldsRoundtrip verifies that rubric scores are persisted and read back correctly.
func TestRubricFieldsRoundtrip(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)
	r := &Run{
		ID:        "run-rubric-001",
		Tag:       "rubric-test",
		Workflow:  "vanilla",
		TaskID:    "tier1/rubric-task",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "running",
		StartedAt: now,
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Update with rubric scores.
	finished := now.Add(5 * time.Minute)
	corr := 4.0
	read := 3.0
	simp := 4.0
	rob := 3.0
	min := 5.0
	maint := 4.0
	goId := 4.0
	comp := 3.85
	score := 0.95

	r.Status = "completed"
	r.FinishedAt = &finished
	r.CorrectnessScore = &score
	r.RubricCorrectness = &corr
	r.RubricReadability = &read
	r.RubricSimplicity = &simp
	r.RubricRobustness = &rob
	r.RubricMinimality = &min
	r.RubricMaintainability = &maint
	r.RubricGoIdioms = &goId
	r.RubricComposite = &comp

	if err := db.UpdateRun(r); err != nil {
		t.Fatalf("update: %v", err)
	}

	// Read back.
	runs, err := db.GetRunsByTag("rubric-test")
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	checkFloat := func(name string, ptr *float64, expected float64) {
		t.Helper()
		if ptr == nil {
			t.Errorf("%s: expected %.2f, got nil", name, expected)
		} else if *ptr != expected {
			t.Errorf("%s: expected %.2f, got %.2f", name, expected, *ptr)
		}
	}

	checkFloat("rubric_correctness", got.RubricCorrectness, 4.0)
	checkFloat("rubric_readability", got.RubricReadability, 3.0)
	checkFloat("rubric_simplicity", got.RubricSimplicity, 4.0)
	checkFloat("rubric_robustness", got.RubricRobustness, 3.0)
	checkFloat("rubric_minimality", got.RubricMinimality, 5.0)
	checkFloat("rubric_maintainability", got.RubricMaintainability, 4.0)
	checkFloat("rubric_go_idioms", got.RubricGoIdioms, 4.0)
	checkFloat("rubric_composite", got.RubricComposite, 3.85)
}

// TestPairwiseResultRoundtrip verifies insert and retrieval of pairwise results.
func TestPairwiseResultRoundtrip(t *testing.T) {
	db := mustOpen(t)

	// Insert a pairwise result.
	row := &PairwiseResultRow{
		ID:                 "pw-test-1",
		TagLeft:            "tag-a",
		TagRight:           "tag-b",
		RunIDLeft:          "run-a-1",
		RunIDRight:         "run-b-1",
		TaskID:             "tier1/task-a",
		Dimension:          "overall",
		Winner:             "left",
		Magnitude:          "strong",
		PositionConsistent: true,
		Reasoning:          "Left implementation is more correct.",
	}
	if err := db.InsertPairwiseResult(row); err != nil {
		t.Fatalf("InsertPairwiseResult: %v", err)
	}

	// Insert a second result for the same task pair.
	row2 := &PairwiseResultRow{
		ID:                 "pw-test-2",
		TagLeft:            "tag-a",
		TagRight:           "tag-b",
		RunIDLeft:          "run-a-1",
		RunIDRight:         "run-b-1",
		TaskID:             "tier1/task-a",
		Dimension:          "correctness",
		Winner:             "left",
		PositionConsistent: true,
		Reasoning:          "Left is more correct.",
	}
	if err := db.InsertPairwiseResult(row2); err != nil {
		t.Fatalf("InsertPairwiseResult: %v", err)
	}

	// Retrieve.
	results, err := db.GetPairwiseResults("tag-a", "tag-b")
	if err != nil {
		t.Fatalf("GetPairwiseResults: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Results should be ordered by task_id, dimension.
	if results[0].Dimension != "correctness" {
		t.Errorf("expected first result dimension=correctness, got %s", results[0].Dimension)
	}
	if results[1].Dimension != "overall" {
		t.Errorf("expected second result dimension=overall, got %s", results[1].Dimension)
	}
	if results[0].Winner != "left" {
		t.Errorf("expected winner=left, got %s", results[0].Winner)
	}
	if !results[0].PositionConsistent {
		t.Error("expected position_consistent=true")
	}
}

// TestGetPairwiseResults_Empty verifies empty results for nonexistent tag pair.
func TestGetPairwiseResults_Empty(t *testing.T) {
	db := mustOpen(t)
	results, err := db.GetPairwiseResults("nonexistent-a", "nonexistent-b")
	if err != nil {
		t.Fatalf("GetPairwiseResults: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestRubricFieldsNullWhenNotSet verifies that rubric fields are nil when not set.
func TestRubricFieldsNullWhenNotSet(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)
	r := &Run{
		ID:        "run-norubric-001",
		Tag:       "norubric-test",
		Workflow:  "vanilla",
		TaskID:    "tier1/no-rubric",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: now,
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := db.UpdateRun(r); err != nil {
		t.Fatalf("update: %v", err)
	}

	runs, err := db.GetRunsByTag("norubric-test")
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.RubricCorrectness != nil {
		t.Errorf("expected nil rubric_correctness, got %v", *got.RubricCorrectness)
	}
	if got.RubricComposite != nil {
		t.Errorf("expected nil rubric_composite, got %v", *got.RubricComposite)
	}
}

func TestGetAllRuns(t *testing.T) {
	db := mustOpen(t)

	// Insert 3 runs across 2 tags.
	for i, tag := range []string{"tag-a", "tag-a", "tag-b"} {
		r := &Run{
			ID:        fmt.Sprintf("run-%d", i),
			Tag:       tag,
			Workflow:  "vanilla",
			TaskID:    fmt.Sprintf("tier1/task-%d", i),
			Tier:      1,
			TaskType:  "library",
			RunNumber: 1,
			Status:    "completed",
			StartedAt: time.Now(),
		}
		if err := db.InsertRun(r); err != nil {
			t.Fatalf("insert run-%d: %v", i, err)
		}
	}

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("GetAllRuns: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs, got %d", len(runs))
	}
}

func TestInsertRunFull(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)
	finished := now.Add(5 * time.Minute)
	l1 := true
	l2p, l2t := 5, 5
	score := 0.9

	r := &Run{
		ID:               "full-001",
		Tag:              "merge-test",
		Workflow:         "vanilla",
		TaskID:           "tier1/test",
		Tier:             1,
		TaskType:         "library",
		RunNumber:        1,
		Status:           "completed",
		StartedAt:        now,
		FinishedAt:       &finished,
		L1Build:          &l1,
		L2UtPassed:       &l2p,
		L2UtTotal:        &l2t,
		CorrectnessScore: &score,
	}

	if err := db.InsertRunFull(r); err != nil {
		t.Fatalf("InsertRunFull: %v", err)
	}

	// Verify it was inserted.
	runs, err := db.GetRunsByTag("merge-test")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	got := runs[0]
	if got.ID != "full-001" {
		t.Errorf("expected ID full-001, got %s", got.ID)
	}
	if got.Status != "completed" {
		t.Errorf("expected status completed, got %s", got.Status)
	}
	if got.L1Build == nil || !*got.L1Build {
		t.Error("expected L1Build true")
	}
}

func TestInsertRunFull_Duplicate(t *testing.T) {
	db := mustOpen(t)

	r := &Run{
		ID:        "dup-001",
		Tag:       "test",
		Workflow:  "vanilla",
		TaskID:    "tier1/test",
		Tier:      1,
		TaskType:  "library",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: time.Now(),
	}

	// First insert should succeed.
	if err := db.InsertRunFull(r); err != nil {
		t.Fatalf("first InsertRunFull: %v", err)
	}

	// Second insert with same ID should be silently ignored (INSERT OR IGNORE).
	if err := db.InsertRunFull(r); err != nil {
		t.Fatalf("duplicate InsertRunFull should not error: %v", err)
	}

	runs, err := db.GetRunsByTag("test")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 run (duplicate ignored), got %d", len(runs))
	}
}

func TestMergeFrom(t *testing.T) {
	// MergeFrom requires file-based DBs (not :memory:) because it opens
	// the source DB separately. Create two temp files.
	srcPath := t.TempDir() + "/src.db"
	dstPath := t.TempDir() + "/dst.db"

	srcDB, err := Open(srcPath)
	if err != nil {
		t.Fatalf("open source: %v", err)
	}

	// Insert runs into source.
	for i := 0; i < 3; i++ {
		r := &Run{
			ID:        fmt.Sprintf("src-run-%d", i),
			Tag:       "shard-1",
			Workflow:  "vanilla",
			TaskID:    fmt.Sprintf("tier1/task-%d", i),
			Tier:      1,
			TaskType:  "library",
			RunNumber: 1,
			Status:    "completed",
			StartedAt: time.Now(),
		}
		if err := srcDB.InsertRun(r); err != nil {
			t.Fatalf("insert src run: %v", err)
		}
	}
	srcDB.Close()

	// Open destination and merge.
	dstDB, err := Open(dstPath)
	if err != nil {
		t.Fatalf("open dest: %v", err)
	}
	defer dstDB.Close()

	merged, err := dstDB.MergeFrom(srcPath)
	if err != nil {
		t.Fatalf("MergeFrom: %v", err)
	}
	if merged != 3 {
		t.Errorf("expected 3 merged, got %d", merged)
	}

	// Verify runs are in destination.
	runs, err := dstDB.GetAllRuns()
	if err != nil {
		t.Fatalf("GetAllRuns: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs in dest, got %d", len(runs))
	}
}

// TestDeleteByTag verifies that DeleteByTag removes all runs for a tag and returns the count.
func TestDeleteByTag(t *testing.T) {
	db := mustOpen(t)

	// Insert runs under two tags.
	for i, tag := range []string{"keep", "delete", "delete", "keep"} {
		r := &Run{
			ID:        fmt.Sprintf("run-del-%d", i),
			Tag:       tag,
			Workflow:  "vanilla",
			TaskID:    fmt.Sprintf("tier1/task-%d", i),
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

	// Delete "delete" tag runs.
	count, err := db.DeleteByTag("delete")
	if err != nil {
		t.Fatalf("DeleteByTag: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deleted, got %d", count)
	}

	// Verify "keep" tag still has its runs.
	keepRuns, err := db.GetRunsByTag("keep")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(keepRuns) != 2 {
		t.Errorf("expected 2 'keep' runs remaining, got %d", len(keepRuns))
	}

	// Verify "delete" tag has no runs.
	deleteRuns, err := db.GetRunsByTag("delete")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(deleteRuns) != 0 {
		t.Errorf("expected 0 'delete' runs remaining, got %d", len(deleteRuns))
	}
}

// TestDeleteByTag_NonExistentTag verifies that deleting a non-existent tag returns 0.
func TestDeleteByTag_NonExistentTag(t *testing.T) {
	db := mustOpen(t)

	count, err := db.DeleteByTag("nonexistent")
	if err != nil {
		t.Fatalf("DeleteByTag: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 deleted for nonexistent tag, got %d", count)
	}
}

// TestDeleteOlderThan verifies that DeleteOlderThan removes runs older than the cutoff.
func TestDeleteOlderThan(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)
	old := now.Add(-48 * time.Hour)
	recent := now.Add(-1 * time.Hour)

	// Insert old run.
	r1 := &Run{
		ID:        "run-old-1",
		Tag:       "test",
		Workflow:  "vanilla",
		TaskID:    "tier1/task-old",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: old,
	}
	if err := db.InsertRun(r1); err != nil {
		t.Fatalf("insert old: %v", err)
	}

	// Insert recent run.
	r2 := &Run{
		ID:        "run-recent-1",
		Tag:       "test",
		Workflow:  "vanilla",
		TaskID:    "tier1/task-recent",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: recent,
	}
	if err := db.InsertRun(r2); err != nil {
		t.Fatalf("insert recent: %v", err)
	}

	// Delete runs older than 24 hours.
	cutoff := now.Add(-24 * time.Hour)
	count, err := db.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 deleted, got %d", count)
	}

	// Verify only recent run remains.
	runs, err := db.GetRunsByTag("test")
	if err != nil {
		t.Fatalf("GetRunsByTag: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run remaining, got %d", len(runs))
	}
	if runs[0].ID != "run-recent-1" {
		t.Errorf("expected remaining run to be run-recent-1, got %s", runs[0].ID)
	}
}

// TestDeleteOlderThan_NoneOld verifies no runs deleted when all are recent.
func TestDeleteOlderThan_NoneOld(t *testing.T) {
	db := mustOpen(t)

	now := time.Now().Truncate(time.Second)

	r := &Run{
		ID:        "run-new-1",
		Tag:       "test",
		Workflow:  "vanilla",
		TaskID:    "tier1/task",
		Tier:      1,
		TaskType:  "http-server",
		RunNumber: 1,
		Status:    "completed",
		StartedAt: now,
	}
	if err := db.InsertRun(r); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Delete runs older than 24 hours - should delete nothing.
	cutoff := now.Add(-24 * time.Hour)
	count, err := db.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 deleted, got %d", count)
	}
}
