# Configuration Reference

## Overview

workflow-bench uses a YAML configuration file for workflow definitions and default parameters. Run `workflow-bench init` to create the default config.

## File Location

| Path | Description |
|------|-------------|
| `~/.claude/workflow-bench/bench.yaml` | Default config location |
| `~/.claude/workflow-bench/results.db` | SQLite database |
| `~/.claude/workflow-bench/` | Home directory for all workflow-bench data |

Use `--config /path/to/bench.yaml` to specify an alternate config file.

If no config file exists and `--config` is not provided, workflow-bench uses built-in defaults.

## bench.yaml Full Reference

```yaml
# Workflow definitions
workflows:
  vanilla:
    adapter: vanilla               # Direct Claude CLI execution

  my-workflow:
    adapter: custom                # User-defined command
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"

# Default run parameters
defaults:
  runs_per_task: 3                 # Number of runs per task (overridden by --runs)
  timeout_multiplier: 3            # Timeout = estimated_minutes * this multiplier

  # Planned for P2+:
  # model: "claude-sonnet-4-20250514"  # (illustrative; update to current model names)
  # cost_budget:
  #   tier1: 0.50
  #   tier2: 1.00
  #   tier3: 2.00
  #   tier4: 5.00

# LLM Judge settings (planned for P3+)
# judge:
#   model: "claude-sonnet-4-20250514"  # (illustrative; update to current model names)
#   ensemble_model: "gpt-4o"  # (illustrative; update to current model names)
#   pairwise_mode: "compact"       # compact | full
#   enable_ensemble: false

# Import settings (planned for P6+)
# import:
#   time_window_hours: 72
#   jaccard_threshold: 0.3
```

## Field Reference

### workflows

Map of workflow name to configuration. Each entry defines a named workflow that can be selected with `--workflow`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `adapter` | string | yes | Name of the adapter implementation to use |

Currently available adapters:
- `vanilla` -- Runs `claude -p` directly with the plan content
- `custom` -- User-defined command execution (see below)

### defaults

Default parameters for benchmark runs.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `runs_per_task` | int | 3 | Number of times to run each task (per workflow). Overridden by `--runs`. |
| `timeout_multiplier` | int | 3 | Multiplied by `estimated_minutes` from task.yaml to get the per-run timeout. |

### defaults.cost_budget (planned)

Per-tier cost budgets for efficiency score normalization.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tier1` | float | 0.50 | Max expected USD cost for a T1 task |
| `tier2` | float | 1.00 | Max expected USD cost for a T2 task |
| `tier3` | float | 2.00 | Max expected USD cost for a T3 task |
| `tier4` | float | 5.00 | Max expected USD cost for a T4 task |

### judge (planned)

LLM Judge configuration for code quality scoring.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `model` | string | `claude-sonnet-4-20250514` | Model for Rubric and Pairwise evaluation (illustrative; update to current model names) |
| `ensemble_model` | string | `gpt-4o` | Second model for ensemble evaluation (illustrative; update to current model names) |
| `pairwise_mode` | string | `compact` | `compact` (14 calls) or `full` (up to 50 calls) |
| `enable_ensemble` | bool | false | Enable multi-model ensemble for T3-T4 tasks |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | yes | API key for Claude CLI and LLM Judge |
| `OPENAI_API_KEY` | no | Required only when `judge.enable_ensemble` is true |
| `HOME` | yes | Used to locate `~/.claude/workflow-bench/` |

## Config Resolution Order

1. Explicit `--config` flag
2. Default path: `~/.claude/workflow-bench/bench.yaml`
3. Built-in defaults (if no file exists)

Config values are merged with defaults: if your config only specifies `workflows`, the `defaults` section still uses built-in values.

## When to Use Which Adapter

| Adapter | Use When | Example |
|---------|----------|---------|
| `vanilla` | Baseline testing with direct Claude CLI | `claude -p` with plan content |
| `custom` | Any other tool, wrapper script, or custom configuration | Aider, Cursor, custom scripts |

Use `vanilla` as the baseline, then compare against `custom` workflows.

## Custom Adapter

The `custom` adapter lets you run any command as a workflow without writing Go code. Configure it in `bench.yaml`:

```yaml
workflows:
  my-workflow:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"
      - "mkdir -p .planning/manager"
```

### Custom Adapter Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entry_command` | string | yes | Main command to execute (run via `bash -c`) |
| `setup_commands` | list of strings | no | Commands to run before entry_command (in order) |

### Environment Variables (set by adapter)

The following environment variables are available to `entry_command`:

| Variable | Description |
|----------|-------------|
| `BENCH_REPO_DIR` | Absolute path to the worktree directory |
| `BENCH_PLAN_FILE` | Absolute path to the plan file (`.bench-plan.md`) |
| `BENCH_PLAN_PROMPT` | Convenience prompt: `"Read the plan from <BENCH_PLAN_FILE> and implement it."` |

### stdout JSON (optional)

If `entry_command` writes JSON to stdout containing a `usage` field, token data is extracted:

```json
{"usage": {"input_tokens": 100, "output_tokens": 50}, "tool_uses": 5}
```

If stdout is not valid JSON or lacks `usage`, token data is reported as N/A.

### Examples

**Example 1: Claude CLI with custom agents**

```yaml
workflows:
  my-agents:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"
      - "mkdir -p .planning/manager"
```

**Example 2: Shell script wrapper**

```yaml
workflows:
  my-script:
    adapter: custom
    entry_command: |
      ~/scripts/run-coding-agent.sh "$BENCH_PLAN_FILE" "$BENCH_REPO_DIR"
```

```bash
# Run with a custom workflow
workflow-bench run --workflow my-agents --tasks tier1 --runs 1 --tag test
```

## Adding a New Adapter (Go)

To add a workflow adapter in Go:

1. Implement the `Adapter` interface in `internal/adapter/`
2. Register it in `adapter.Registry`
3. Add a workflow entry in your `bench.yaml`
4. Run with `--workflow <name>`
