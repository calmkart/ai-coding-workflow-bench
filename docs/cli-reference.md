# CLI Reference

## Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `~/.claude/workflow-bench/bench.yaml` | Path to config file |
| `-v, --verbose` | bool | false | Enable verbose (debug) logging |

---

## workflow-bench run

**Description**: Run a benchmark against selected tasks using a workflow adapter.

**Usage**: `workflow-bench run --tasks <selector> --tag <tag> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--workflow` | string | `vanilla` | Workflow adapter: `vanilla`, `v4-claude`, or custom name |
| `--tasks` | string | *(required)* | Task selector: `tier1`, `tier1/fix-handler-bug`, or `all` |
| `--tag` | string | *(required)* | Tag to label this benchmark run |
| `--runs` | int | from config (default 3) | Number of runs per task |
| `--plan` | string | | Path to plan file override (replaces task's plan.md) |
| `--tasks-dir` | string | `tasks` | Path to tasks directory |

**Examples**:

```bash
# Run all tier 1 tasks with vanilla workflow
workflow-bench run --workflow vanilla --tasks tier1 --runs 3 --tag baseline

# Run with v4-claude multi-agent workflow
workflow-bench run --workflow v4-claude --tasks tier1 --runs 3 --tag v4-test

# Run with a custom workflow (defined in bench.yaml)
workflow-bench run --workflow my-workflow --tasks tier1 --runs 1 --tag custom-test

# Run a single task
workflow-bench run --tasks tier1/fix-handler-bug --runs 1 --tag quick-test

# Use a custom plan
workflow-bench run --tasks tier1/fix-handler-bug --plan ./my-plan.md --tag custom-plan
```

**Notes**:
- Each run creates an isolated git worktree. The original task repo is never modified.
- If a run with the same `(tag, workflow, task_id, run_number)` already completed, it is skipped (checkpoint/resume).
- Timeout per task = `estimated_minutes * timeout_multiplier` (default 3x).
- The `--plan` flag overrides the plan for the current run only; it does not modify files on disk.

---

## workflow-bench report

**Description**: Generate a Markdown summary report for a tagged benchmark run.

**Usage**: `workflow-bench report --tag <tag> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | *(required)* | Tag to generate report for |
| `--format` | string | `markdown` | Output format (`markdown` or `md`) |

**Examples**:

```bash
# Print report to stdout
workflow-bench report --tag baseline

# Save report to file
workflow-bench report --tag baseline > report.md
```

**Notes**:
- Output includes overall pass rate, average correctness, and per-task L1-L4 breakdowns.
- When multiple runs per task exist, results are grouped with aggregated pass counts.

---

## workflow-bench list tasks

**Description**: List all available tasks discovered from the tasks directory.

**Usage**: `workflow-bench list tasks [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tasks-dir` | string | `tasks` | Path to tasks directory |

**Examples**:

```bash
workflow-bench list tasks
```

**Output**:

```
ID                             TIER  TYPE            EST
tier1/fix-handler-bug          T1    http-server     5m
tier1/add-health-check         T1    http-server     5m

2 tasks found
```

---

## workflow-bench list workflows

**Description**: List available workflow adapters.

**Usage**: `workflow-bench list workflows`

**Examples**:

```bash
workflow-bench list workflows
```

**Output**:

```
Available workflows:
  vanilla    - Claude CLI direct execution
  v4-claude  - Claude CLI multi-agent (--agent manager)
  custom     - User-defined command execution
```

---

## workflow-bench list tags

**Description**: List all benchmark tags with run counts, dates, and workflows.

**Usage**: `workflow-bench list tags`

**Examples**:

```bash
workflow-bench list tags
```

**Output**:

```
TAG                       RUNS   DATE                 WORKFLOW
baseline                  6      2026-03-30 14:00     vanilla
v4-test                   3      2026-03-31 10:30     vanilla
```

---

## workflow-bench validate

**Description**: Validate task definitions for correctness. Checks task.yaml fields, repo structure, build, plan.md, and E2E test files.

**Usage**: `workflow-bench validate [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tasks` | string | `all` | Task selector to validate |
| `--tasks-dir` | string | `tasks` | Path to tasks directory |

**Validation checks**:
- `id` is non-empty
- `tier` is 1-4
- `type` is non-empty
- `repo/` directory exists
- `repo/go.mod` exists
- `go build ./...` passes in repo/ (30s timeout)
- `plan.md` exists and is non-empty
- `verify/e2e_test.go` or `verify/e2e_test.go.src` exists
- `estimated_minutes` > 0

**Examples**:

```bash
# Validate all tasks
workflow-bench validate

# Validate only tier 1 tasks with verbose output
workflow-bench validate --tasks tier1 -v
```

**Notes**:
- With `-v` (verbose), each check prints OK/FAIL with details.
- Exit code is non-zero if any task fails validation.

---

## workflow-bench init

**Description**: Initialize the configuration directory at `~/.claude/workflow-bench/` and create a default `bench.yaml` if one does not exist.

**Usage**: `workflow-bench init`

**Examples**:

```bash
workflow-bench init
```

**Output**:

```
Initialized: /Users/you/.claude/workflow-bench
Config: /Users/you/.claude/workflow-bench/bench.yaml
```

**Notes**:
- Safe to run multiple times. If `bench.yaml` already exists, it is not overwritten.

---

## workflow-bench version

**Description**: Print the version string.

**Usage**: `workflow-bench version`

**Examples**:

```bash
workflow-bench version
```

**Output**:

```
workflow-bench 0.1.0
```
