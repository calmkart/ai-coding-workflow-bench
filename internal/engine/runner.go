package engine

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/adapter"
	"github.com/calmkart/ai-coding-workflow-bench/internal/config"
	"github.com/calmkart/ai-coding-workflow-bench/internal/judge"
	"github.com/calmkart/ai-coding-workflow-bench/internal/metrics"
	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// progressTracker provides real-time progress feedback during benchmark execution.
// It tracks completed runs, pass rate, and computes an ETA based on elapsed time.
//
// @implements P19 (CLI progress ETA + intermediate summary)
type progressTracker struct {
	total     int
	completed atomic.Int64
	passed    atomic.Int64
	startTime time.Time
	mu        sync.Mutex
}

// newProgressTracker creates a new progress tracker for the given total number of runs.
func newProgressTracker(total int) *progressTracker {
	return &progressTracker{
		total:     total,
		startTime: time.Now(),
	}
}

// report logs a progress line after a run completes. It is safe for concurrent use.
func (p *progressTracker) report(taskID string, runNum int, success bool) {
	c := p.completed.Add(1)
	if success {
		p.passed.Add(1)
	}

	elapsed := time.Since(p.startTime)
	elapsedSecs := elapsed.Seconds()

	var etaStr string
	if elapsedSecs > 0 && c < int64(p.total) {
		rate := float64(c) / elapsedSecs
		remaining := float64(int64(p.total)-c) / rate
		eta := time.Duration(remaining * float64(time.Second))
		etaStr = eta.Truncate(time.Second).String()
	} else {
		etaStr = "0s"
	}

	passRate := float64(p.passed.Load()) / float64(c) * 100
	pct := float64(c) / float64(p.total) * 100

	p.mu.Lock()
	fmt.Printf("  [%d/%d] %.0f%% done | pass: %.0f%% | ETA: %s\n",
		c, p.total, pct, passRate, etaStr)
	p.mu.Unlock()
}

// runCounter provides a monotonically increasing sequence to guarantee run ID
// uniqueness even when multiple runs start within the same millisecond.
//
// @implements P11 (RunID uniqueness enhancement)
var runCounter atomic.Int64

// generateRunID creates a unique run identifier using tag, task ID, run number,
// millisecond timestamp, and an atomic sequence number to prevent collisions
// in parallel execution.
//
// @implements P11 (RunID uniqueness enhancement)
func generateRunID(tag, taskID string, runNum int) string {
	seq := runCounter.Add(1)
	return fmt.Sprintf("%s-%s-run%d-%d-%d", tag, taskID, runNum, time.Now().UnixMilli(), seq)
}

// RunConfig holds parameters for a benchmark run.
type RunConfig struct {
	Tag               string
	Workflow          string
	TaskSelector      string
	RunsPerTask       int
	Parallel          int    // Number of tasks to run in parallel (1 = serial)
	PlanOverride      string // If set, use this plan instead of task's plan.md
	TimeoutMultiplier int
	TasksDir          string
	DBPath            string
	HomeDir           string         // Base directory for raw output storage
	WorkflowCfg       map[string]any // Per-workflow config passed to adapter constructor

	// Judge configuration (LLM-as-Judge code quality scoring).
	JudgeEnabled bool
	JudgeModel   string
	JudgeAPIKey  string

	// Cost model pricing (from config).
	InputPricePerMTok  float64
	OutputPricePerMTok float64

	// P4: Debug experience.
	KeepWorktree bool // If true, don't delete worktrees after runs.

	// P22: Task sharding for multi-machine execution.
	// ShardIndex is 1-indexed (1..ShardTotal). 0/0 means no sharding.
	ShardIndex int
	ShardTotal int
}

