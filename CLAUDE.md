# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
make build                          # Build binary -> ./workflow-bench
make test                           # Run all tests with race detector
make clean                          # Remove binary and test cache

go test ./internal/metrics/... -v   # Run tests for a specific package
go test ./... -count=1 -race        # Equivalent to make test
```

No CGO required -- the SQLite driver (`modernc.org/sqlite`) is pure Go.

## Architecture

This is a Go CLI tool (cobra) that benchmarks multi-agent coding workflows. It runs an AI coding agent against curated Go tasks, then verifies the output through a 4-layer check (build, unit tests, static analysis, E2E tests) and scores the result.

### Execution Pipeline

`runner.go` orchestrates: **discover tasks -> create git worktree -> adapter.Run -> verify.sh -> parse BENCH_RESULT -> score -> store in SQLite**

Each benchmark run gets an isolated git worktree (created from `tasks/tier*/*/repo/`) so runs don't contaminate each other. Worktrees are cleaned up after each run.

### Key Interfaces

- **Adapter** (`internal/adapter/adapter.go`): `Setup(ctx, worktreeDir)` + `Run(ctx, worktreeDir, planContent) -> RunOutput`. Two built-in: `vanilla` (claude -p), `custom` (user-defined shell command). Registered in `adapter.Registry`.

- **Verify output protocol**: verify.sh must emit `BENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5` -- parsed by `collector.go` via regex.

### Correctness Scoring

L1 (build) is a gate -- if it fails, score is 0. Otherwise: `0.20 * L2 + 0.10 * L3 + 0.70 * L4`. Critical VT failures deduct 0.1 each (clamped to 0).

### Task Structure

Each task lives at `tasks/tier{1-4}/{task-name}/` and contains:
- `task.yaml` -- metadata (id, tier, type, estimated_minutes)
- `plan.md` -- the plan given to the AI agent
- `repo/` -- a compilable Go project (must be a git repo for worktree creation)
- `verify/e2e_test.go.src` -- ground-truth E2E tests (`.src` extension avoids Go tooling compiling it in place)

Tasks are discovered via `filepath.Glob("tasks/tier*/*/task.yaml")` -- no index file.

### Embedded Templates

Verify script templates and the DB schema are embedded via `//go:embed`. Changes to `internal/engine/templates/*.sh.tmpl` or `internal/store/schema.sql` require rebuilding.

### Config and Storage

- Config: `~/.claude/workflow-bench/bench.yaml` (created by `workflow-bench init`)
- Results DB: `~/.claude/workflow-bench/results.db` (SQLite, WAL mode)
- When `--config` is specified, DB path is derived from the config file's directory (full isolation per config location)

## Roadmap Context

P1-P2 (CLI, adapters, 100 tasks, L1-L4 verify) are done. P3+ (LLM judge, pairwise ranking, stability scoring) are planned but not yet implemented.
