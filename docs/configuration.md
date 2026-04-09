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

# LLM Judge settings
judge:
  enabled: false                     # Set to true to enable rubric scoring
  model: "claude-sonnet-4-20250514"  # Model for Rubric and Pairwise evaluation
  input_price_per_mtok: 3.0         # Input token price per million tokens
  output_price_per_mtok: 15.0       # Output token price per million tokens
  repeat: 1                          # Number of judge evaluations per run
```

## Field Reference

### workflows

Map of workflow name to configuration. Each entry defines a named workflow that can be selected with `--workflow`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `adapter` | string | yes | Name of the adapter implementation to use |

Currently available adapters:
- `vanilla` -- Runs `claude --bare -p` directly with the plan content. The `--bare` flag ensures a pure model baseline: no plugins, no CLAUDE.md, no session-start hooks.
- `custom` -- User-defined command execution (see below)

### defaults

Default parameters for benchmark runs.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `runs_per_task` | int | 3 | Number of times to run each task (per workflow). Overridden by `--runs`. |
| `timeout_multiplier` | int | 3 | Multiplied by `estimated_minutes` from task.yaml to get the per-run timeout. |

### defaults.cost_budget

Per-tier cost budgets for efficiency score normalization.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tier1` | float | 0.50 | Max expected USD cost for a T1 task |
| `tier2` | float | 1.00 | Max expected USD cost for a T2 task |
| `tier3` | float | 2.00 | Max expected USD cost for a T3 task |
| `tier4` | float | 5.00 | Max expected USD cost for a T4 task |

### judge

LLM Judge configuration for code quality scoring.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | false | Enable rubric scoring after each run |
| `model` | string | `claude-sonnet-4-20250514` | Model for Rubric and Pairwise evaluation |
| `input_price_per_mtok` | float | 3.0 | Input token price per million tokens (for cost tracking) |
| `output_price_per_mtok` | float | 15.0 | Output token price per million tokens |
| `repeat` | int | 1 | Number of judge evaluations per run (for score averaging) |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | yes | API key for Claude CLI and LLM Judge |
| `HOME` | yes | Used to locate `~/.claude/workflow-bench/` |

## Config Resolution Order

1. Explicit `--config` flag
2. Default path: `~/.claude/workflow-bench/bench.yaml`
3. Built-in defaults (if no file exists)

Config values are merged with defaults: if your config only specifies `workflows`, the `defaults` section still uses built-in values.

## When to Use Which Adapter

| Adapter | Use When | Example |
|---------|----------|---------|
| `vanilla` | Pure model baseline (no plugins, no CLAUDE.md, no hooks) | `claude --bare -p` with plan content |
| `custom` | Any other tool, wrapper script, multi-agent workflow, or custom configuration | Aider, Cursor, custom scripts, `claude --agent manager` |

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

**Example 1: Multi-agent workflow (replaces the former v4-claude adapter)**

```yaml
workflows:
  multi-agent:
    adapter: custom
    setup_commands:
      - "mkdir -p .claude/agents"
      - "cp -r ~/.claude/agents/*.md .claude/agents/"
      - "cp -r ~/.claude/agents/reference .claude/agents/ 2>/dev/null || true"
      - "mkdir -p .planning/manager"
    entry_command: |
      claude --agent manager -p "You are running a benchmark evaluation. Execute your FULL multi-agent workflow:
      1. Read the plan from $BENCH_PLAN_FILE
      2. Spawn Architect agent to formalize the plan into a spec
      3. Spawn Coding agent to implement from the spec
      4. Spawn Testing agent to write scenario tests
      5. Spawn Challenger agent to review the implementation
      6. Fix any issues found by Challenger
      7. Repeat until Challenger passes
      IMPORTANT: Do NOT skip any phase. All permission gates are pre-approved." --output-format json --dangerously-skip-permissions
```

**Example 2: Claude CLI with custom agents**

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

**Example 3: Shell script wrapper**

```yaml
workflows:
  my-script:
    adapter: custom
    entry_command: |
      ~/scripts/run-coding-agent.sh "$BENCH_PLAN_FILE" "$BENCH_REPO_DIR"
```

```bash
# Run with a custom workflow
workflow-bench run --workflow multi-agent --tasks tier1 --runs 1 --tag test
```

## Benchmarking Superpowers

To benchmark [superpowers](https://github.com/obra/superpowers-marketplace) against the vanilla baseline:

1. Install the plugin:
   ```
   /plugin marketplace add obra/superpowers-marketplace
   /plugin install superpowers@superpowers-marketplace
   ```

2. Add a workflow to `bench.yaml`:
   ```yaml
   workflows:
     vanilla:
       adapter: vanilla

     superpowers:
       adapter: custom
       entry_command: |
         claude --plugin-dir ~/.claude/plugins/cache/claude-plugins-official/superpowers -p "$BENCH_PLAN_PROMPT" --output-format json --dangerously-skip-permissions
   ```

3. Run both workflows and compare:
   ```bash
   workflow-bench run --workflow vanilla --tasks tier1 --runs 3 --tag vanilla-baseline
   workflow-bench run --workflow superpowers --tasks tier1 --runs 3 --tag superpowers-v1
   workflow-bench compare --left vanilla-baseline --right superpowers-v1
   ```

**Note**: The `vanilla` adapter uses `--bare` mode, which disables all plugins, CLAUDE.md, and hooks. This provides a fair pure-model baseline. If you want to benchmark with your full environment (all installed plugins + CLAUDE.md), use a `default` custom workflow instead:

```yaml
workflows:
  default:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json --dangerously-skip-permissions
```

## Adding a New Adapter (Go)

To add a workflow adapter in Go:

1. Implement the `Adapter` interface in `internal/adapter/`
2. Register it in `adapter.Registry`
3. Add a workflow entry in your `bench.yaml`
4. Run with `--workflow <name>`
