# Contributing to workflow-bench

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

```bash
# Clone the repository
git clone https://github.com/calmkart/ai-coding-workflow-bench.git
cd ai-coding-workflow-bench

# Build
make build

# Run tests
make test

# Run all checks (fmt, vet, lint, test)
make check
```

### Prerequisites

- Go 1.23+
- git
- (Optional) [staticcheck](https://staticcheck.dev/) for linting: `go install honnef.co/go/tools/cmd/staticcheck@latest`

## How to Contribute

### Reporting Bugs

Open an issue using the **Bug Report** template. Include:
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS

### Suggesting Features

Open an issue using the **Feature Request** template.

### Submitting Code

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run checks: `make check`
5. Commit with a clear message
6. Open a Pull Request

### Adding New Tasks

See [docs/tasks.md](docs/tasks.md) for the task authoring guide. Quick checklist:

1. Create `tasks/tier{N}/{task-name}/`
2. Write `task.yaml`, `plan.md`, `repo/` (compilable Go project), `verify/e2e_test.go.src`
3. Validate: `workflow-bench validate --tasks tier{N}/{task-name} -v`

### Adding a New Adapter

For simple cases, use the `custom` adapter (YAML-only, no Go code needed).

For complex cases, implement the `Adapter` interface in `internal/adapter/` and register it in `adapter.Registry`. See [docs/development.md](docs/development.md).

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Wrap errors with context: `fmt.Errorf("load config: %w", err)`
- Public functions must have godoc comments
- No CGO dependencies

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
