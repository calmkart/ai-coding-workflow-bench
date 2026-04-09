// Package store provides SQLite persistence for benchmark runs.
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var initialSchema string

// Run represents a single benchmark run record.
type Run struct {
	ID         string
	Tag        string
	Workflow   string
	TaskID     string
	Tier       int
	TaskType   string
	RunNumber  int
	Status     string // running|completed|failed|timeout
	StartedAt  time.Time
	FinishedAt *time.Time

	// Correctness
	L1Build          *bool
	L2UtPassed       *int
	L2UtTotal        *int
	L3LintIssues     *int
	L4E2EPassed      *int
	L4E2ETotal       *int
	CorrectnessScore *float64

	// Efficiency
	InputTokens  *int
	OutputTokens *int
	TotalTokens  *int
	CostUSD      *float64
	WallTimeSecs *float64
	ToolUses     *int

	// Code metrics
	SecurityIssues *int
	SecurityFail   *bool

	// Rework
	IterationCycles  *int
	FirstPassSuccess *bool

	// LLM Judge (rubric scores, 0-5 scale)
	RubricCorrectness          *float64
	RubricReadability          *float64
	RubricSimplicity           *float64
	RubricRobustness           *float64
	RubricMinimality           *float64
	RubricMaintainability      *float64
	RubricGoIdioms             *float64
	RubricComposite            *float64
	RubricConsistencyWarnings  *string  // P13: comma-separated boolean-to-score consistency warnings

	// Composite
	EfficiencyScore *float64
	FinalScore      *float64

	// Raw
	PlanContent string
}

// DB wraps the SQLite database connection.
type DB struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite database at the given path and applies migrations.
// Use ":memory:" for in-memory databases (testing).
//
// It configures the connection for concurrent multi-process access:
//   - busy_timeout=30000: wait up to 30s for locks instead of failing with SQLITE_BUSY
//   - WAL journal mode: allows concurrent readers alongside a single writer
//   - MaxOpenConns=1: SQLite only supports one writer; pooling avoids contention
//
// @implements REQ-STORE-OPEN (open SQLite with WAL mode and migration)
func Open(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// SQLite only supports a single writer connection; limit the pool to avoid
	// Go-level contention on top of SQLite's own locking.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Set busy timeout to wait up to 30 seconds for locks instead of failing
	// immediately with SQLITE_BUSY when another process holds the lock.
	if _, err := db.Exec("PRAGMA busy_timeout = 30000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}

	// Enable WAL mode for concurrent safety.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	// Apply migrations.
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{db: db}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read user_version: %w", err)
	}

	migrations := []struct {
		Version int
		SQL     string
	}{
		{1, initialSchema},
		{2, `ALTER TABLE runs ADD COLUMN rubric_consistency_warnings TEXT;`},
	}

	for _, m := range migrations {
		if m.Version > version {
			if _, err := db.Exec(m.SQL); err != nil {
				return fmt.Errorf("migration v%d: %w", m.Version, err)
			}
			if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version)); err != nil {
				return fmt.Errorf("set user_version %d: %w", m.Version, err)
			}
		}
	}
	return nil
}

// execWithRetry executes a write query with retry logic for SQLITE_BUSY errors.
// When multiple processes compete for the write lock, busy_timeout handles most
// contention, but transient lock failures can still occur. This retries up to 5
// times with linear backoff (200ms, 400ms, 600ms, ...) before giving up.
func (d *DB) execWithRetry(query string, args ...any) (sql.Result, error) {
	const maxRetries = 5
	var result sql.Result
	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err = d.db.Exec(query, args...)
		if err == nil {
			return result, nil
		}
		// Only retry on lock-related errors; anything else is a real failure.
		errMsg := err.Error()
		if !strings.Contains(errMsg, "SQLITE_BUSY") && !strings.Contains(errMsg, "database is locked") {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}
	return nil, fmt.Errorf("after %d retries: %w", maxRetries, err)
}