// Execute runs the full benchmark pipeline.
// For each task matching the selector, it runs the adapter the specified number of times,
// verifies results, collects metrics, and stores results in the database.
//
// @implements REQ-RUNNER (orchestrate: load tasks -> adapter.Run -> verify -> collect -> store)
func Execute(ctx context.Context, cfg RunConfig) error {
	// P12: Clean up orphaned worktrees from previous crashed runs.
	if !cfg.KeepWorktree {
		CleanupOrphanedWorktrees()
	}

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

	// P22: Apply sharding if configured.
	if cfg.ShardTotal > 0 {
		filtered = ShardTasks(filtered, cfg.ShardIndex, cfg.ShardTotal)
		if len(filtered) == 0 {
			fmt.Printf("Shard %d/%d: no tasks assigned to this shard\n", cfg.ShardIndex, cfg.ShardTotal)
			return nil
		}
		fmt.Printf("Shard %d/%d: %d tasks assigned\n", cfg.ShardIndex, cfg.ShardTotal, len(filtered))
	}

	// Get adapter.
	// Resolve adapter type from WorkflowCfg if present (e.g., custom workflow
	// names like "my-workflow" map to an adapter type like "custom" via the
	// "adapter" key in WorkflowCfg). Fall back to the workflow name itself
	// for built-in adapters like "vanilla".
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

	parallel := cfg.Parallel
	if parallel <= 0 {
		parallel = 1
	}

	totalRuns := len(filtered) * cfg.RunsPerTask

	fmt.Printf("\nworkflow-bench | workflow: %s | tasks: %d | runs/task: %d | total: %d | parallel: %d\n\n",
		cfg.Workflow, len(filtered), cfg.RunsPerTask, totalRuns, parallel)

	if parallel > 1 {
		return executeParallel(ctx, db, adpt, filtered, cfg, totalRuns, parallel)
	}

	return executeSerial(ctx, db, adpt, filtered, cfg, totalRuns)
}

// executeSerial runs tasks one at a time (original behavior).
func executeSerial(ctx context.Context, db *store.DB, adpt adapter.Adapter, filtered []*config.TaskMeta, cfg RunConfig, totalRuns int) error {
	runIdx := 0
	passCount := 0
	skippedCount := 0
	pt := newProgressTracker(totalRuns)

	for _, task := range filtered {
		for runNum := 1; runNum <= cfg.RunsPerTask; runNum++ {
			runIdx++
			runID := generateRunID(cfg.Tag, task.ID, runNum)

			// Check for existing completed run (checkpoint/resume).
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
			if err := db.DeleteIncompleteRun(cfg.Tag, cfg.Workflow, task.ID, runNum); err != nil {
				slog.Error("delete incomplete run", "task", task.ID, "run", runNum, "error", err)
			}

			fmt.Printf("[%d/%d] %s run#%d ...\n", runIdx, totalRuns, task.ID, runNum)

			passed, err := executeOneRun(ctx, db, adpt, task, runID, cfg, runNum, nil)
			if err != nil {
				slog.Error("run failed", "task", task.ID, "run", runNum, "error", err)
				pt.report(task.ID, runNum, false)
				continue
			}
			if passed {
				passCount++
			}
			pt.report(task.ID, runNum, passed)
		}
	}

	printSummary(cfg.Tag, totalRuns, passCount, skippedCount)
	return nil
}

