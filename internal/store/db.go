// Package store provides SQLite persistence for benchmark runs.
//
// Spec: .planning/workflow-bench.md (section 5.4)
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
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
	L1Build         *bool
	L2UtPassed      *int
	L2UtTotal       *int
	L3LintIssues    *int
	L4E2EPassed     *int
	L4E2ETotal      *int
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
// @implements REQ-STORE-OPEN (open SQLite with WAL mode and migration)
func Open(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
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

// InsertRun inserts a new run record.
//
// @implements REQ-STORE-INSERT (insert a new run record)
func (d *DB) InsertRun(r *Run) error {
	_, err := d.db.Exec(`
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

	result, err := d.db.Exec(`
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
			efficiency_score = ?,
			final_score = ?
		WHERE id = ?`,
		r.Status, finishedAt,
		r.L1Build, r.L2UtPassed, r.L2UtTotal, r.L3LintIssues,
		r.L4E2EPassed, r.L4E2ETotal, r.CorrectnessScore,
		r.InputTokens, r.OutputTokens, r.TotalTokens,
		r.CostUSD, r.WallTimeSecs, r.ToolUses,
		r.SecurityIssues, r.SecurityFail,
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
