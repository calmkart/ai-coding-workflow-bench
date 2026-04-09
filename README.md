English | [中文](README.zh-CN.md)

# workflow-bench

[![CI](https://github.com/calmkart/ai-coding-workflow-bench/actions/workflows/ci.yml/badge.svg)](https://github.com/calmkart/ai-coding-workflow-bench/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A benchmark tool for evaluating multi-agent coding workflow strategies. Given the same plan, different workflows produce different code -- workflow-bench measures which one does it better.

## Features

- **Deterministic 4-layer verification**: build, unit tests, static analysis, E2E tests
- **Four-dimension composite scoring**: correctness, efficiency, quality, stability
- **LLM Judge**: rubric-based code quality scoring via Anthropic API (7 dimensions)
- **Pairwise comparison**: head-to-head code comparison with position bias detection
- **Multiple workflow adapters**: compare vanilla Claude CLI, multi-agent workflows, or custom commands
- **Built-in task library**: curated Go tasks across 4 difficulty tiers
- **Isolated execution**: each run gets its own git worktree -- no cross-contamination
- **Checkpoint/resume**: interrupted runs pick up where they left off
- **Parallel execution**: run multiple tasks concurrently with `--parallel`
- **Sharded execution**: distribute work across machines with `--shard`
- **Markdown and HTML reports**: auto-generated summaries, comparisons, and trend charts
- **Trend tracking**: view metrics evolution across multiple tagged runs
- **Data management**: export (JSON/CSV), import from git history, merge sharded results, clean old data
- **SQLite storage**: all results persisted locally for querying and comparison
- **Environment diagnostics**: `doctor` command validates prerequisites

## Quick Start

### Prerequisites

- Go 1.23+
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) installed and configured
- `ANTHROPIC_API_KEY` environment variable set

### Install

```bash
go install github.com/calmkart/ai-coding-workflow-bench/cmd/workflow-bench@latest
```

### Build from Source

```bash
git clone https://github.com/calmkart/ai-coding-workflow-bench.git
cd ai-coding-workflow-bench

# Build the binary
go build -o workflow-bench ./cmd/workflow-bench

# Initialize config directory (~/.claude/workflow-bench/)
./workflow-bench init

# Validate built-in tasks
./workflow-bench validate --tasks tier1

# Run a benchmark (vanilla: Claude CLI direct)
./workflow-bench run --workflow vanilla --tasks tier1 --runs 1 --tag my-first-run

# View results
./workflow-bench report --tag my-first-run
```

## Architecture

### Package Structure (10 packages)

```
workflow-bench/
├── cmd/workflow-bench/     CLI entry point (cobra)
├── internal/
│   ├── config/             bench.yaml loading + task discovery
│   ├── engine/
│   │   ├── runner.go       Orchestrates: load tasks -> adapter.Run -> verify -> store
│   │   ├── isolation.go    Git worktree creation/cleanup
│   │   ├── verify.go       Generates verify.sh from embedded templates
│   │   └── collector.go    Parses BENCH_RESULT output into L1-L4 scores
│   ├── adapter/
│   │   ├── adapter.go      Adapter interface + registry
│   │   ├── vanilla.go      Claude CLI direct execution
│   │   └── custom.go       User-defined command execution
│   ├── judge/              LLM Judge: rubric scoring + pairwise comparison
│   │   ├── rubric.go       7-dimension rubric evaluation via Anthropic API
│   │   └── pairwise.go     Head-to-head code comparison with position bias detection
│   ├── metrics/
│   │   ├── correctness.go  Correctness score formula
│   │   ├── cost.go         Unified cost estimation
│   │   └── statistics.go   Wilson CI, significance testing, stability scoring
│   ├── store/
│   │   ├── db.go           SQLite CRUD (pure Go, no CGO)
│   │   └── schema.sql      Database schema
│   ├── report/
│   │   ├── summary.go      Markdown/HTML summary reports
│   │   ├── compare.go      Side-by-side comparison reports
│   │   ├── trend.go        Multi-tag trend reports
│   │   ├── export.go       JSON/CSV data export
│   │   └── html.go         HTML report rendering
│   ├── importer/           Task import from git history
│   └── taskgen/            Task variant generation
└── tasks/                  Built-in task library (100 tasks)
    ├── tier1/              20 simple tasks (~5 min)
    ├── tier2/              32 medium tasks (~10 min)
    ├── tier3/              29 complex tasks (~15-20 min)
    └── tier4/              19 advanced tasks (~25-30 min)
```

### Data Flow

```
                    ┌─────────────┐
                    │  bench.yaml │
                    └──────┬──────┘
                           │
  ┌──────────┐      ┌──────▼──────┐      ┌──────────┐
  │  tasks/  │─────►│   runner    │─────►│  SQLite   │
  │ task.yaml│      │  (engine)   │      │ results.db│
  │ plan.md  │      └──┬──────┬───┘      └─────┬─────┘
  │ repo/    │         │      │                │
  └──────────┘    ┌────▼──┐ ┌─▼──────┐   ┌────▼─────┐
                  │adapter│ │verify.sh│   │  report   │
                  │(claude│ │L1-L4   │   │(markdown) │
                  │  CLI) │ │checks  │   └──────────┘
                  └───────┘ └────────┘
```

1. **runner** discovers tasks, creates an isolated git worktree per run
2. **adapter** executes the workflow (e.g., `claude -p` with the plan)
3. **verify.sh** runs L1-L4 checks against the modified worktree
4. **collector** parses the verify output into structured scores
5. Results are stored in SQLite; **report** renders them as Markdown

## CLI Commands

| Command | Description |
|---------|-------------|
| `run` | Run benchmark against tasks (`--parallel`, `--shard`, `--keep-worktree`, `--pairwise`) |
| `report` | Generate summary report (`--format markdown\|html`) |
| `compare` | Compare two tagged runs side-by-side (`--pairwise` for LLM comparison) |
| `trend` | Show metrics trend across multiple tags (`--tags v1,v2,v3`) |
| `export` | Export raw data as JSON or CSV |
| `inspect` | View raw output (verify.log, diff.patch) for a specific run |
| `import` | Create task from git commit history |
| `generate-variant` | Generate task variant with renamed identifiers |
| `merge` | Merge multiple result databases (for sharded execution) |
| `clean` | Delete runs by tag, by age, or clean orphaned worktrees |
| `doctor` | Check environment prerequisites |
| `list tasks` | List all available tasks |
| `list workflows` | List available workflow adapters |
| `list tags` | List all benchmark tags with run counts |
| `validate` | Validate task definitions (structure, build, tests) |
| `init` | Initialize config directory and default bench.yaml |
| `version` | Print version |

See [docs/cli-reference.md](docs/cli-reference.md) for full flag reference and examples.

## Built-in Tasks

100 tasks across 4 difficulty tiers and 5 code types:

| Tier | Count | Difficulty | Est. Time |
|------|-------|------------|-----------|
| T1   | 20    | Simple     | ~5 min    |
| T2   | 32    | Medium     | ~10 min   |
| T3   | 29    | Complex    | ~15-20 min|
| T4   | 19    | Advanced   | ~25-30 min|

| Type | Count | Examples |
|------|-------|---------|
| http-server | 32 | CRUD fixes, middleware, auth, rate limiting, RBAC |
| library | 24 | String utils, LRU cache, circuit breaker, B-tree |
| cli | 15 | Flag parsing, config loading, interactive mode |
| concurrency | 15 | Worker pool, pub/sub, actor model, MapReduce |
| reconciler | 14 | State machine, finalizer, leader election, GC |

Run `workflow-bench list tasks` to see the full list. See [docs/tasks.md](docs/tasks.md) for the task authoring guide.

## Scoring

### Four-Layer Verification (L1-L4)

| Layer | What | Weight |
|-------|------|--------|
| L1 Build | `go build ./...` | Gate (fail = score 0) |
| L2 Unit Tests | `go test -race ./...` | 20% |
| L3 Static Analysis | `go vet` + optional `staticcheck` | 10% |
| L4 E2E Tests | Ground-truth E2E via `httptest` | 70% |

### Correctness Formula

```
if L1 == FAIL:
    correctness = 0.0
else:
    l2 = passed / total
    l3 = max(0, 1.0 - issues * 0.05)
    l4 = passed / total
    correctness = 0.20 * l2 + 0.10 * l3 + 0.70 * l4
```

Critical verification target (VT) failures each deduct an additional 0.1 from the correctness score (clamped to 0).

See [docs/scoring.md](docs/scoring.md) for the full scoring breakdown.

## Configuration

Config lives at `~/.claude/workflow-bench/bench.yaml` (created by `init`).

```yaml
workflows:
  vanilla:
    adapter: vanilla

  # Example: multi-agent workflow using custom adapter
  # multi-agent:
  #   adapter: custom
  #   setup_commands:
  #     - "mkdir -p .claude/agents"
  #     - "cp -r ~/.claude/agents/*.md .claude/agents/"
  #     - "cp -r ~/.claude/agents/reference .claude/agents/ 2>/dev/null || true"
  #     - "mkdir -p .planning/manager"
  #   entry_command: |
  #     claude --agent manager -p "You are running a benchmark evaluation. Execute your FULL multi-agent workflow:
  #     1. Read the plan from $BENCH_PLAN_FILE
  #     2. Spawn Architect agent to formalize the plan into a spec
  #     3. Spawn Coding agent to implement from the spec
  #     4. Spawn Testing agent to write scenario tests
  #     5. Spawn Challenger agent to review the implementation
  #     6. Fix any issues found by Challenger
  #     7. Repeat until Challenger passes
  #     IMPORTANT: Do NOT skip any phase. All permission gates are pre-approved." --output-format json --dangerously-skip-permissions

defaults:
  runs_per_task: 3
  timeout_multiplier: 3
```

See [docs/configuration.md](docs/configuration.md) for full field reference including both adapters (`vanilla`, `custom`).

## Development

```bash
# Build
make build

# Run tests
make test

# Clean
make clean
```

### Custom Workflows (No Go Code Required)

Use the `custom` adapter to run any command as a workflow:

```yaml
workflows:
  my-workflow:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"
```

```bash
./workflow-bench run --workflow my-workflow --tasks tier1 --runs 1 --tag test
```

See [docs/configuration.md](docs/configuration.md) for full custom adapter reference.

### Adding a New Adapter (Go)

1. Create `internal/adapter/myadapter.go` implementing the `Adapter` interface
2. Register it in `adapter.Registry` in `adapter.go`
3. Add a workflow entry in `bench.yaml`

See [docs/development.md](docs/development.md) for the full development guide.

## Roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| **v1 (P1-P6)** | Wilson CI, VT detection, data export, inspect/doctor, 4-dim composite | Done |
| **v2 (P7-P12)** | verify.sh JSON, unified cost, VT mapping, Judge timeout, stability score | Done |
| **v2 (P13-P18)** | LLM Judge rubric, HTML reports, compare enhancements, pairwise comparison | Done |
| **v2 (P19-P22)** | Task import from git, variant generation, sharded execution, DB merge | Done |
| **v2 (P23-P25)** | Trend reports, clean command, documentation sync | Done |

## License

MIT