// executeParallel runs tasks with a semaphore controlling concurrency.
func executeParallel(ctx context.Context, db *store.DB, adpt adapter.Adapter, filtered []*config.TaskMeta, cfg RunConfig, totalRuns, parallel int) error {
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup
	var mu sync.Mutex // protects fmt.Printf output
	var passCount int64
	var skippedCount int64
	var runIdx int64
	pt := newProgressTracker(totalRuns)

	for _, task := range filtered {
		for runNum := 1; runNum <= cfg.RunsPerTask; runNum++ {
			// Check for existing completed run before spawning goroutine.
			exists, dbErr := db.RunExists(cfg.Tag, cfg.Workflow, task.ID, runNum)
			if dbErr != nil {
				slog.Error("check run exists", "task", task.ID, "run", runNum, "error", dbErr)
			}
			if exists {
				slog.Info("skipping completed run", "task", task.ID, "run", runNum)
				atomic.AddInt64(&skippedCount, 1)
				atomic.AddInt64(&runIdx, 1)
				continue
			}
			// Delete any incomplete previous attempt.
			if err := db.DeleteIncompleteRun(cfg.Tag, cfg.Workflow, task.ID, runNum); err != nil {
				slog.Error("delete incomplete run", "task", task.ID, "run", runNum, "error", err)
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(t *config.TaskMeta, rn int) {
				defer wg.Done()
				defer func() { <-sem }()

				idx := atomic.AddInt64(&runIdx, 1)
				runID := generateRunID(cfg.Tag, t.ID, rn)

				mu.Lock()
				fmt.Printf("[%d/%d] %s run#%d ...\n", idx, totalRuns, t.ID, rn)
				mu.Unlock()

				passed, err := executeOneRun(ctx, db, adpt, t, runID, cfg, rn, &mu)
				if err != nil {
					slog.Error("run failed", "task", t.ID, "run", rn, "error", err)
					pt.report(t.ID, rn, false)
					return
				}
				if passed {
					atomic.AddInt64(&passCount, 1)
				}
				pt.report(t.ID, rn, passed)
			}(task, runNum)
		}
	}

	wg.Wait()

	printSummary(cfg.Tag, totalRuns, int(passCount), int(skippedCount))
	return nil
}

// printSummary prints the final summary line after all runs are complete.
func printSummary(tag string, totalRuns, passCount, skippedCount int) {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	executedRuns := totalRuns - skippedCount
	if skippedCount > 0 {
		fmt.Printf("Summary: %d/%d passed (%d skipped)\n", passCount, executedRuns, skippedCount)
	} else {
		fmt.Printf("Summary: %d/%d passed\n", passCount, totalRuns)
	}
	fmt.Printf("Results stored: tag=%s | %d runs in results.db\n", tag, executedRuns)
	fmt.Printf("Run `workflow-bench report --tag %s` for full report.\n", tag)
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

// prepareRun creates the initial Run record in the database and loads the plan content.
// Returns the Run, the plan content, and any error.
//
// @implements P16 (runner.go simplification — extracted from executeOneRun)
func prepareRun(db *store.DB, task *config.TaskMeta, runID string, cfg RunConfig, runNum int) (*store.Run, string, error) {
	now := time.Now()

	run := &store.Run{
		ID:        runID,
		Tag:       cfg.Tag,
		Workflow:  cfg.Workflow,
		TaskID:    task.ID,
		Tier:      task.Tier,
		TaskType:  task.Type,
		RunNumber: runNum,
		Status:    "running",
		StartedAt: now,
	}

	if err := db.InsertRun(run); err != nil {
		return nil, "", fmt.Errorf("insert run: %w", err)
	}

	planContent := cfg.PlanOverride
	if planContent == "" {
		var err error
		planContent, err = config.LoadPlan(task.Dir)
		if err != nil {
			markRunFailed(db, run, "failed")
			return nil, "", fmt.Errorf("load plan: %w", err)
		}
	}
	run.PlanContent = planContent

	return run, planContent, nil
}

// recordTokens records token usage, cost, and efficiency data from the adapter output
// onto the Run record.
//
// @implements P16 (runner.go simplification — extracted from executeOneRun)
func recordTokens(run *store.Run, output *adapter.RunOutput, cfg RunConfig, tier int) {
	if output.TokenUsage != nil {
		run.InputTokens = &output.TokenUsage.InputTokens
		run.OutputTokens = &output.TokenUsage.OutputTokens
		total := output.TokenUsage.InputTokens + output.TokenUsage.OutputTokens
		run.TotalTokens = &total

		cost := metrics.EstimateCost(output.TokenUsage.InputTokens, output.TokenUsage.OutputTokens, cfg.InputPricePerMTok, cfg.OutputPricePerMTok)
		run.CostUSD = &cost

		// Efficiency score: 1.0 - min(1.0, cost/budget)
		budget := tierCostBudget(tier)
		eff := 1.0 - math.Min(1.0, cost/budget)
		run.EfficiencyScore = &eff
	}
	wallSecs := output.WallTime.Seconds()
	run.WallTimeSecs = &wallSecs
	if output.ToolUses > 0 {
		run.ToolUses = &output.ToolUses
	}
}

// computeScores calculates correctness, runs the judge if enabled, and computes
// the composite final score. It populates the corresponding fields on the Run.
//
// @implements P16 (runner.go simplification — extracted from executeOneRun)
func computeScores(ctx context.Context, run *store.Run, verifyResult *VerifyResult, task *config.TaskMeta, cfg RunConfig, worktreeDir, repoDir, planContent string) float64 {
	// Populate run with verify results.
	run.L1Build = &verifyResult.L1Build
	run.L2UtPassed = &verifyResult.L2Passed
	run.L2UtTotal = &verifyResult.L2Total
	run.L3LintIssues = &verifyResult.L3Issues
	run.L4E2EPassed = &verifyResult.L4Passed
	run.L4E2ETotal = &verifyResult.L4Total

	// P2: Detect critical VT failures based on L1/L4 results and task VTs.
	criticalVTFails := countCriticalVTFailures(task.VerificationTargets, verifyResult)

	// Calculate correctness score.
	correctness := metrics.CalculateCorrectness(metrics.CorrectnessInput{
		L1Build:             verifyResult.L1Build,
		L2Passed:            verifyResult.L2Passed,
		L2Total:             verifyResult.L2Total,
		L3Issues:            verifyResult.L3Issues,
		L4Passed:            verifyResult.L4Passed,
		L4Total:             verifyResult.L4Total,
		CriticalVTFailCount: criticalVTFails,
	})
	run.CorrectnessScore = &correctness

	// LLM Judge: score code quality if enabled.
	if cfg.JudgeEnabled && cfg.JudgeAPIKey != "" {
		runJudge(ctx, run, cfg, worktreeDir, repoDir, planContent)
	}

	// P6: Four-dimension composite score with proportional weight redistribution.
	// Design weights: correctness=0.40, efficiency=0.25, quality=0.25, stability=0.10
	// When a dimension is unavailable, its weight is redistributed proportionally.
	var rubricComposite *float64
	if run.RubricComposite != nil {
		// Normalize rubric from 0-5 scale to 0-1.
		v := *run.RubricComposite / 5.0
		rubricComposite = &v
	}
	final := ComputeCompositeScore(correctness, run.EfficiencyScore, rubricComposite, nil)
	run.FinalScore = &final

	return correctness
}

// printRunResult formats and prints the run result to stdout.
// If outputMu is non-nil, output is locked to prevent interleaved lines in parallel mode.
//
// @implements P16 (runner.go simplification — extracted from executeOneRun)
func printRunResult(run *store.Run, verifyResult *VerifyResult, output *adapter.RunOutput, correctness float64, outputMu *sync.Mutex) {
	l2Str := fmt.Sprintf("%d/%d", verifyResult.L2Passed, verifyResult.L2Total)
	l4Str := fmt.Sprintf("%d/%d", verifyResult.L4Passed, verifyResult.L4Total)
	l1Str := "PASS"
	if !verifyResult.L1Build {
		l1Str = "FAIL"
	}
	if outputMu != nil {
		outputMu.Lock()
	}
	fmt.Printf("  ├─ adapter: %s | exit=%d\n", output.WallTime.Truncate(time.Second), output.ExitCode)
	fmt.Printf("  ├─ verify:  L1=%s L2=%s L3=%d L4=%s\n", l1Str, l2Str, verifyResult.L3Issues, l4Str)
	if run.RubricComposite != nil {
		fmt.Printf("  ├─ judge:   composite=%.2f/5\n", *run.RubricComposite)
	}
	fmt.Printf("  └─ result:  %s (correctness=%.2f)\n\n", run.Status, correctness)
	if outputMu != nil {
		outputMu.Unlock()
	}
}

// executeOneRun orchestrates a single benchmark run: prepare -> adapter -> verify -> score -> store.
//
// @implements REQ-RUNNER (single run orchestration)
func executeOneRun(ctx context.Context, db *store.DB, adpt adapter.Adapter, task *config.TaskMeta, runID string, cfg RunConfig, runNum int, outputMu *sync.Mutex) (bool, error) {
	// Phase 1: Prepare run record and load plan.
	run, planContent, err := prepareRun(db, task, runID, cfg, runNum)
	if err != nil {
		return false, err
	}

	// Phase 2: Create isolated worktree.
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
	if cfg.KeepWorktree {
		slog.Info("keeping worktree", "path", worktreeDir)
	} else {
		defer CleanupWorktree(repoDir, worktreeDir)
	}

	// Phase 3: Run adapter with timeout.
	timeout := time.Duration(task.EstimatedMinutes) * time.Duration(cfg.TimeoutMultiplier) * time.Minute
	if timeout == 0 {
		timeout = 15 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := adpt.Setup(runCtx, worktreeDir); err != nil {
		markRunFailed(db, run, "failed")
		return false, fmt.Errorf("adapter setup: %w", err)
	}

	output, err := adpt.Run(runCtx, worktreeDir, planContent)
	if err != nil {
		status := "failed"
		if runCtx.Err() == context.DeadlineExceeded {
			status = "timeout"
		}
		markRunFailed(db, run, status)
		return false, fmt.Errorf("adapter run: %w", err)
	}

	// Phase 4: Record token usage and efficiency.
	recordTokens(run, output, cfg, task.Tier)

	// Capture agent's code changes as a diff patch.
	if cfg.HomeDir != "" {
		captureDiff(worktreeDir, cfg.HomeDir, runID)
	}

	// Phase 5: Generate and run verify.
	verifyDir, err := GenerateVerifyDir(VerifyConfig{
		TaskType:     task.Type,
		TaskDir:      task.Dir,
		TaskLanguage: task.Language,
		RunID:        runID,
	})
	if err != nil {
		return false, fmt.Errorf("generate verify: %w", err)
	}
	defer os.RemoveAll(verifyDir)

	verifyOutput, err := RunVerify(verifyDir, worktreeDir)
	if err != nil {
		slog.Warn("verify execution error", "error", err)
	}

	// Store raw verify output for later inspection.
	if cfg.HomeDir != "" {
		rawDir := filepath.Join(cfg.HomeDir, "raw", runID)
		if mkErr := os.MkdirAll(rawDir, 0755); mkErr != nil {
			slog.Warn("create raw output dir", "error", mkErr)
		} else {
			if wErr := os.WriteFile(filepath.Join(rawDir, "verify.log"), []byte(verifyOutput), 0644); wErr != nil {
				slog.Warn("write verify.log", "error", wErr)
			}
		}
	}

	// Parse verify results.
	verifyResult, err := ParseVerifyOutput(verifyOutput)
	if err != nil {
		slog.Warn("parse verify output failed", "error", err, "output", verifyOutput)
		markRunFailed(db, run, "failed")
		return false, nil
	}

	// Phase 6: Compute scores and finalize.
	correctness := computeScores(ctx, run, verifyResult, task, cfg, worktreeDir, repoDir, planContent)

	run.Status = "completed"
	finished := time.Now()
	run.FinishedAt = &finished

	if err := db.UpdateRun(run); err != nil {
		return false, fmt.Errorf("update run: %w", err)
	}

	printRunResult(run, verifyResult, output, correctness, outputMu)

	// Fix 3: Match report/summary.go pass definition — L1 build + all L4 E2E tests passing.
	passed := verifyResult.L1Build && verifyResult.L4Total > 0 && verifyResult.L4Passed == verifyResult.L4Total
	return passed, nil
}

// captureDiff captures the agent's code changes as a diff patch and saves it to
// <homeDir>/raw/<runID>/diff.patch. It stages all files (including untracked) to
// ensure the diff captures all changes.
func captureDiff(worktreeDir, homeDir, runID string) {
	// Stage all files to capture untracked files in the diff.
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = worktreeDir
	if err := addCmd.Run(); err != nil {
		slog.Warn("capture diff: git add", "error", err)
		return
	}

	diffCmd := exec.Command("git", "diff", "--cached", "HEAD")
	diffCmd.Dir = worktreeDir
	diffOutput, err := diffCmd.CombinedOutput()
	if err != nil {
		slog.Warn("capture diff: git diff", "error", err)
		return
	}

	if len(diffOutput) == 0 {
		return // No changes to save.
	}

	rawDir := filepath.Join(homeDir, "raw", runID)
	if mkErr := os.MkdirAll(rawDir, 0755); mkErr != nil {
		slog.Warn("capture diff: create raw dir", "error", mkErr)
		return
	}
	if wErr := os.WriteFile(filepath.Join(rawDir, "diff.patch"), diffOutput, 0644); wErr != nil {
		slog.Warn("capture diff: write diff.patch", "error", wErr)
	}
}

// runJudge performs LLM-as-Judge scoring on the code changes in a worktree.
// It gets the git diff, reads original code, calls the judge, and populates
// rubric fields on the Run. Errors are logged but do not fail the run.
//
// P10: Uses an independent 60-second timeout for the judge API call to prevent
// slow judge responses from blocking the entire run pipeline.
func runJudge(ctx context.Context, run *store.Run, cfg RunConfig, worktreeDir, repoDir, planContent string) {
	// Get git diff of changes in worktree vs original.
	diff, err := getWorktreeDiff(worktreeDir)
	if err != nil {
		slog.Warn("judge: get worktree diff", "error", err)
		return
	}
	if diff == "" {
		slog.Info("judge: no code changes detected, skipping scoring")
		return
	}

	// Read original code listing from repo dir for context.
	originalCode := getOriginalCode(repoDir)

	// P10: Independent timeout for judge API call (60 seconds).
	// Uses context.Background() so the judge timeout is independent from the
	// adapter run timeout (which may already be nearly expired).
	judgeCtx, judgeCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer judgeCancel()

	// Call the LLM judge.
	score, err := judge.ScoreRun(judgeCtx, cfg.JudgeAPIKey, cfg.JudgeModel, planContent, originalCode, diff)
	if err != nil {
		slog.Warn("judge: score run", "error", err)
		return
	}

	// Populate rubric fields on the run.
	if dim, ok := score.Dimensions["correctness"]; ok {
		v := float64(dim.Score)
		run.RubricCorrectness = &v
	}
	if dim, ok := score.Dimensions["readability"]; ok {
		v := float64(dim.Score)
		run.RubricReadability = &v
	}
	if dim, ok := score.Dimensions["simplicity"]; ok {
		v := float64(dim.Score)
		run.RubricSimplicity = &v
	}
	if dim, ok := score.Dimensions["robustness"]; ok {
		v := float64(dim.Score)
		run.RubricRobustness = &v
	}
	if dim, ok := score.Dimensions["minimality"]; ok {
		v := float64(dim.Score)
		run.RubricMinimality = &v
	}
	if dim, ok := score.Dimensions["maintainability"]; ok {
		v := float64(dim.Score)
		run.RubricMaintainability = &v
	}
	if score.GoIdioms != nil {
		v := float64(score.GoIdioms.Score)
		run.RubricGoIdioms = &v
	}
	run.RubricComposite = &score.Composite

	// P13: Store boolean-to-score consistency warnings if any were detected.
	if len(score.ConsistencyWarnings) > 0 {
		warnings := strings.Join(score.ConsistencyWarnings, "; ")
		run.RubricConsistencyWarnings = &warnings
	}
}

// ComputeCompositeScore calculates the four-dimension composite score.
// Design weights: correctness=0.40, efficiency=0.25, quality=0.25, stability=0.10.
// When a dimension is unavailable (nil), its weight is redistributed proportionally
// to the available dimensions.
//
// @implements P6 (composite formula consistency fix)
func ComputeCompositeScore(correctness float64, efficiency, quality, stability *float64) float64 {
	type dim struct {
		weight float64
		score  *float64
	}
	// Correctness is always available (passed directly as a float64).
	corrScore := correctness
	dims := []dim{
		{0.40, &corrScore},
		{0.25, efficiency},
		{0.25, quality},
		{0.10, stability},
	}

	// Sum available weights.
	var totalAvailWeight float64
	for _, d := range dims {
		if d.score != nil {
			totalAvailWeight += d.weight
		}
	}
	if totalAvailWeight == 0 {
		return 0
	}

	// Compute weighted sum with redistributed weights.
	var result float64
	for _, d := range dims {
		if d.score != nil {
			result += (d.weight / totalAvailWeight) * (*d.score)
		}
	}
	return result
}

// countCriticalVTFailures counts how many critical-severity VTs failed based on
// verify results. Uses expanded detection type mapping:
//   - "go build", "build" -> L1=FAIL means this VT failed
//   - "e2e test case", "e2e test", "e2e" -> L4 has any failure means this VT failed
//   - "unit test", "ut", "go test" -> L2 has any failure means this VT failed
//   - "go vet", "staticcheck", "lint", "errcheck", "bodyclose" -> L3 issues > 0
//   - "race detector", "go test -race", "goleak" -> L2 has failures (race is embedded in L2)
//
// Only VTs with severity "critical" are counted.
// Unknown detection types are logged as warnings and not counted.
//
// @implements P2, P9 (VT detection for critical VT fail count with expanded mapping)
func countCriticalVTFailures(vts []config.VerificationTarget, result *VerifyResult) int {
	count := 0
	for _, vt := range vts {
		if !strings.EqualFold(vt.Severity, "critical") {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(vt.Detection)) {
		case "go build", "build":
			if !result.L1Build {
				count++
			}
		case "e2e test case", "e2e test", "e2e":
			if result.L4Total > 0 && result.L4Passed < result.L4Total {
				count++
			}
		case "unit test", "ut", "go test":
			if result.L2Total > 0 && result.L2Passed < result.L2Total {
				count++
			}
		case "go vet", "staticcheck", "lint", "errcheck", "bodyclose":
			if result.L3Issues > 0 {
				count++
			}
		case "race detector", "go test -race", "goleak":
			// Race detection results are embedded in L2 (go test -race).
			// If L2 has failures, it may be due to race conditions.
			if result.L2Total > 0 && result.L2Passed < result.L2Total {
				count++
			}
		default:
			slog.Warn("unknown VT detection type", "type", vt.Detection, "vt_id", vt.ID)
		}
	}
	return count
}

// tierCostBudget returns the expected cost budget in USD for a given task tier.
// Used to calculate the efficiency score as 1.0 - min(1.0, cost/budget).
func tierCostBudget(tier int) float64 {
	switch tier {
	case 1:
		return 0.50
	case 2:
		return 1.00
	case 3:
		return 2.00
	case 4:
		return 5.00
	default:
		return 2.00
	}
}


// getWorktreeDiff returns the git diff of uncommitted + committed changes in a worktree
// compared to the worktree's base commit (HEAD).
func getWorktreeDiff(worktreeDir string) (string, error) {
	// Get diff of all changes (staged + unstaged + untracked) vs HEAD.
	// First, add all files to index to capture untracked files in the diff.
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = worktreeDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %w\noutput: %s", err, string(out))
	}

	diffCmd := exec.Command("git", "diff", "--cached", "HEAD")
	diffCmd.Dir = worktreeDir
	out, err := diffCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff: %w\noutput: %s", err, string(out))
	}
	return string(out), nil
}

// getOriginalCode retrieves all Go file contents from HEAD in the given repo directory.
// It uses git ls-tree to list files and git show to read each one, concatenating them
// into a single string for use as judge context.
func getOriginalCode(repoDir string) string {
	listCmd := exec.Command("git", "ls-tree", "-r", "--name-only", "HEAD")
	listCmd.Dir = repoDir
	output, err := listCmd.Output()
	if err != nil {
		return ""
	}

	var buf strings.Builder
	for _, file := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if !strings.HasSuffix(file, ".go") {
			continue
		}
		showCmd := exec.Command("git", "show", "HEAD:"+file)
		showCmd.Dir = repoDir
		content, err := showCmd.Output()
		if err != nil {
			continue
		}
		fmt.Fprintf(&buf, "// === %s ===\n%s\n\n", file, string(content))
	}
	return buf.String()
}

// ShardTasks selects a subset of tasks for the given shard.
// index is 1-indexed (1..total). Tasks are assigned round-robin by position.
//
// @implements P22 (Task sharding for multi-machine execution)
func ShardTasks(tasks []*config.TaskMeta, index, total int) []*config.TaskMeta {
	if total <= 0 || index <= 0 || index > total {
		return tasks
	}
	var result []*config.TaskMeta
	for i, t := range tasks {
		if i%total == index-1 {
			result = append(result, t)
		}
	}
	return result
}
