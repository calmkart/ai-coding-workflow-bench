# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [0.1.0] - 2026-04-08

### Added
- CLI with commands: `run`, `report`, `list`, `validate`, `init`, `version`
- Workflow adapters: `vanilla` (Claude CLI direct), `custom` (user-defined commands)
- 4-layer verification: L1 build, L2 unit tests, L3 static analysis, L4 E2E tests
- Correctness scoring with weighted formula
- 100 built-in Go tasks across 4 difficulty tiers
- SQLite storage for benchmark results (pure Go, no CGO)
- Markdown report generation
- Git worktree isolation per run
- Checkpoint/resume for interrupted runs
- Configuration via `bench.yaml`
