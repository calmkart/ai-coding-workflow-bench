[English](README.md) | 中文

# workflow-bench

[![CI](https://github.com/calmkart/ai-coding-workflow-bench/actions/workflows/ci.yml/badge.svg)](https://github.com/calmkart/ai-coding-workflow-bench/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

多智能体编码工作流的基准测试工具。给定相同的计划，不同的工作流会产出不同的代码——workflow-bench 衡量哪种工作流做得更好。

## 特性

- **确定性四层验证**：构建、单元测试、静态分析、端到端测试
- **四维复合评分**：正确性、效率、质量、稳定性
- **LLM Judge**：基于 Anthropic API 的 Rubric 代码质量评分（7 个维度）
- **成对比较**：代码头对头比较，支持位置偏差检测
- **多 adapter 支持**：对比原生 Claude CLI、多智能体工作流或自定义命令
- **内置任务库**：4 个难度等级的 Go 编码任务
- **隔离执行**：每次运行使用独立 git worktree，互不干扰
- **断点续跑**：中断的运行可以自动恢复
- **并行执行**：通过 `--parallel` 并发运行多个任务
- **分片执行**：通过 `--shard` 分布到多台机器
- **Markdown 和 HTML 报告**：自动生成汇总、对比和趋势图
- **趋势追踪**：查看多个 tag 间的指标变化趋势
- **数据管理**：导出（JSON/CSV）、从 git 历史导入、合并分片结果、清理旧数据
- **SQLite 存储**：所有结果本地持久化，支持查询和对比
- **环境诊断**：`doctor` 命令验证前置条件

## 快速开始

### 前置条件

- Go 1.23+
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli) 已安装并配置
- 设置 `ANTHROPIC_API_KEY` 环境变量

### 安装

```bash
go install github.com/calmkart/ai-coding-workflow-bench/cmd/workflow-bench@latest
```

### 从源码构建

```bash
git clone https://github.com/calmkart/ai-coding-workflow-bench.git
cd ai-coding-workflow-bench

# 构建
go build -o workflow-bench ./cmd/workflow-bench

# 初始化配置目录 (~/.claude/workflow-bench/)
./workflow-bench init

# 验证内置任务
./workflow-bench validate --tasks tier1

# 运行基准测试（vanilla：直接调用 Claude CLI）
./workflow-bench run --workflow vanilla --tasks tier1 --runs 1 --tag my-first-run

# 查看结果
./workflow-bench report --tag my-first-run
```

## 架构

### 包结构（10 个包）

```
workflow-bench/
├── cmd/workflow-bench/     CLI 入口 (cobra)
├── internal/
│   ├── config/             bench.yaml 加载 + 任务发现
│   ├── engine/
│   │   ├── runner.go       编排：加载任务 -> adapter.Run -> 验证 -> 存储
│   │   ├── isolation.go    Git worktree 创建/清理
│   │   ├── verify.go       从嵌入模板生成 verify.sh
│   │   └── collector.go    解析 BENCH_RESULT 输出为 L1-L4 分数
│   ├── adapter/
│   │   ├── adapter.go      Adapter 接口 + 注册表
│   │   ├── vanilla.go      Claude CLI 直接执行
│   │   └── custom.go       用户自定义命令执行
│   ├── judge/              LLM Judge：Rubric 评分 + 成对比较
│   │   ├── rubric.go       7 维 Rubric 评估（通过 Anthropic API）
│   │   └── pairwise.go     代码头对头比较（含位置偏差检测）
│   ├── metrics/
│   │   ├── correctness.go  正确性评分公式
│   │   ├── cost.go         统一成本估算
│   │   └── statistics.go   Wilson CI、显著性检验、稳定性评分
│   ├── store/
│   │   ├── db.go           SQLite CRUD (纯 Go，无 CGO)
│   │   └── schema.sql      数据库 schema
│   ├── report/
│   │   ├── summary.go      Markdown/HTML 汇总报告
│   │   ├── compare.go      并排对比报告
│   │   ├── trend.go        多 tag 趋势报告
│   │   ├── export.go       JSON/CSV 数据导出
│   │   └── html.go         HTML 报告渲染
│   ├── importer/           从 git 历史导入任务
│   └── taskgen/            任务变体生成
└── tasks/                  内置任务库（100 个任务）
    ├── tier1/              20 个简单任务（~5 分钟）
    ├── tier2/              32 个中等任务（~10 分钟）
    ├── tier3/              29 个复杂任务（~15-20 分钟）
    └── tier4/              19 个高级任务（~25-30 分钟）
```

### 数据流

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

1. **runner** 发现任务，为每次运行创建隔离的 git worktree
2. **adapter** 执行工作流（如通过 `claude -p` 传递计划）
3. **verify.sh** 对修改后的 worktree 运行 L1-L4 检查
4. **collector** 将验证输出解析为结构化分数
5. 结果存入 SQLite；**report** 渲染为 Markdown

## CLI 命令

| 命令 | 说明 |
|------|------|
| `run` | 运行基准测试（`--parallel`、`--shard`、`--keep-worktree`、`--pairwise`） |
| `report` | 生成汇总报告（`--format markdown\|html`） |
| `compare` | 并排对比两个 tag 的结果（`--pairwise` 启用 LLM 比较） |
| `trend` | 显示多个 tag 的指标趋势（`--tags v1,v2,v3`） |
| `export` | 导出原始数据为 JSON 或 CSV |
| `inspect` | 查看特定运行的原始输出（verify.log、diff.patch） |
| `import` | 从 git 提交历史创建任务 |
| `generate-variant` | 生成重命名标识符的任务变体 |
| `merge` | 合并多个结果数据库（用于分片执行） |
| `clean` | 按 tag、按时间删除运行或清理孤立 worktree |
| `doctor` | 检查环境前置条件 |
| `list tasks` | 列出所有可用任务 |
| `list workflows` | 列出可用的 workflow adapter |
| `list tags` | 列出所有 tag 及运行次数 |
| `validate` | 验证任务定义（结构、构建、测试） |
| `init` | 初始化配置目录和默认 bench.yaml |
| `version` | 打印版本号 |

完整命令参考见 [docs/zh-CN/cli-reference.md](docs/zh-CN/cli-reference.md)。

## 内置任务

100 个任务，覆盖 4 个难度等级和 5 种代码类型：

| 等级 | 数量 | 难度 | 预估时间 |
|------|------|------|----------|
| T1   | 20   | 简单 | ~5 分钟   |
| T2   | 32   | 中等 | ~10 分钟  |
| T3   | 29   | 复杂 | ~15-20 分钟|
| T4   | 19   | 高级 | ~25-30 分钟|

| 类型 | 数量 | 示例 |
|------|------|------|
| http-server | 32 | CRUD 修复、中间件、认证、限流、RBAC |
| library | 24 | 字符串工具、LRU 缓存、熔断器、B 树 |
| cli | 15 | 参数解析、配置加载、交互模式 |
| concurrency | 15 | 工作池、发布/订阅、Actor 模型、MapReduce |
| reconciler | 14 | 状态机、Finalizer、Leader 选举、GC |

运行 `workflow-bench list tasks` 查看完整列表。任务编写指南见 [docs/zh-CN/tasks.md](docs/zh-CN/tasks.md)。

## 评分

### 四层验证 (L1-L4)

| 层级 | 内容 | 权重 |
|------|------|------|
| L1 构建 | `go build ./...` | 门控（失败则总分为 0） |
| L2 单元测试 | `go test -race ./...` | 20% |
| L3 静态分析 | `go vet` + 可选 `staticcheck` | 10% |
| L4 端到端测试 | 基于 `httptest` 的 E2E 测试 | 70% |

### 正确性公式

```
if L1 == FAIL:
    correctness = 0.0
else:
    l2 = passed / total
    l3 = max(0, 1.0 - issues * 0.05)
    l4 = passed / total
    correctness = 0.20 * l2 + 0.10 * l3 + 0.70 * l4
```

关键 verification target (VT) 失败会额外扣除 0.1 分（下限为 0）。

详见 [docs/zh-CN/scoring.md](docs/zh-CN/scoring.md)。

## 配置

配置文件位于 `~/.claude/workflow-bench/bench.yaml`（由 `init` 创建）。

```yaml
workflows:
  vanilla:
    adapter: vanilla

  # 示例：使用 custom adapter 配置多智能体工作流
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

完整配置参考见 [docs/zh-CN/configuration.md](docs/zh-CN/configuration.md)，包含两种 adapter（`vanilla`、`custom`）的详细说明。

## 开发

```bash
# 构建
make build

# 运行测试
make test

# 清理
make clean
```

### 自定义 Workflow（无需编写 Go 代码）

使用 `custom` adapter 配置任意命令作为 workflow：

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

详见 [docs/zh-CN/configuration.md](docs/zh-CN/configuration.md) 获取 custom adapter 完整参考。

### 添加新 Adapter（Go）

1. 在 `internal/adapter/myadapter.go` 中实现 `Adapter` 接口
2. 在 `adapter.go` 的 `adapter.Registry` 中注册
3. 在 `bench.yaml` 中添加 workflow 条目

详见 [docs/zh-CN/development.md](docs/zh-CN/development.md)。

## 路线图

| 阶段 | 范围 | 状态 |
|------|------|------|
| **v1 (P1-P6)** | Wilson CI、VT 检测、数据导出、inspect/doctor、四维复合评分 | 已完成 |
| **v2 (P7-P12)** | verify.sh JSON、统一成本、VT 映射、Judge 超时、稳定性评分 | 已完成 |
| **v2 (P13-P18)** | LLM Judge Rubric、HTML 报告、对比增强、成对比较 | 已完成 |
| **v2 (P19-P22)** | 任务导入、变体生成、分片执行、数据库合并 | 已完成 |
| **v2 (P23-P25)** | 趋势报告、clean 命令、文档同步 | 已完成 |

## 许可证

MIT
