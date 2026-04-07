package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/calmp/workflow-bench/internal/adapter"
	"github.com/calmp/workflow-bench/internal/config"
	"github.com/calmp/workflow-bench/internal/metrics"
	"github.com/calmp/workflow-bench/internal/store"
)

// RunConfig holds parameters for a benchmark run.
type RunConfig struct {
	Tag               string
	Workflow          string
	TaskSelector      string
	RunsPerTask       int
	PlanOverride      string // If set, use this plan instead of task's plan.md
	TimeoutMultiplier int
	TasksDir          string
	DBPath            string
	WorkflowCfg      map[string]any // Per-workflow config passed to adapter constructor
}

// Execute runs the full benchmark pipeline.
// For each task matching the selector, it runs the adapter the specified number of times,
// verifies results, collects metrics, and stores results in the database.
//
// @implements REQ-RUNNER (orchestrate: load tasks -> adapter.Run -> verify -> collect -> store)
func Execute(ctx context.Context, cfg RunConfig) error {
	// Open database.
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	// Discover and filter tasks.
	tasks, err := config.DiscoverTasks(cfg.TasksDir)
	if err != nil {
		return fmt.Errorf("discover tasks: %w", err)
	}
	filtered := config.FilterTasks(tasks, cfg.TaskSelector)
	if len(filtered) == 0 {
		return fmt.Errorf("no tasks matched filter %q", cfg.TaskSelector)
	}

	// Get adapter.
	// Resolve adapter type from WorkflowCfg if present (e.g., custom workflow
	// names like "my-workflow" map to an adapter type like "custom" via the
	// "adapter" key in WorkflowCfg). Fall back to the workflow name itself
	// for built-in adapters like "vanilla" or "v4-claude".
	adapterName := cfg.Workflow
	if cfg.WorkflowCfg != nil {
		if a, ok := cfg.WorkflowCfg["adapter"].(string); ok && a != "" {
			adapterName = a
		}
	}
	adpt, err := adapter.Get(adapterName, cfg.WorkflowCfg)
	if err != nil {
		return fmt.Errorf("get adapter: %w", err)
	}

	totalRuns := len(filtered) * cfg.RunsPerTask
	runIdx := 0
	passCount := 0
	skippedCount := 0

	fmt.Printf("\nworkflow-bench | workflow: %s | tasks: %d | runs/task: %d | total: %d\n\n",
		cfg.Workflow, len(filtered), cfg.RunsPerTask, totalRuns)

	for _, task := range filtered {
		for runNum := 1; runNum <= cfg.RunsPerTask; runNum++ {
			runIdx++
			runID := fmt.Sprintf("%s-%s-run%d-%d", cfg.Tag, task.ID, runNum, time.Now().UnixNano())

			// Check for existing completed run (checkpoint/resume).
			// Fix 7: Handle DB error from RunExists.
			exists, dbErr := db.RunExists(cfg.Tag, cfg.Workflow, task.ID, runNum)
			if dbErr != nil {
				slog.Error("check run exists", "task", task.ID, "run", runNum, "error", dbErr)
			}
			if exists {
				slog.Info("skipping completed run", "task", task.ID, "run", runNum)
				skippedCount++
				continue
			}
			// Delete any incomplete previous attempt.
			// Fix 8: Handle DB error from DeleteIncompleteRun.
			if err := db.DeleteIncompleteRun(cfg.Tag, cfg.Workflow, task.ID, runNum); err != nil {
				slog.Error("delete incomplete run", "task", task.ID, "run", runNum, "error", err)
			}

			fmt.Printf("[%d/%d] %s run#%d ...\n", runIdx, totalRuns, task.ID, runNum)

			passed, err := executeOneRun(ctx, db, adpt, task, runID, cfg, runNum)
			if err != nil {
				slog.Error("run failed", "task", task.ID, "run", runNum, "error", err)
				continue
			}
			if passed {
				passCount++
			}
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	executedRuns := totalRuns - skippedCount
	if skippedCount > 0 {
		fmt.Printf("Summary: %d/%d passed (%d skipped)\n", passCount, executedRuns, skippedCount)
	} else {
		fmt.Printf("Summary: %d/%d passed\n", passCount, totalRuns)
	}
	fmt.Printf("Results stored: tag=%s | %d runs in results.db\n", cfg.Tag, executedRuns)
	fmt.Printf("Run `workflow-bench report --tag %s` for full report.\n", cfg.Tag)

	return nil
}

// markRunFailed sets the run status and finishedAt, then persists to DB.
// Used to avoid repeating the "set status + set finishedAt + db.UpdateRun + log error" pattern.
func markRunFailed(db *store.DB, run *store.Run, status string) {
	run.Status = status
	finished := time.Now()
	run.FinishedAt = &finished
	if dbErr := db.UpdateRun(run); dbErr != nil {
		slog.Error("update run record", "run", run.ID, "error", dbErr)
	}
}

func executeOneRun(ctx context.Context, db *store.DB, adpt adapter.Adapter, task *config.TaskMeta, runID string, cfg RunConfig, runNum int) (bool, error) {
	now := time.Now()

	// Insert initial run record.
	run := &store.Run{
		ID:          runID,
		Tag:         cfg.Tag,
		Workflow:    cfg.Workflow,
		TaskID:      task.ID,
		Tier:        task.Tier,
		TaskType:    task.Type,
		RunNumber:   runNum,
		Status:      "running",
		StartedAt:   now,
	}

	if err := db.InsertRun(run); err != nil {
		return false, fmt.Errorf("insert run: %w", err)
	}

	// Load plan.
	planContent := cfg.PlanOverride
	if planContent == "" {
		var err error
		planContent, err = config.LoadPlan(task.Dir)
		if err != nil {
			markRunFailed(db, run, "failed")
			return false, fmt.Errorf("load plan: %w", err)
		}
	}
	run.PlanContent = planContent

	// Create isolated worktree.
	repoDir := filepath.Join(task.Dir, "repo")
	if err := EnsureGitRepo(repoDir); err != nil {
		markRunFailed(db, run, "failed")
		return false, fmt.Errorf("ensure git repo: %w", err)
	}

	worktreeDir, err := CreateWorktree(repoDir, runID)
	if err != nil {
		markRunFailed(db, run, "failed")
		return false, fmt.Errorf("create worktree: %w", err)
	}
	defer CleanupWorktree(repoDir, worktreeDir)

	// Copy go.mod to worktree if it exists in repo (worktree should have it).
	// The worktree is a detached checkout, so files should already be there.

	// Setup adapter.
	timeout := time.Duration(task.EstimatedMinutes) * time.Duration(cfg.TimeoutMultiplier) * time.Minute
	if timeout == 0 {
		timeout = 15 * time.Minute // Default timeout.
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := adpt.Setup(runCtx, worktreeDir); err != nil {
		markRunFailed(db, run, "failed")
		return false, fmt.Errorf("adapter setup: %w", err)
	}

	// Run adapter.
	output, err := adpt.Run(runCtx, worktreeDir, planContent)
	if err != nil {
		status := "failed"
		if runCtx.Err() == context.DeadlineExceeded {
			status = "timeout"
		}
		markRunFailed(db, run, status)
		return false, fmt.Errorf("adapter run: %w", err)
	}

	// Record token usage.
	if output.TokenUsage != nil {
		run.InputTokens = &output.TokenUsage.InputTokens
		run.OutputTokens = &output.TokenUsage.OutputTokens
		total := output.TokenUsage.InputTokens + output.TokenUsage.OutputTokens
		run.TotalTokens = &total
	}
	wallSecs := output.WallTime.Seconds()
	run.WallTimeSecs = &wallSecs
	if output.ToolUses > 0 {
		run.ToolUses = &output.ToolUses
	}

	// Generate and run verify.
	verifyDir, err := GenerateVerifyDir(VerifyConfig{
		TaskType: task.Type,
		TaskDir:  task.Dir,
		RunID:    runID,
	})
	if err != nil {
		return false, fmt.Errorf("generate verify: %w", err)
	}
	defer os.RemoveAll(verifyDir)

	verifyOutput, err := RunVerify(verifyDir, worktreeDir)
	if err != nil {
		slog.Warn("verify execution error", "error", err)
	}

	// Parse verify results.
	verifyResult, err := ParseVerifyOutput(verifyOutput)
	if err != nil {
		slog.Warn("parse verify output failed", "error", err, "output", verifyOutput)
		markRunFailed(db, run, "failed")
		return false, nil
	}

	// Populate run with verify results.
	run.L1Build = &verifyResult.L1Build
	run.L2UtPassed = &verifyResult.L2Passed
	run.L2UtTotal = &verifyResult.L2Total
	run.L3LintIssues = &verifyResult.L3Issues
	run.L4E2EPassed = &verifyResult.L4Passed
	run.L4E2ETotal = &verifyResult.L4Total

	// Calculate correctness score.
	correctness := metrics.CalculateCorrectness(metrics.CorrectnessInput{
		L1Build:    verifyResult.L1Build,
		L2Passed:   verifyResult.L2Passed,
		L2Total:    verifyResult.L2Total,
		L3Issues:   verifyResult.L3Issues,
		L4Passed:   verifyResult.L4Passed,
		L4Total:    verifyResult.L4Total,
		CriticalVTFailCount: 0, // P1: not yet implementing VT detection
	})
	run.CorrectnessScore = &correctness

	// For P1, final_score = correctness (no efficiency/quality/stability yet).
	run.FinalScore = &correctness

	run.Status = "completed"
	finished := time.Now()
	run.FinishedAt = &finished

	if err := db.UpdateRun(run); err != nil {
		return false, fmt.Errorf("update run: %w", err)
	}

	// Print result.
	l2Str := fmt.Sprintf("%d/%d", verifyResult.L2Passed, verifyResult.L2Total)
	l4Str := fmt.Sprintf("%d/%d", verifyResult.L4Passed, verifyResult.L4Total)
	l1Str := "PASS"
	if !verifyResult.L1Build {
		l1Str = "FAIL"
	}
	fmt.Printf("  ├─ adapter: %s | exit=%d\n", output.WallTime.Truncate(time.Second), output.ExitCode)
	fmt.Printf("  ├─ verify:  L1=%s L2=%s L3=%d L4=%s\n", l1Str, l2Str, verifyResult.L3Issues, l4Str)
	fmt.Printf("  └─ result:  %s (correctness=%.2f)\n\n", run.Status, correctness)

	// Fix 3: Match report/summary.go pass definition — L1 build + all L4 E2E tests passing.
	passed := verifyResult.L1Build && verifyResult.L4Total > 0 && verifyResult.L4Passed == verifyResult.L4Total
	return passed, nil
}
