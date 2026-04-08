# Development Guide

## Prerequisites

- Go 1.22+
- git
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) (for running benchmarks, not needed for development)

Optional tools (improve L3 static analysis results):
- `staticcheck` -- `go install honnef.co/go/tools/cmd/staticcheck@latest`
- `gosec` -- `go install github.com/securego/gosec/v2/cmd/gosec@latest`

## Project Structure

```
cmd/workflow-bench/main.go       CLI entry point, cobra command definitions
internal/
  config/config.go               Config loading, task discovery, task filtering
  engine/
    runner.go                    Main execution pipeline (discover -> run -> verify -> store)
    isolation.go                 Git worktree creation and cleanup
    verify.go                    Verify script generation from embedded templates
    collector.go                 Parse BENCH_RESULT output into L1-L4 scores
    templates/
      http_server.sh.tmpl        Verify script template for http-server tasks
  adapter/
    adapter.go                   Adapter interface, RunOutput type, registry
    vanilla.go                   VanillaAdapter: runs `claude -p` with the plan
    custom.go                    CustomAdapter: user-defined command execution
  metrics/
    correctness.go               Correctness score calculation (weighted L1-L4)
  store/
    db.go                        SQLite database operations (pure Go, no CGO)
    schema.sql                   Database schema (embedded via go:embed)
  report/
    summary.go                   Markdown report generation
    templates/
      summary.md.tmpl            Report template (embedded via go:embed)
tasks/
  tier1/
    fix-handler-bug/             T1 task: fix pagination off-by-one
    add-health-check/            T1 task: add /health endpoint
```

## Building

```bash
# Build binary
make build
# or
go build -o workflow-bench ./cmd/workflow-bench

# Run tests
make test
# or
go test ./... -count=1 -race

# Clean
make clean
```

## Running Tests

```bash
# All tests
go test ./... -count=1

# With race detector
go test ./... -count=1 -race

# Specific package
go test ./internal/metrics/... -v

# With coverage
go test ./... -count=1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Organization

- `*_test.go` -- Unit tests alongside source files
- `scenario_test.go` -- Scenario/integration tests within each package
- `integration_test.go` -- Cross-package integration tests (in engine/)

The store package uses `:memory:` SQLite for testing, so tests don't need file cleanup.

## Key Dependencies

| Dependency | Purpose |
|-----------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | YAML parsing for bench.yaml and task.yaml |
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGO required) |

## Code Patterns

### Embedded Templates

Templates are embedded at compile time using `//go:embed`:

```go
//go:embed templates/http_server.sh.tmpl
var httpServerTemplate string
```

This means template changes require rebuilding the binary.

### Error Handling

Functions return errors with context wrapping:

```go
return nil, fmt.Errorf("load plan: %w", err)
```

Infrastructure errors are propagated. Expected failures (e.g., test failures in verify) are handled as non-error results with structured data.

### Database Operations

All DB operations go through `internal/store/db.go`. Migrations are embedded via `schema.sql` and applied automatically on `Open()`. The database uses WAL mode for concurrent safety.

### Task Discovery

Tasks are discovered by scanning `tasks/tier*/*/task.yaml` using `filepath.Glob`. No index file is needed -- adding a new task directory with a valid `task.yaml` is sufficient.

## Adding a New Adapter

**For simple cases**: use the `custom` adapter -- it requires only YAML configuration in `bench.yaml`, no Go code. Define `entry_command` and optional `setup_commands` to run any external tool. See [configuration.md](configuration.md) for details and examples.

**For complex cases** (custom token parsing, special error handling, etc.): write a Go adapter as shown below.

1. Create `internal/adapter/myadapter.go`:

```go
package adapter

import "context"

type MyAdapter struct{}

func NewMyAdapter(cfg map[string]any) (Adapter, error) {
    return &MyAdapter{}, nil
}

func (a *MyAdapter) Name() string { return "my-adapter" }

func (a *MyAdapter) Setup(ctx context.Context, worktreeDir string) error {
    // Prepare the worktree (copy agent files, configs, etc.)
    return nil
}

func (a *MyAdapter) Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error) {
    // Execute your workflow and return results
    start := time.Now()
    // ...
    return &RunOutput{
        ExitCode: 0,
        WallTime: time.Since(start),
    }, nil
}
```

2. Register in `adapter.go`:

```go
var Registry = map[string]func(cfg map[string]any) (Adapter, error){
    "vanilla":    NewVanilla,
    "my-adapter": NewMyAdapter,
}
```

3. Add to `bench.yaml`:

```yaml
workflows:
  my-workflow:
    adapter: my-adapter
```

## Adding a New Task Type

All task types currently use the same generic verify template. To add a type-specific template:

1. Create a verify template at `internal/engine/templates/<type>.sh.tmpl`
2. Embed it in `verify.go`:

```go
//go:embed templates/my_type.sh.tmpl
var myTypeTemplate string
```

3. Add the case in `GenerateVerifyDir`:

```go
switch cfg.TaskType {
case "http-server":
    tmplStr = httpServerTemplate
case "my-type":
    tmplStr = myTypeTemplate
}
```

4. The template receives the worktree path as `$1` and must output a line:
```
BENCH_RESULT: L1=PASS L2=N/M L3=K L4=N/M
```

## Adding New Tasks

See [docs/tasks.md](tasks.md) for the full task authoring guide. Quick checklist:

1. Create directory: `tasks/tier{N}/{task-name}/`
2. Write `task.yaml` with all required fields
3. Write `plan.md` with clear instructions
4. Prepare `repo/` as a compilable git repository
5. Write `verify/e2e_test.go.src` with ground-truth tests
6. Validate: `workflow-bench validate --tasks tier{N}/{task-name} -v`

## Debugging

### Verbose Logging

```bash
workflow-bench run --tasks tier1 --tag debug -v
```

The `-v` flag enables debug-level structured logging via `slog`.

### Inspecting Results

Results are stored in SQLite at `~/.claude/workflow-bench/results.db`. You can query directly:

```bash
sqlite3 ~/.claude/workflow-bench/results.db "SELECT task_id, status, correctness_score FROM runs WHERE tag='my-tag'"
```

### Verify Script

The verify script is generated to a temp directory and deleted after each run. To inspect it, add a debug print or temporarily comment out the `defer os.RemoveAll(verifyDir)` in `runner.go`.
