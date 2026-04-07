package adapter

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCustom_MissingEntryCommand(t *testing.T) {
	_, err := NewCustom(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !strings.Contains(err.Error(), "entry_command") {
		t.Errorf("expected error to mention entry_command, got: %v", err)
	}
}

func TestNewCustom_EmptyEntryCommand(t *testing.T) {
	cfg := map[string]any{
		"entry_command": "",
	}
	_, err := NewCustom(cfg)
	if err == nil {
		t.Fatal("expected error for empty entry_command")
	}
	if !strings.Contains(err.Error(), "entry_command") {
		t.Errorf("expected error to mention entry_command, got: %v", err)
	}
}

func TestNewCustom_ValidConfig(t *testing.T) {
	cfg := map[string]any{
		"entry_command":  "echo hello",
		"setup_commands": []string{"mkdir -p .planning"},
	}
	a, err := NewCustom(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ca := a.(*CustomAdapter)
	if ca.entryCommand != "echo hello" {
		t.Errorf("expected entry_command 'echo hello', got %q", ca.entryCommand)
	}
	if len(ca.setupCommands) != 1 || ca.setupCommands[0] != "mkdir -p .planning" {
		t.Errorf("expected setup_commands [mkdir -p .planning], got %v", ca.setupCommands)
	}
}

func TestNewCustom_SetupCommandsAsAnySlice(t *testing.T) {
	// YAML unmarshaling often produces []any instead of []string.
	cfg := map[string]any{
		"entry_command":  "echo hello",
		"setup_commands": []any{"cmd1", "cmd2"},
	}
	a, err := NewCustom(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ca := a.(*CustomAdapter)
	if len(ca.setupCommands) != 2 {
		t.Fatalf("expected 2 setup commands, got %d", len(ca.setupCommands))
	}
	if ca.setupCommands[0] != "cmd1" || ca.setupCommands[1] != "cmd2" {
		t.Errorf("unexpected setup commands: %v", ca.setupCommands)
	}
}

func TestNewCustom_NoSetupCommands(t *testing.T) {
	cfg := map[string]any{
		"entry_command": "echo hello",
	}
	a, err := NewCustom(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ca := a.(*CustomAdapter)
	if len(ca.setupCommands) != 0 {
		t.Errorf("expected no setup commands, got %v", ca.setupCommands)
	}
}

func TestNewCustom_CustomName(t *testing.T) {
	cfg := map[string]any{
		"entry_command": "echo hello",
		"name":          "my-workflow",
	}
	a, err := NewCustom(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Name() != "my-workflow" {
		t.Errorf("expected name 'my-workflow', got %q", a.Name())
	}
}

func TestCustomAdapter_Name(t *testing.T) {
	a := &CustomAdapter{name: "custom", entryCommand: "echo test"}
	if a.Name() != "custom" {
		t.Errorf("expected name 'custom', got %q", a.Name())
	}
}

func TestCustomAdapter_Setup_RunsCommands(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: "echo hello",
		setupCommands: []string{
			"mkdir -p subdir1",
			"touch subdir1/file.txt",
		},
	}

	if err := a.Setup(context.Background(), dir); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Verify the commands executed in the worktree dir.
	filePath := filepath.Join(dir, "subdir1", "file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected %s to exist after setup commands", filePath)
	}
}

func TestCustomAdapter_Setup_NoCommands(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:         "test",
		entryCommand:  "echo hello",
		setupCommands: nil,
	}

	if err := a.Setup(context.Background(), dir); err != nil {
		t.Fatalf("Setup should succeed with no commands: %v", err)
	}
}

func TestCustomAdapter_Setup_FailingCommand(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: "echo hello",
		setupCommands: []string{
			"false", // always fails
		},
	}

	err := a.Setup(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for failing setup command")
	}
	if !strings.Contains(err.Error(), "setup command") {
		t.Errorf("expected error to mention 'setup command', got: %v", err)
	}
}

func TestCustomAdapter_Setup_StopsOnFirstFailure(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: "echo hello",
		setupCommands: []string{
			"false",                  // fails
			"touch should-not-exist", // should not run
		},
	}

	err := a.Setup(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error for failing setup command")
	}
	// Verify second command did NOT run.
	notExist := filepath.Join(dir, "should-not-exist")
	if _, err := os.Stat(notExist); !os.IsNotExist(err) {
		t.Error("second setup command should not have executed after first failure")
	}
}

