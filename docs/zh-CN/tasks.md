# 任务编写指南

## 概述

任务是自包含的编码挑战，workflow-bench 用它来评估不同工作流策略的效果。每个任务包含一个初始代码库（repo）、一份描述任务目标的计划（plan），以及用于验证结果的 E2E 测试。

## 目录结构

任务按难度等级组织在 `tasks/` 下：

```
tasks/
├── tier1/
│   └── fix-handler-bug/
│       ├── task.yaml              # 任务元数据和 verification target
│       ├── plan.md                # 提供给 workflow 的计划
│       ├── repo/                  # 初始代码库（git 仓库）
│       │   ├── go.mod
│       │   ├── main.go
│       │   ├── handlers.go
│       │   └── handlers_test.go
│       └── verify/
│           └── e2e_test.go.src    # E2E 测试（真值）
├── tier2/
├── tier3/
└── tier4/
```

## 创建新任务

### 第一步：创建目录

```bash
mkdir -p tasks/tier2/my-new-task/{repo,verify}
```

### 第二步：编写 task.yaml

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

### 第三步：编写 plan.md

plan 是 workflow adapter 提供给编码智能体的指令。用清晰的语言编写：

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

### 第四步：准备 repo

repo 是 workflow 操作的初始代码库，必须是有效的 git 仓库。

```bash
cd tasks/tier2/my-new-task/repo
go mod init example.com/my-task
# 编写带有刻意代码异味的源文件...
git init && git add . && git commit -m "initial"
```

repo 要求：
- 必须包含 `go.mod`
- 必须可编译（`go build ./...`）
- `http-server` 类型：必须在 `main.go` 中导出 `func setupRouter() http.Handler`
- 建议包含基础单元测试（`*_test.go`）

### 第五步：编写 E2E 测试

创建 `verify/e2e_test.go.src`（注意 `.src` 扩展名——防止 Go 工具链在任务目录中尝试编译它）。

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
        // 验证重构结果...
    })
}
```

E2E 测试文件要求：
- 必须属于 `main` 包（验证时会被复制到 repo 根目录）
- 测试函数名必须为 `TestBenchE2E`（或 `TestBenchE2E_*` 子测试）
- `http-server` 类型必须调用 `setupRouter()` 获取 HTTP handler
- 该文件是真值——workflow 无法访问它

### 第六步：验证

```bash
workflow-bench validate --tasks tier2/my-new-task -v
```

## task.yaml 字段参考

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 是 | 唯一标识符，格式：`tier{N}/{name}` |
| `name` | string | 是 | 可读的任务名称 |
| `tier` | int | 是 | 难度等级（1-4） |
| `type` | string | 是 | 任务类型：`http-server`、`k8s-operator`、`library`、`cli` |
| `language` | string | 是 | 编程语言（目前为 `go`） |
| `estimated_minutes` | int | 是 | 预估完成时间，用于计算超时 |
| `api_contract` | list[string] | 否 | 智能体不可破坏的公共函数签名 |
| `verification_targets` | list[VT] | 否 | 需要检查的已知陷阱 |
| `refactoring_targets` | list[string] | 否 | 智能体应执行的重构内容 |
| `code_smells` | list[string] | 否 | 初始代码中刻意留下的问题 |
| `metadata.files_to_modify` | list[string] | 否 | 预期修改的文件 |
| `metadata.tags` | list[string] | 否 | 分类标签 |

### Verification Target (VT) 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 如 `VT-ERROR-01` |
| `category` | string | 可选值：concurrency、error_handling、memory、interface、package、http、distributed、k8s、test |
| `name` | string | 简称 |
| `severity` | string | `critical`、`high` 或 `medium` |
| `description` | string | 可能出现的问题 |
| `detection` | string | 检测方式：`errcheck`、`goleak`、`go build`、`e2e test case` 等 |

## 任务类型

### http-server

使用 Go 标准库 `net/http` 的 HTTP 服务器。repo 必须导出 `setupRouter() http.Handler` 供 E2E 测试使用。无需外部依赖。

### k8s-operator（计划中）

使用 `controller-runtime` 的 Kubernetes operator。E2E 测试使用 `envtest`（真实 apiserver + etcd）。通过 `metadata.envtest_k8s_version` 指定 K8s 版本。

### library（计划中）

Go 库包。E2E 测试直接导入并调用该包。

### cli（计划中）

命令行工具。E2E 测试执行二进制文件并检查输出。

## 难度等级指南

| 等级 | 复杂度 | 预估时间 | 示例 |
|------|--------|----------|------|
| T1 | 单文件缺陷修复或功能添加 | 5 分钟 | 修复 off-by-one、添加端点 |
| T2 | 接口提取、简单重构 | 10 分钟 | 提取存储接口、添加中间件 |
| T3 | 多文件架构重构 | 15-20 分钟 | 服务层分离、K8s operator 清理 |
| T4 | 复杂的横切关注点变更 | 25-30 分钟 | 认证中间件、并发 fanout、完整 operator 重构 |

## 内置任务

| ID | 等级 | 类型 | 说明 |
|----|------|------|------|
| tier1/fix-handler-bug | T1 | http-server | 修复 GET /todos 的分页 off-by-one |
| tier1/add-health-check | T1 | http-server | 添加 GET /health 端点 |
