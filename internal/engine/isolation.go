// Package engine provides the benchmark execution engine.
package engine

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// CreateWorktree creates an isolated git worktree from a task repo.
// It returns the path to the new worktree directory.
//
// @implements REQ-ISOLATION (git worktree isolation for each run)
func CreateWorktree(repoDir string, runID string) (string, error) {
	worktreeDir := filepath.Join(os.TempDir(), "bench-worktree-"+runID)

	// Ensure the source repo is a git repo.
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		return "", fmt.Errorf("repo dir %s is not a git repository", repoDir)
	}

	// Create worktree.
	cmd := exec.Command("git", "worktree", "add", "--detach", worktreeDir)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree add: %w\noutput: %s", err, string(out))
	}

	return worktreeDir, nil
}

// CleanupWorktree removes a git worktree and its directory.
func CleanupWorktree(repoDir string, worktreeDir string) error {
	// Remove from git worktree list.
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreeDir)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		// Try manual cleanup if git worktree remove fails.
		os.RemoveAll(worktreeDir)
		// Also try to prune.
		pruneCmd := exec.Command("git", "worktree", "prune")
		pruneCmd.Dir = repoDir
		pruneCmd.Run()
		return fmt.Errorf("git worktree remove: %w\noutput: %s", err, string(out))
	}
	return nil
}

// CleanupOrphanedWorktrees removes stale worktree and verify temp directories
// that are older than 2 hours. These can accumulate when runs crash or are killed
// without proper cleanup.
//
// Patterns cleaned:
//   - bench-worktree-* (orphaned git worktrees)
//   - bench-verify-* (orphaned verify temp dirs)
//
// @implements P12 (worktree orphan cleanup)
func CleanupOrphanedWorktrees() {
	cutoff := time.Now().Add(-2 * time.Hour)

	// Clean up orphaned worktrees.
	worktreePattern := filepath.Join(os.TempDir(), "bench-worktree-*")
	cleanupOldDirs(worktreePattern, cutoff)

	// Clean up orphaned verify temp dirs.
	verifyPattern := filepath.Join(os.TempDir(), "bench-verify-*")
	cleanupOldDirs(verifyPattern, cutoff)
}

// cleanupOldDirs removes directories matching a glob pattern that are older than cutoff.
func cleanupOldDirs(pattern string, cutoff time.Time) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, dir := range matches {
		info, err := os.Stat(dir)
		if err == nil && info.ModTime().Before(cutoff) {
			slog.Info("cleaning up orphaned temp dir", "dir", dir)
			os.RemoveAll(dir)
		}
	}
}

// ensureGitMu serializes EnsureGitRepo calls to prevent TOCTOU races
// when multiple goroutines initialize the same repo concurrently.
var ensureGitMu sync.Mutex

// EnsureGitRepo initializes a git repo in dir if one doesn't exist.
// Used to ensure task repo/ directories are git repos for worktree operations.
func EnsureGitRepo(dir string) error {
	ensureGitMu.Lock()
	defer ensureGitMu.Unlock()

	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil // Already a git repo.
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init in %s: %w\noutput: %s", dir, err, string(out))
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add in %s: %w\noutput: %s", dir, err, string(out))
	}

	cmd = exec.Command("git", "commit", "-m", "initial", "--allow-empty")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=workflow-bench",
		"GIT_AUTHOR_EMAIL=noreply@workflow-bench",
		"GIT_COMMITTER_NAME=workflow-bench",
		"GIT_COMMITTER_EMAIL=noreply@workflow-bench",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit in %s: %w\noutput: %s", dir, err, string(out))
	}

	return nil
}