// InsertRun inserts a new run record.
// Uses retry logic to handle transient SQLITE_BUSY errors from concurrent writers.
//
// @implements REQ-STORE-INSERT (insert a new run record)
func (d *DB) InsertRun(r *Run) error {
	_, err := d.execWithRetry(`
		INSERT INTO runs (id, tag, workflow, task_id, tier, task_type, run_number, status, started_at, plan_content)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Tag, r.Workflow, r.TaskID, r.Tier, r.TaskType, r.RunNumber, r.Status,
		r.StartedAt.Format(time.RFC3339), r.PlanContent,
	)
	if err != nil {
		return fmt.Errorf("insert run %s: %w", r.ID, err)
	}
	return nil
}

// UpdateRun updates a run with results after completion.
//
// @implements REQ-STORE-UPDATE (update run results after completion)
func (d *DB) UpdateRun(r *Run) error {
	var finishedAt *string
	if r.FinishedAt != nil {
		s := r.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	result, err := d.execWithRetry(`
		UPDATE runs SET
			status = ?,
			finished_at = ?,
			l1_build = ?,
			l2_ut_passed = ?,
			l2_ut_total = ?,
			l3_lint_issues = ?,
			l4_e2e_passed = ?,
			l4_e2e_total = ?,
			correctness_score = ?,
			input_tokens = ?,
			output_tokens = ?,
			total_tokens = ?,
			cost_usd = ?,
			wall_time_secs = ?,
			tool_uses = ?,
			security_issues = ?,
			security_fail = ?,
			rubric_correctness = ?,
			rubric_readability = ?,
			rubric_simplicity = ?,
			rubric_robustness = ?,
			rubric_minimality = ?,
			rubric_maintainability = ?,
			rubric_go_idioms = ?,
			rubric_composite = ?,
			rubric_consistency_warnings = ?,
			efficiency_score = ?,
			final_score = ?
		WHERE id = ?`,
		r.Status, finishedAt,
		r.L1Build, r.L2UtPassed, r.L2UtTotal, r.L3LintIssues,
		r.L4E2EPassed, r.L4E2ETotal, r.CorrectnessScore,
		r.InputTokens, r.OutputTokens, r.TotalTokens,
		r.CostUSD, r.WallTimeSecs, r.ToolUses,
		r.SecurityIssues, r.SecurityFail,
		r.RubricCorrectness, r.RubricReadability, r.RubricSimplicity,
		r.RubricRobustness, r.RubricMinimality, r.RubricMaintainability,
		r.RubricGoIdioms, r.RubricComposite,
		r.RubricConsistencyWarnings,
		r.EfficiencyScore, r.FinalScore,
		r.ID,
	)
	if err != nil {
		return fmt.Errorf("update run %s: %w", r.ID, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected for run %s: %w", r.ID, err)
	}
	if rows == 0 {
		return fmt.Errorf("run %s not found", r.ID)
	}
	return nil
}

// GetRunsByTag returns all runs with the given tag.
//
// @implements REQ-STORE-QUERY (query runs by tag)
func (d *DB) GetRunsByTag(tag string) ([]*Run, error) {
	rows, err := d.db.Query(`
		SELECT id, tag, workflow, task_id, tier, task_type, run_number, status,
			started_at, finished_at,
			l1_build, l2_ut_passed, l2_ut_total, l3_lint_issues,
			l4_e2e_passed, l4_e2e_total, correctness_score,
			input_tokens, output_tokens, total_tokens, cost_usd,
			wall_time_secs, tool_uses,
			security_issues, security_fail,
			rubric_correctness, rubric_readability, rubric_simplicity,
			rubric_robustness, rubric_minimality, rubric_maintainability,
			rubric_go_idioms, rubric_composite,
			rubric_consistency_warnings,
			efficiency_score, final_score, plan_content
		FROM runs WHERE tag = ? ORDER BY task_id, run_number`, tag)
	if err != nil {
		return nil, fmt.Errorf("query runs by tag %q: %w", tag, err)
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		r := &Run{}
		var startedAt, finishedAt *string
		if err := rows.Scan(
			&r.ID, &r.Tag, &r.Workflow, &r.TaskID, &r.Tier, &r.TaskType,
			&r.RunNumber, &r.Status,
			&startedAt, &finishedAt,
			&r.L1Build, &r.L2UtPassed, &r.L2UtTotal, &r.L3LintIssues,
			&r.L4E2EPassed, &r.L4E2ETotal, &r.CorrectnessScore,
			&r.InputTokens, &r.OutputTokens, &r.TotalTokens, &r.CostUSD,
			&r.WallTimeSecs, &r.ToolUses,
			&r.SecurityIssues, &r.SecurityFail,
			&r.RubricCorrectness, &r.RubricReadability, &r.RubricSimplicity,
			&r.RubricRobustness, &r.RubricMinimality, &r.RubricMaintainability,
			&r.RubricGoIdioms, &r.RubricComposite,
			&r.RubricConsistencyWarnings,
			&r.EfficiencyScore, &r.FinalScore, &r.PlanContent,
		); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		// Fix 10: Log warnings for time parse errors instead of silently ignoring.
		if startedAt != nil {
			t, err := time.Parse(time.RFC3339, *startedAt)
			if err != nil {
				slog.Warn("parse started_at", "value", *startedAt, "error", err)
			}
			r.StartedAt = t
		}
		if finishedAt != nil {
			t, err := time.Parse(time.RFC3339, *finishedAt)
			if err != nil {
				slog.Warn("parse finished_at", "value", *finishedAt, "error", err)
			}
			r.FinishedAt = &t
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// GetTags returns all distinct tags.
func (d *DB) GetTags() ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT tag FROM runs ORDER BY tag")
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// RunExists checks if a completed run already exists for the given combination.
// Used for checkpoint/resume: skip already-completed runs.
func (d *DB) RunExists(tag, workflow, taskID string, runNumber int) (bool, error) {
	var count int
	err := d.db.QueryRow(`
		SELECT COUNT(*) FROM runs
		WHERE tag = ? AND workflow = ? AND task_id = ? AND run_number = ? AND status = 'completed'`,
		tag, workflow, taskID, runNumber).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check run exists: %w", err)
	}
	return count > 0, nil
}

// TagSummary holds aggregated information about a tag for display.
type TagSummary struct {
	Tag       string
	Runs      int
	MinDate   string
	MaxDate   string
	Workflows []string
}

// GetTagSummaries returns summary info for each distinct tag, ordered alphabetically.
func (d *DB) GetTagSummaries() ([]TagSummary, error) {
	rows, err := d.db.Query(`
		SELECT tag, COUNT(*) as run_count,
			MIN(started_at) as min_date,
			MAX(started_at) as max_date
		FROM runs GROUP BY tag ORDER BY tag`)
	if err != nil {
		return nil, fmt.Errorf("query tag summaries: %w", err)
	}
	defer rows.Close()

	var summaries []TagSummary
	for rows.Next() {
		var s TagSummary
		if err := rows.Scan(&s.Tag, &s.Runs, &s.MinDate, &s.MaxDate); err != nil {
			return nil, fmt.Errorf("scan tag summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch distinct workflows per tag.
	for i, s := range summaries {
		wfRows, err := d.db.Query(`SELECT DISTINCT workflow FROM runs WHERE tag = ? ORDER BY workflow`, s.Tag)
		if err != nil {
			return nil, fmt.Errorf("query workflows for tag %q: %w", s.Tag, err)
		}
		for wfRows.Next() {
			var wf string
			if err := wfRows.Scan(&wf); err != nil {
				wfRows.Close()
				return nil, fmt.Errorf("scan workflow: %w", err)
			}
			summaries[i].Workflows = append(summaries[i].Workflows, wf)
		}
		wfRows.Close()
	}

	return summaries, nil
}

// DeleteIncompleteRun removes a non-completed run so it can be re-run.
func (d *DB) DeleteIncompleteRun(tag, workflow, taskID string, runNumber int) error {
	_, err := d.db.Exec(`
		DELETE FROM runs
		WHERE tag = ? AND workflow = ? AND task_id = ? AND run_number = ? AND status != 'completed'`,
		tag, workflow, taskID, runNumber)
	return err
}

// PairwiseResultRow represents a single pairwise comparison result stored in the DB.
//
// @implements P17 (Pairwise result persistence)
type PairwiseResultRow struct {
	ID                  string
	TagLeft             string
	TagRight            string
	RunIDLeft           string
	RunIDRight          string
	TaskID              string
	Dimension           string
	Winner              string
	Magnitude           string
	PositionConsistent  bool
	Reasoning           string
}

// InsertPairwiseResult inserts a pairwise comparison result into the database.
//
// @implements P17 (Pairwise result persistence)
func (d *DB) InsertPairwiseResult(r *PairwiseResultRow) error {
	_, err := d.execWithRetry(`
		INSERT INTO pairwise_results (id, tag_left, tag_right, run_id_left, run_id_right,
			task_id, dimension, winner, magnitude, position_consistent, reasoning)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.TagLeft, r.TagRight, r.RunIDLeft, r.RunIDRight,
		r.TaskID, r.Dimension, r.Winner, r.Magnitude, r.PositionConsistent, r.Reasoning,
	)
	if err != nil {
		return fmt.Errorf("insert pairwise result %s: %w", r.ID, err)
	}
	return nil
}

