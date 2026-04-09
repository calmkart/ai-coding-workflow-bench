package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCleanupOrphanedWorktrees_RemovesOld verifies that orphaned directories
// older than 2 hours are removed.
func TestCleanupOrphanedWorktrees_RemovesOld(t *testing.T) {
	// Create a fake old worktree dir in TempDir.
	oldDir := filepath.Join(os.TempDir(), "bench-worktree-test-orphan-old")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(oldDir) // Safety cleanup.

	// Set the mtime to 3 hours ago.
	oldTime := time.Now().Add(-3 * time.Hour)
	os.Chtimes(oldDir, oldTime, oldTime)

	CleanupOrphanedWorktrees()

	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("expected old worktree to be cleaned up")
	}
}

// TestCleanupOrphanedWorktrees_KeepsRecent verifies that recent directories
// are NOT removed.
func TestCleanupOrphanedWorktrees_KeepsRecent(t *testing.T) {
	// Create a recent worktree dir.
	recentDir := filepath.Join(os.TempDir(), "bench-worktree-test-orphan-recent")
	if err := os.MkdirAll(recentDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(recentDir)

	// mtime is now (recent), so it should NOT be cleaned up.
	CleanupOrphanedWorktrees()

	if _, err := os.Stat(recentDir); os.IsNotExist(err) {
		t.Error("expected recent worktree to be kept")
	}
}

// TestCleanupOrphanedWorktrees_RemovesOldVerifyDirs verifies that old
// bench-verify-* directories are also cleaned up.
func TestCleanupOrphanedWorktrees_RemovesOldVerifyDirs(t *testing.T) {
	oldDir := filepath.Join(os.TempDir(), "bench-verify-test-orphan-old")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(oldDir)

	oldTime := time.Now().Add(-3 * time.Hour)
	os.Chtimes(oldDir, oldTime, oldTime)

	CleanupOrphanedWorktrees()

	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("expected old verify dir to be cleaned up")
	}
}

// TestCleanupOldDirs_NoMatch verifies that cleanupOldDirs handles
// a pattern with no matches gracefully.
func TestCleanupOldDirs_NoMatch(t *testing.T) {
	// Should not panic or error.
	cleanupOldDirs(filepath.Join(os.TempDir(), "bench-nonexistent-pattern-*"), time.Now())
}
