# Task Authoring Guide

## Overview

Tasks are self-contained coding challenges that workflow-bench uses to evaluate workflow strategies. Each task includes a starting codebase (repo), a plan describing what to do, and E2E tests that verify the result.

## Directory Structure

Tasks live under `tasks/` organized by tier:

```
tasks/
├── tier1/
│   └── fix-handler-bug/
│       ├── task.yaml              # Task metadata and verification targets
│       ├── plan.md                # The plan given to the workflow
│       ├── repo/                  # Starting codebase (a git repository)
│       │   ├── go.mod
│       │   ├── main.go
│       │   ├── handlers.go
│       │   └── handlers_test.go
│       └── verify/
│           └── e2e_test.go.src    # Ground-truth E2E tests
├── tier2/
├── tier3/
└── tier4/
```

## Creating a New Task

### Step 1: Create the Directory

```bash
mkdir -p tasks/tier2/my-new-task/{repo,verify}
```

### Step 2: Write task.yaml

```yaml
id: "tier2/my-new-task"
name: "Short description of the task"
tier: 2
type: "http-server"
language: "go"
estimated_minutes: 10

api_contract:
  - "func setupRouter() http.Handler"

verification_targets:
  - id: VT-ERROR-01
    category: error_handling
    name: "Error swallowing"
    severity: high
    description: "Agent may introduce silent error drops during refactoring"
    detection: "errcheck"

refactoring_targets:
  - "Extract storage interface from handler"

code_smells:
  - "Direct database access in HTTP handler"

metadata:
  files_to_modify: ["handlers.go"]
  tags: ["refactoring", "interface-extraction"]
```

### Step 3: Write plan.md

The plan is what the workflow adapter gives to the coding agent. Write it as clear instructions:

```markdown
# Task: Extract Storage Interface

## Goal
Separate the storage logic from HTTP handlers by introducing a Storage interface.

## Requirements
- REQ-1: Define a `Storage` interface with `List`, `Get`, `Create`, `Delete` methods
- REQ-2: Move in-memory storage into a concrete `MemoryStorage` struct
- REQ-3: Inject the storage into handlers via constructor

## Constraints
- setupRouter() function signature must not change
- All existing tests must continue to pass

## Do Not
- Add external dependencies
- Change API endpoints or response formats
```

### Step 4: Prepare the Repo

The repo is the starting codebase that the workflow operates on. It must be a valid git repository.

```bash
cd tasks/tier2/my-new-task/repo
go mod init example.com/my-task
# Create your source files with intentional code smells...
git init && git add . && git commit -m "initial"
```

Requirements for the repo:
- Must contain `go.mod`
- Must compile (`go build ./...`)
- For `http-server` type: must export `func setupRouter() http.Handler` in `main.go`
- Should contain basic unit tests in `*_test.go` files

### Step 5: Write E2E Tests

Create `verify/e2e_test.go.src` (note the `.src` extension -- this prevents Go tooling from trying to compile it in the task directory).

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestBenchE2E(t *testing.T) {
    router := setupRouter()
    srv := httptest.NewServer(router)
    defer srv.Close()

    t.Run("storage_interface_exists", func(t *testing.T) {
        // Verify the refactoring result...
    })
}
```

The E2E test file:
- Must be in package `main` (it is copied into the repo root during verification)
- Must use `TestBenchE2E` as the test function name (or `TestBenchE2E_*` subtests)
- Must call `setupRouter()` to get the HTTP handler (for http-server type)
- Is the ground truth -- the workflow has no access to this file

### Step 6: Validate

```bash
workflow-bench validate --tasks tier2/my-new-task -v
```

## task.yaml Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Unique identifier, format: `tier{N}/{name}` |
| `name` | string | yes | Human-readable task name |
| `tier` | int | yes | Difficulty tier (1-4) |
| `type` | string | yes | Task type: `http-server`, `k8s-operator`, `library`, `cli` |
| `language` | string | yes | Programming language (currently `go`) |
| `estimated_minutes` | int | yes | Expected completion time, used for timeout calculation |
| `api_contract` | list[string] | no | Public function signatures the agent must not break |
| `verification_targets` | list[VT] | no | Known pitfalls to check for |
| `refactoring_targets` | list[string] | no | What the agent should refactor |
| `code_smells` | list[string] | no | Intentional issues in the starting code |
| `metadata.files_to_modify` | list[string] | no | Expected files to change |
| `metadata.tags` | list[string] | no | Categorization tags |

### Verification Target (VT) Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | e.g., `VT-ERROR-01` |
| `category` | string | One of: concurrency, error_handling, memory, interface, package, http, distributed, k8s, test |
| `name` | string | Short name |
| `severity` | string | `critical` or `high` or `medium` |
| `description` | string | What can go wrong |
| `detection` | string | How it is detected: `errcheck`, `goleak`, `go build`, `e2e test case`, etc. |

## Task Types

### http-server

Standard HTTP server using Go stdlib `net/http`. The repo must export `setupRouter() http.Handler` for E2E tests to use. No external dependencies required.

### k8s-operator (planned)

Kubernetes operator using `controller-runtime`. E2E tests use `envtest` (real apiserver + etcd). The `metadata.envtest_k8s_version` field specifies the K8s version.

### library (planned)

Go library package. E2E tests import and call the package directly.

### cli (planned)

Command-line tool. E2E tests execute the binary and check output.

## Tier Guidelines

| Tier | Complexity | Est. Time | Example |
|------|-----------|-----------|---------|
| T1 | Single-file bug fix or feature addition | 5 min | Fix off-by-one, add endpoint |
| T2 | Interface extraction, simple refactoring | 10 min | Extract storage interface, add middleware |
| T3 | Multi-file architectural refactoring | 15-20 min | Service layer separation, K8s operator cleanup |
| T4 | Complex cross-cutting changes | 25-30 min | Auth middleware, concurrent fanout, full operator refactor |

## Built-in Tasks

| ID | Tier | Type | Description |
|----|------|------|-------------|
| tier1/fix-handler-bug | T1 | http-server | Fix pagination off-by-one in GET /todos |
| tier1/add-health-check | T1 | http-server | Add GET /health endpoint |