// GetPairwiseResults retrieves all pairwise comparison results for a tag pair.
//
// @implements P17 (Pairwise result retrieval)
func (d *DB) GetPairwiseResults(tagLeft, tagRight string) ([]*PairwiseResultRow, error) {
	rows, err := d.db.Query(`
		SELECT id, tag_left, tag_right, run_id_left, run_id_right,
			task_id, dimension, winner, magnitude, position_consistent, reasoning
		FROM pairwise_results
		WHERE tag_left = ? AND tag_right = ?
		ORDER BY task_id, dimension`, tagLeft, tagRight)
	if err != nil {
		return nil, fmt.Errorf("query pairwise results: %w", err)
	}
	defer rows.Close()

	var results []*PairwiseResultRow
	for rows.Next() {
		r := &PairwiseResultRow{}
		if err := rows.Scan(
			&r.ID, &r.TagLeft, &r.TagRight, &r.RunIDLeft, &r.RunIDRight,
			&r.TaskID, &r.Dimension, &r.Winner, &r.Magnitude,
			&r.PositionConsistent, &r.Reasoning,
		); err != nil {
			return nil, fmt.Errorf("scan pairwise result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetAllRuns returns all runs in the database.
//
// @implements P22 (merge support — read all runs for DB merge)
func (d *DB) GetAllRuns() ([]*Run, error) {
	rows, err := d.db.Query(`
		SELECT id, tag, workflow, task_id, tier, task_type, run_number, status,
			started_at, finished_at,
			l1_build, l2_ut_passed, l2_ut_total, l3_lint_issues,
			l4_e2e_passed, l4_e2e_total, correctness_score,
			input_tokens, output_tokens, total_tokens, cost_usd,
			wall_time_secs, tool_uses,
			security_issues, security_fail,
			rubric_correctness, rubric_readability, rubric_simplicity,
			rubric_robustness, rubric_minimality, rubric_maintainability,
			rubric_go_idioms, rubric_composite,
			rubric_consistency_warnings,
			efficiency_score, final_score, plan_content
		FROM runs ORDER BY tag, task_id, run_number`)
	if err != nil {
		return nil, fmt.Errorf("query all runs: %w", err)
	}
	defer rows.Close()

	var runs []*Run
	for rows.Next() {
		r := &Run{}
		var startedAt, finishedAt *string
		if err := rows.Scan(
			&r.ID, &r.Tag, &r.Workflow, &r.TaskID, &r.Tier, &r.TaskType,
			&r.RunNumber, &r.Status,
			&startedAt, &finishedAt,
			&r.L1Build, &r.L2UtPassed, &r.L2UtTotal, &r.L3LintIssues,
			&r.L4E2EPassed, &r.L4E2ETotal, &r.CorrectnessScore,
			&r.InputTokens, &r.OutputTokens, &r.TotalTokens, &r.CostUSD,
			&r.WallTimeSecs, &r.ToolUses,
			&r.SecurityIssues, &r.SecurityFail,
			&r.RubricCorrectness, &r.RubricReadability, &r.RubricSimplicity,
			&r.RubricRobustness, &r.RubricMinimality, &r.RubricMaintainability,
			&r.RubricGoIdioms, &r.RubricComposite,
			&r.RubricConsistencyWarnings,
			&r.EfficiencyScore, &r.FinalScore, &r.PlanContent,
		); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		if startedAt != nil {
			t, err := time.Parse(time.RFC3339, *startedAt)
			if err != nil {
				slog.Warn("parse started_at", "value", *startedAt, "error", err)
			}
			r.StartedAt = t
		}
		if finishedAt != nil {
			t, err := time.Parse(time.RFC3339, *finishedAt)
			if err != nil {
				slog.Warn("parse finished_at", "value", *finishedAt, "error", err)
			}
			r.FinishedAt = &t
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// InsertRunFull inserts a fully populated run record (for merge operations).
// Unlike InsertRun which only sets initial fields, this sets all columns.
//
// @implements P22 (merge support — insert complete run record)
func (d *DB) InsertRunFull(r *Run) error {
	var finishedAt *string
	if r.FinishedAt != nil {
		s := r.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	_, err := d.execWithRetry(`
		INSERT OR IGNORE INTO runs (
			id, tag, workflow, task_id, tier, task_type, run_number, status,
			started_at, finished_at,
			l1_build, l2_ut_passed, l2_ut_total, l3_lint_issues,
			l4_e2e_passed, l4_e2e_total, correctness_score,
			input_tokens, output_tokens, total_tokens, cost_usd,
			wall_time_secs, tool_uses,
			security_issues, security_fail,
			rubric_correctness, rubric_readability, rubric_simplicity,
			rubric_robustness, rubric_minimality, rubric_maintainability,
			rubric_go_idioms, rubric_composite,
			rubric_consistency_warnings,
			efficiency_score, final_score, plan_content
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Tag, r.Workflow, r.TaskID, r.Tier, r.TaskType, r.RunNumber, r.Status,
		r.StartedAt.Format(time.RFC3339), finishedAt,
		r.L1Build, r.L2UtPassed, r.L2UtTotal, r.L3LintIssues,
		r.L4E2EPassed, r.L4E2ETotal, r.CorrectnessScore,
		r.InputTokens, r.OutputTokens, r.TotalTokens, r.CostUSD,
		r.WallTimeSecs, r.ToolUses,
		r.SecurityIssues, r.SecurityFail,
		r.RubricCorrectness, r.RubricReadability, r.RubricSimplicity,
		r.RubricRobustness, r.RubricMinimality, r.RubricMaintainability,
		r.RubricGoIdioms, r.RubricComposite,
		r.RubricConsistencyWarnings,
		r.EfficiencyScore, r.FinalScore, r.PlanContent,
	)
	if err != nil {
		return fmt.Errorf("insert run full %s: %w", r.ID, err)
	}
	return nil
}

// MergeFrom reads all runs from the source database and inserts them into this database.
// Duplicate run IDs are silently ignored (INSERT OR IGNORE).
//
// @implements P22 (DB merge for multi-machine shard results)
func (d *DB) MergeFrom(sourcePath string) (int, error) {
	srcDB, err := Open(sourcePath)
	if err != nil {
		return 0, fmt.Errorf("open source DB %s: %w", sourcePath, err)
	}
	defer srcDB.Close()

	runs, err := srcDB.GetAllRuns()
	if err != nil {
		return 0, fmt.Errorf("get runs from %s: %w", sourcePath, err)
	}

	inserted := 0
	for _, r := range runs {
		if err := d.InsertRunFull(r); err != nil {
			slog.Warn("merge run", "id", r.ID, "error", err)
			continue
		}
		inserted++
	}
	return inserted, nil
}

// DeleteByTag deletes all runs with the given tag and returns the count of deleted rows.
//
// @implements P24 (clean command - delete by tag)
func (d *DB) DeleteByTag(tag string) (int, error) {
	result, err := d.execWithRetry("DELETE FROM runs WHERE tag = ?", tag)
	if err != nil {
		return 0, fmt.Errorf("delete runs by tag %q: %w", tag, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count deleted rows: %w", err)
	}
	return int(rows), nil
}

// DeleteOlderThan deletes all runs started before the given time and returns the count
// of deleted rows.
//
// @implements P24 (clean command - delete older than)
func (d *DB) DeleteOlderThan(before time.Time) (int, error) {
	result, err := d.execWithRetry("DELETE FROM runs WHERE started_at < ?", before.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete runs older than %s: %w", before.Format(time.RFC3339), err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("count deleted rows: %w", err)
	}
	return int(rows), nil
}

// GetTagTrendSummary retrieves per-tag summary data suitable for trend reports.
// Returns tag name, run count, pass count, average correctness, average wall time,
// and distinct task count for the given tag.
//
// @implements P23 (trend data query support)
func (d *DB) GetTagTrendSummary(tag string) (passCount, totalRuns, distinctTasks int, avgCorrectness, avgWallTime float64, minStartedAt string, err error) {
	err = d.db.QueryRow(`
		SELECT
			COUNT(*) as total_runs,
			COUNT(DISTINCT task_id) as distinct_tasks,
			COALESCE(AVG(correctness_score), 0) as avg_correctness,
			COALESCE(AVG(wall_time_secs), 0) as avg_wall_time,
			MIN(started_at) as min_started_at
		FROM runs WHERE tag = ?`, tag).Scan(&totalRuns, &distinctTasks, &avgCorrectness, &avgWallTime, &minStartedAt)
	if err != nil {
		return 0, 0, 0, 0, 0, "", fmt.Errorf("query tag trend summary for %q: %w", tag, err)
	}

	// Count passes separately.
	err = d.db.QueryRow(`
		SELECT COUNT(*) FROM runs
		WHERE tag = ? AND l1_build = 1 AND l4_e2e_total > 0 AND l4_e2e_passed = l4_e2e_total`,
		tag).Scan(&passCount)
	if err != nil {
		return 0, 0, 0, 0, 0, "", fmt.Errorf("query tag pass count for %q: %w", tag, err)
	}

	return passCount, totalRuns, distinctTasks, avgCorrectness, avgWallTime, minStartedAt, nil
}
