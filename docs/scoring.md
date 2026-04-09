# Scoring System

## Overview

workflow-bench uses a four-dimension composite scoring system:

1. **Correctness** (40%) -- deterministic L1-L4 verification
2. **Efficiency** (25%) -- token usage and cost
3. **Quality** (25%) -- LLM Judge rubric scoring (7 dimensions)
4. **Stability** (10%) -- consistency across multiple runs

## Four-Layer Verification (L1-L4)

Every run goes through four verification layers, executed by an auto-generated `verify.sh` script.

### L1: Build

```bash
go build ./...
```

Binary gate. If the code does not compile, correctness = 0.0 and no further checks run.

### L2: Unit Tests

```bash
go test -json ./... -count=1 -race
```

Runs the task's existing unit tests with the race detector. Uses `go test -json` for precise pass/fail counting (only top-level test functions are counted; subtests are filtered out).

### L3: Static Analysis

```bash
go vet ./...
staticcheck ./...   # if installed
gosec ./...          # if installed
```

Counts lint issues. Each issue deducts 5% from the L3 sub-score. Only `go vet` is required; `staticcheck` and `gosec` are used if available.

### L4: E2E Tests

```bash
go test -json -run TestBenchE2E -count=1 -race ./...
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

### Verification Target (VT) Deduction

Critical VT failures deduct from the correctness score:

```
correctness = max(0, correctness - 0.1 * critical_vt_fail_count)
```

Each critical VT failure costs 0.1 points (10%), with a floor of 0. VT detection types supported:

| Detection Type | Verification Layer |
|---------------|-------------------|
| `go build`, `build` | L1 |
| `unit test`, `ut` | L2 |
| `go vet`, `staticcheck`, `lint` | L3 |
| `e2e test case`, `e2e test`, `e2e` | L4 |
| `race detector` | Requires `go test -race` |

### Example Calculations

**Perfect run**: L1=PASS, L2=8/8, L3=0 issues, L4=5/5

```
correctness = 0.20*1.0 + 0.10*1.0 + 0.70*1.0 = 1.00
```

**Partial success**: L1=PASS, L2=6/8, L3=2 issues, L4=3/5

```
correctness = 0.20*0.75 + 0.10*0.90 + 0.70*0.60 = 0.66
```

**Build failure**: L1=FAIL -> correctness = 0.0

## Efficiency Score

```
efficiency = 1.0 - min(1.0, cost_usd / cost_budget)
```

Where `cost_budget` is per-tier (T1=$0.50, T2=$1.00, T3=$2.00, T4=$5.00 by default). Cost is computed using the unified `metrics.EstimateCost()` function with configurable per-model pricing.

If token data is unavailable, efficiency is N/A and excluded from the composite.

## LLM Judge -- Rubric Scoring

When `judge.enabled: true` in config, each completed run is evaluated by an LLM Judge using a structured rubric. Seven dimensions are scored on a 0-5 scale:

| Dimension | Weight | What it measures |
|-----------|--------|-----------------|
| Correctness | 25% | Logic correctness, edge cases, hidden bugs |
| Readability | 15% | Naming, structure, comments, control flow |
| Simplicity | 15% | No over-engineering, simplest working solution |
| Robustness | 15% | Error handling, resource management, concurrency safety |
| Minimality | 15% | Clean diff, no unrelated changes, proportional scope |
| Maintainability | 15% | Cohesion, coupling, extensibility, pattern consistency |
| Go Idioms | (supplementary) | Go-specific style conformance |

The rubric composite is a weighted average of the 6 main dimensions (Go Idioms is reported separately).

### Consistency Validation

The judge response includes boolean indicators (e.g., "has_hidden_bugs: true") alongside numeric scores. A consistency check flags cases where boolean indicators contradict the numeric score (e.g., 5/6 booleans positive but score = 2). Warnings are included in the report.

### Judge Configuration

```yaml
judge:
  enabled: true
  model: "claude-sonnet-4-20250514"
  input_price_per_mtok: 3.0
  output_price_per_mtok: 15.0
  repeat: 1  # number of judge evaluations per run (for averaging)
```

## Pairwise Comparison

With `--pairwise` flag on the `compare` command, the LLM Judge performs head-to-head comparisons between code from two tags. For each common task:

1. Both implementations' diffs are presented to the Judge
2. The Judge evaluates across multiple dimensions (correctness, readability, etc.)
3. **Position bias detection**: the comparison is run twice with swapped order; results are marked as "position consistent" when both orderings agree

Results are stored in the `pairwise_results` table and displayed in the comparison report.

## Stability Score

```
stability = pass_count / K
```

Where K is `runs_per_task` (need K >= 3 for meaningful measurement). The stability score measures how consistently the workflow produces correct results across multiple runs of the same task.

## Composite Score (Four Dimensions)

```
final = 0.40 * correctness + 0.25 * efficiency + 0.25 * quality + 0.10 * stability
```

When a dimension is N/A (e.g., no token data for efficiency, no Judge for quality, single run for stability), its weight is redistributed proportionally to the remaining dimensions. For example, if only correctness and efficiency are available:

```
final = (0.40/(0.40+0.25)) * correctness + (0.25/(0.40+0.25)) * efficiency
      = 0.615 * correctness + 0.385 * efficiency
```

## Statistical Methods

### Wilson Score Confidence Interval

Pass rates are reported with 95% Wilson Score confidence intervals:

```
Pass Rate: 85.0% [72.3-93.1]
```

This provides more reliable bounds than naive binomial intervals, especially at small sample sizes.

### Significance Testing

When comparing two tags, statistical significance is determined by checking if the Wilson CIs for the left and right pass rates overlap. Non-overlapping CIs indicate a statistically significant difference (marked with `*` in the report).

### Low Sample Size Warning

When the total number of runs (K) is less than 5, a warning is displayed:

```
Low sample size (K < 5): results may not be statistically reliable
```

## Pass/Fail Determination

A run is considered "passed" when:
- L1 build succeeds AND
- All L4 E2E tests pass (L4 passed == L4 total, with total > 0)

This is used for the pass rate metric in reports.

## Security Veto

If `gosec` finds High/Critical issues, the run is flagged `SECURITY_FAIL` (reported separately, does not affect the numeric score).
