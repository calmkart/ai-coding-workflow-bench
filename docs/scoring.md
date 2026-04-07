# Scoring System

## Overview

workflow-bench uses a multi-layer scoring system. Currently (P1), only the correctness score is implemented. Future phases will add efficiency, code quality (LLM Judge), and stability scores that combine into a final composite score.

## Four-Layer Verification (L1-L4)

Every run goes through four verification layers, executed by an auto-generated `verify.sh` script.

### L1: Build

```bash
go build ./...
```

Binary gate. If the code does not compile, correctness = 0.0 and no further checks run.

### L2: Unit Tests

```bash
go test ./... -count=1 -race -v
```

Runs the task's existing unit tests with the race detector. Counts passing and failing tests.

### L3: Static Analysis

```bash
go vet ./...
staticcheck ./...   # if installed
gosec ./...          # if installed
```

Counts lint issues. Each issue deducts 5% from the L3 sub-score. Only `go vet` is required; `staticcheck` and `gosec` are used if available.

### L4: E2E Tests

```bash
go test -v -run TestBenchE2E -count=1 -race ./...
```

Runs the ground-truth E2E tests from the task's `verify/` directory. These tests are not visible to the workflow agent -- they are copied into the worktree only during verification.

For `http-server` type tasks, E2E tests use `httptest` and call `setupRouter()` to exercise the API.

## Correctness Score Formula

The correctness score is a weighted combination of L2, L3, and L4 results, gated by L1:

```
if L1 == FAIL:
    correctness = 0.0
else:
    l2_score = ut_passed / ut_total         # 0.0-1.0
    l3_score = max(0, 1.0 - issues * 0.05)  # each issue deducts 5%
    l4_score = e2e_passed / e2e_total       # 0.0-1.0

    correctness = 0.20 * l2_score + 0.10 * l3_score + 0.70 * l4_score
```

**Edge cases**:
- If `ut_total == 0` (no unit tests), `l2_score = 1.0`
- If `e2e_total == 0` (no E2E tests), `l4_score = 1.0`

### Example Calculations

**Perfect run**: L1=PASS, L2=8/8, L3=0 issues, L4=5/5

```
l2 = 8/8 = 1.0
l3 = max(0, 1.0 - 0*0.05) = 1.0
l4 = 5/5 = 1.0
correctness = 0.20*1.0 + 0.10*1.0 + 0.70*1.0 = 1.00
```

**Partial success**: L1=PASS, L2=6/8, L3=2 issues, L4=3/5

```
l2 = 6/8 = 0.75
l3 = max(0, 1.0 - 2*0.05) = 0.90
l4 = 3/5 = 0.60
correctness = 0.20*0.75 + 0.10*0.90 + 0.70*0.60 = 0.15 + 0.09 + 0.42 = 0.66
```

**Build failure**: L1=FAIL

```
correctness = 0.0
```

## Verification Targets (VT)

Tasks can define verification targets -- known pitfalls that an agent might introduce during refactoring. VTs are categorized across 9 categories (78 total patterns defined):

| Category | Example VTs |
|----------|------------|
| Concurrency | goroutine leaks, data races, deadlocks |
| Error handling | error swallowing, nil interface trap |
| Memory/Resources | unclosed HTTP body, file descriptor leaks |
| Interface/Types | nil interface pitfall, type assertion panic |
| Package/Dependencies | circular imports, init() order dependence |
| HTTP | missing server timeout, handler panic recovery |
| Distributed | no backoff retry, non-idempotent retry |
| K8s Operator | infinite reconcile, finalizer not cleaned |
| Testing | test pollution, time-dependent tests |

### VT Deduction (planned)

When VT detection is implemented, critical VT failures deduct from the correctness score:

```
correctness = max(0, correctness - 0.1 * critical_vt_fail_count)
```

Each critical VT failure costs 0.1 points (10%), with a floor of 0.

## Pass/Fail Determination

A run is considered "passed" when:
- L1 build succeeds AND
- All L4 E2E tests pass (L4 passed == L4 total, with total > 0)

This is used for the pass rate metric in reports.

## Future Scoring Dimensions (P2+)

### Efficiency Score

```
efficiency = 1.0 - min(1.0, cost_usd / cost_budget)
```

Where `cost_budget` is per-tier (T1=$0.50, T2=$1.00, T3=$2.00, T4=$5.00 by default). If token data is unavailable, efficiency is N/A and excluded from the composite.

### LLM Quality Score

Six dimensions evaluated by an LLM Judge using a Rubric (0-5 scale per dimension):

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Correctness | 25% | Logic correctness, edge cases, hidden bugs |
| Readability | 15% | Naming, structure, comments, control flow |
| Simplicity | 15% | No over-engineering, simplest working solution |
| Robustness | 15% | Error handling, resource management, concurrency safety |
| Minimality | 15% | Clean diff, no unrelated changes, proportional scope |
| Maintainability | 15% | Cohesion, coupling, extensibility, pattern consistency |

Plus two supplementary dimensions reported separately:
- **Go Idioms**: Go-specific style conformance
- **Workflow Private**: Plan adherence (dynamically generated from plan content)

### Stability Score

```
stability = pass_count / K
```

Where K is `runs_per_task` (need K >= 3 for meaningful measurement).

### Composite Score

```
final = 0.40 * correctness + 0.25 * efficiency + 0.25 * quality + 0.10 * stability
```

When a dimension is N/A, its weight is redistributed proportionally to the remaining dimensions.

**Security veto**: If `gosec` finds High/Critical issues, the run is flagged `SECURITY_FAIL` (reported separately, does not affect the numeric score).