func TestCustomAdapter_Run_EnvironmentVariables(t *testing.T) {
	dir := t.TempDir()

	// entry_command writes env vars to a file so we can inspect them.
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `env | grep ^BENCH_ > env_output.txt`,
	}

	_, err := a.Run(context.Background(), dir, "test plan content")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "env_output.txt"))
	if err != nil {
		t.Fatalf("read env output: %v", err)
	}
	envStr := string(data)

	// Verify BENCH_REPO_DIR is set to the worktree directory.
	if !strings.Contains(envStr, "BENCH_REPO_DIR="+dir) {
		t.Errorf("expected BENCH_REPO_DIR=%s in env, got:\n%s", dir, envStr)
	}

	// Verify BENCH_PLAN_FILE is set.
	planFile := filepath.Join(dir, ".bench-plan.md")
	if !strings.Contains(envStr, "BENCH_PLAN_FILE="+planFile) {
		t.Errorf("expected BENCH_PLAN_FILE=%s in env, got:\n%s", planFile, envStr)
	}

	// Verify BENCH_PLAN_PROMPT is set.
	if !strings.Contains(envStr, "BENCH_PLAN_PROMPT=") {
		t.Errorf("expected BENCH_PLAN_PROMPT in env, got:\n%s", envStr)
	}
}

func TestCustomAdapter_Run_PlanFileContent(t *testing.T) {
	dir := t.TempDir()

	// entry_command copies the plan file to a known location so we can check content.
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `cp "$BENCH_PLAN_FILE" plan_copy.md`,
	}

	planContent := "# Test Plan\n\nDo the thing."
	_, err := a.Run(context.Background(), dir, planContent)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "plan_copy.md"))
	if err != nil {
		t.Fatalf("read plan copy: %v", err)
	}
	if string(data) != planContent {
		t.Errorf("expected plan content %q, got %q", planContent, string(data))
	}
}

func TestCustomAdapter_Run_CapturesStdout(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo "hello from custom"`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !strings.Contains(result.Stdout, "hello from custom") {
		t.Errorf("expected stdout to contain 'hello from custom', got %q", result.Stdout)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestCustomAdapter_Run_NonZeroExit(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo "some output" && exit 1`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("unexpected infrastructure error: %v", err)
	}
	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}

func TestCustomAdapter_Run_ParsesJSON(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo '{"usage":{"input_tokens":100,"output_tokens":50},"tool_uses":5}'`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.TokenUsage == nil {
		t.Fatal("expected token usage to be parsed")
	}
	if result.TokenUsage.InputTokens != 100 {
		t.Errorf("expected input_tokens=100, got %d", result.TokenUsage.InputTokens)
	}
	if result.TokenUsage.OutputTokens != 50 {
		t.Errorf("expected output_tokens=50, got %d", result.TokenUsage.OutputTokens)
	}
	if result.ToolUses != 5 {
		t.Errorf("expected tool_uses=5, got %d", result.ToolUses)
	}
}

func TestCustomAdapter_Run_NoJSON(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo "not json"`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.TokenUsage != nil {
		t.Errorf("expected nil token usage for non-JSON output, got %+v", result.TokenUsage)
	}
}

func TestCustomAdapter_Run_WorksInWorktreeDir(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `pwd`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	// pwd output should match the worktree directory.
	// On macOS, /var is a symlink to /private/var, so resolve both paths.
	got := strings.TrimSpace(result.Stdout)
	resolvedGot, _ := filepath.EvalSymlinks(got)
	resolvedDir, _ := filepath.EvalSymlinks(dir)
	if resolvedGot != resolvedDir {
		t.Errorf("expected working dir %q, got %q", resolvedDir, resolvedGot)
	}
}

func TestCustomAdapter_Run_CleansPlanFile(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo ok`,
	}

	_, err := a.Run(context.Background(), dir, "plan content")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Plan file should be cleaned up after Run.
	planFile := filepath.Join(dir, ".bench-plan.md")
	if _, err := os.Stat(planFile); !os.IsNotExist(err) {
		t.Error("expected .bench-plan.md to be cleaned up after Run")
	}
}

func TestCustomAdapter_Run_WallTime(t *testing.T) {
	dir := t.TempDir()
	a := &CustomAdapter{
		name:        "test",
		entryCommand: `echo ok`,
	}

	result, err := a.Run(context.Background(), dir, "plan")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.WallTime <= 0 {
		t.Errorf("expected positive wall time, got %v", result.WallTime)
	}
}

func TestRegistryContainsCustom(t *testing.T) {
	if _, ok := Registry["custom"]; !ok {
		t.Error("expected custom in adapter Registry")
	}
}

func TestGetCustom(t *testing.T) {
	a, err := Get("custom", map[string]any{"entry_command": "echo test"})
	if err != nil {
		t.Fatalf("Get custom failed: %v", err)
	}
	if a.Name() != "custom" {
		t.Errorf("expected name 'custom', got %q", a.Name())
	}
}
