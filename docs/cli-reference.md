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
| `--workflow` | string | `vanilla` | Workflow adapter: `vanilla`, `custom`, or user-defined name |
| `--tasks` | string | *(required)* | Task selector: `tier1`, `tier1/fix-handler-bug`, or `all` |
| `--tag` | string | *(required)* | Tag to label this benchmark run |
| `--runs` | int | from config (default 3) | Number of runs per task |
| `--plan` | string | | Path to plan file override (replaces task's plan.md) |
| `--tasks-dir` | string | `tasks` | Path to tasks directory |
| `--parallel` | int | `1` | Number of tasks to run in parallel |
| `--keep-worktree` | bool | false | Don't delete worktrees after runs (for debugging) |
| `--shard` | string | | Shard index/total for distributed execution (e.g. `1/4`) |

**Examples**:

```bash
# Run all tier 1 tasks with vanilla workflow
workflow-bench run --workflow vanilla --tasks tier1 --runs 3 --tag baseline

# Run with multi-agent workflow (configured as custom adapter in bench.yaml)
workflow-bench run --workflow multi-agent --tasks tier1 --runs 1 --tag multi-agent

# Run in parallel (4 tasks at once)
workflow-bench run --tasks all --runs 1 --tag fast --parallel 4

# Distributed execution: shard 1 of 4
workflow-bench run --tasks all --runs 1 --tag distributed --shard 1/4

# Keep worktrees for debugging
workflow-bench run --tasks tier1/fix-handler-bug --runs 1 --tag debug --keep-worktree

# Use a custom plan
workflow-bench run --tasks tier1/fix-handler-bug --plan ./my-plan.md --tag custom-plan
```

**Notes**:
- Each run creates an isolated git worktree. The original task repo is never modified.
- If a run with the same `(tag, workflow, task_id, run_number)` already completed, it is skipped (checkpoint/resume).
- Timeout per task = `estimated_minutes * timeout_multiplier` (default 3x).
- `--shard` splits the task list into equal parts; use `merge` to combine results.

---

## workflow-bench report

**Description**: Generate a summary report for a tagged benchmark run.

**Usage**: `workflow-bench report --tag <tag> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | *(required)* | Tag to generate report for |
| `--format` | string | `markdown` | Output format: `markdown`, `md`, or `html` |

**Examples**:

```bash
# Print markdown report to stdout
workflow-bench report --tag baseline

# Generate HTML report
workflow-bench report --tag baseline --format html > report.html
```

**Notes**:
- Output includes overall pass rate (with Wilson CI), average correctness, per-tier summary, and per-task L1-L4 breakdowns.
- When LLM Judge is enabled, rubric scores are included (7 dimensions + composite).
- When multiple runs per task exist, stability data is shown.

---

## workflow-bench compare

**Description**: Compare benchmark results between two tagged runs.

**Usage**: `workflow-bench compare --left <tag> --right <tag> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--left` | string | *(required)* | Left tag for comparison |
| `--right` | string | *(required)* | Right tag for comparison |
| `--format` | string | `markdown` | Output format: `markdown`, `md`, or `html` |
| `--pairwise` | bool | false | Run LLM pairwise comparison (requires judge enabled) |

**Examples**:

```bash
# Compare vanilla vs multi-agent
workflow-bench compare --left vanilla-tag --right v4-tag

# Compare with pairwise LLM judgment
workflow-bench compare --left vanilla-tag --right v4-tag --pairwise

# HTML comparison report
workflow-bench compare --left v1 --right v2 --format html > compare.html
```

**Notes**:
- Matches tasks by ID across both tags; unmatched tasks are reported separately.
- Statistical significance is computed using Wilson CI non-overlap test.
- `--pairwise` requires `ANTHROPIC_API_KEY` and judge enabled in config.

---

## workflow-bench trend

**Description**: Show metrics trend across multiple benchmark tags.

**Usage**: `workflow-bench trend --tags <tag1,tag2,...> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tags` | string | *(required)* | Comma-separated list of tags to compare |
| `--format` | string | `markdown` | Output format: `markdown`, `md`, or `html` |

**Examples**:

```bash
# View trend across three versions
workflow-bench trend --tags v1,v2,v3

# HTML trend chart
workflow-bench trend --tags v1,v2,v3 --format html > trend.html
```

**Output**:

```markdown
# Trend Report

| Tag | Date | Pass Rate | Avg Correctness | Avg Wall Time | Tasks |
|-----|------|-----------|-----------------|---------------|-------|
| v1  | 2026-04-01 | 75.0% | 0.82 | 2m30s | 20 |
| v2  | 2026-04-05 | 82.0% | 0.88 | 2m15s | 20 |
| v3  | 2026-04-09 | 90.0% | 0.95 | 1m50s | 20 |

Trend: Pass Rate +15.0%, Correctness +0.13, Wall Time -26%
```

---

## workflow-bench export

**Description**: Export benchmark data as JSON or CSV.

**Usage**: `workflow-bench export --tag <tag> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | *(required)* | Tag to export |
| `--format` | string | `json` | Output format: `json` or `csv` |

**Examples**:

```bash
# Export as JSON
workflow-bench export --tag baseline --format json > results.json

# Export as CSV
workflow-bench export --tag baseline --format csv > results.csv
```

---

## workflow-bench inspect

**Description**: View raw output of a specific benchmark run.

**Usage**: `workflow-bench inspect --run-id <id>`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--run-id` | string | *(required)* | Run ID to inspect |

**Examples**:

```bash
workflow-bench inspect --run-id baseline-tier1-fix-handler-bug-run1-1234567890
```

**Notes**:
- Displays `verify.log` and `diff.patch` from the raw output directory.
- Use `--keep-worktree` during `run` to preserve the full worktree for deeper inspection.

---

## workflow-bench import

**Description**: Create a workflow-bench task from a git commit range.

**Usage**: `workflow-bench import --repo <path> --commit <range> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--repo` | string | *(required)* | Path to git repository |
| `--commit` | string | *(required)* | Commit range (e.g. `abc123..def456`) |
| `--tier` | int | `0` | Override tier (0 = auto-detect from diff size) |
| `--type` | string | | Override task type (empty = auto-detect) |
| `--output` | string | | Output directory (default: `tasks/imported/<name>/`) |

**Examples**:

```bash
# Import from git history
workflow-bench import --repo /path/to/project --commit abc123..def456

# Import with explicit tier and type
workflow-bench import --repo . --commit HEAD~1..HEAD --tier 2 --type http-server
```

---

## workflow-bench generate-variant

**Description**: Generate a task variant by copying a source task and applying identifier renames.

**Usage**: `workflow-bench generate-variant --source <path> [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--source` | string | *(required)* | Source task directory path |
| `--output` | string | `<source>-variant` | Output directory |
| `--seed` | int64 | `0` | Random seed (0 = random) |

**Examples**:

```bash
# Generate a variant of an existing task
workflow-bench generate-variant --source tasks/tier1/fix-handler-bug

# Generate with a fixed seed for reproducibility
workflow-bench generate-variant --source tasks/tier1/fix-handler-bug --seed 42
```

---

## workflow-bench merge

**Description**: Merge multiple result databases into one (for sharded execution).

**Usage**: `workflow-bench merge --from <db1> --from <db2> --to <target>`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string[] | *(required)* | Source database paths (can be repeated) |
| `--to` | string | *(required)* | Target database path |

**Examples**:

```bash
# Merge shard results into a single DB
workflow-bench merge --from shard1.db --from shard2.db --from shard3.db --to combined.db
```

**Notes**:
- Duplicate run IDs are silently ignored (INSERT OR IGNORE).
- The target database is created if it does not exist.

---

## workflow-bench clean

**Description**: Clean up benchmark data or orphaned worktrees.

**Usage**: `workflow-bench clean [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tag` | string | | Delete all runs with this tag |
| `--older-than` | string | | Delete runs older than this duration (e.g. `30d`, `24h`) |
| `--worktrees` | bool | false | Clean up orphaned worktree directories in /tmp |

**Examples**:

```bash
# Delete all runs for a specific tag
workflow-bench clean --tag old-test

# Delete runs older than 30 days
workflow-bench clean --older-than 30d

# Clean up orphaned worktrees
workflow-bench clean --worktrees

# Combine: delete old data and clean worktrees
workflow-bench clean --older-than 7d --worktrees
```

**Notes**:
- At least one of `--tag`, `--older-than`, or `--worktrees` must be specified.
- `--worktrees` removes `/tmp/bench-worktree-*` and `/tmp/bench-verify.*` directories.

---

## workflow-bench doctor

**Description**: Check environment prerequisites for running benchmarks.

**Usage**: `workflow-bench doctor [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tasks-dir` | string | `tasks` | Path to tasks directory |

**Examples**:

```bash
workflow-bench doctor
```

**Output**:

```
Checking environment...
  + Go: 1.23.0
  + Claude CLI: 1.0.12
  + ANTHROPIC_API_KEY: set
  + ~/.claude/agents/: 3 files
  + SQLite: OK
  + Tasks: 100 found
  + staticcheck: found
  x gosec: not found (optional)

All required tools present. Ready to run benchmarks.
```

---

## workflow-bench list tasks

**Description**: List all available tasks discovered from the tasks directory.

**Usage**: `workflow-bench list tasks [flags]`

**Flags**:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tasks-dir` | string | `tasks` | Path to tasks directory |

---

## workflow-bench list workflows

**Description**: List available workflow adapters.

**Usage**: `workflow-bench list workflows`

---

## workflow-bench list tags

**Description**: List all benchmark tags with run counts, dates, and workflows.

**Usage**: `workflow-bench list tags`

---

## workflow-bench validate

**Description**: Validate task definitions for correctness.

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
- `repo/` directory exists with `go.mod`
- `go build ./...` passes (30s timeout)
- `plan.md` exists and is non-empty
- `verify/e2e_test.go` or `verify/e2e_test.go.src` exists
- `estimated_minutes` > 0

---

## workflow-bench init

**Description**: Initialize the configuration directory and create a default `bench.yaml`.

**Usage**: `workflow-bench init`

**Notes**: Safe to run multiple times. Existing `bench.yaml` is not overwritten.

---

## workflow-bench version

**Description**: Print the version string.

**Usage**: `workflow-bench version`
