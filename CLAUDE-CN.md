# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指引。

## 构建与测试命令

```bash
make build                          # 构建二进制文件 -> ./workflow-bench
make test                           # 运行所有测试（带竞态检测）
make clean                          # 删除二进制文件并清除测试缓存

go test ./internal/metrics/... -v   # 运行单个包的测试
go test ./... -count=1 -race        # 等同于 make test
```

无需 CGO —— SQLite 驱动 (`modernc.org/sqlite`) 是纯 Go 实现。

## 架构

这是一个 Go CLI 工具（基于 cobra），用于评测多智能体编码工作流。它将 AI 编码智能体在预设的 Go 任务上运行，然后通过四层验证（构建、单元测试、静态分析、端到端测试）检查输出并评分。

### 执行流水线

`runner.go` 负责编排：**发现任务 -> 创建 git worktree -> adapter.Run -> verify.sh -> 解析 BENCH_RESULT -> 评分 -> 存入 SQLite**

每次评测运行都会获得一个隔离的 git worktree（从 `tasks/tier*/*/repo/` 创建），确保运行之间互不污染。运行结束后 worktree 会被清理。

### 核心接口

- **Adapter**（`internal/adapter/adapter.go`）：`Setup(ctx, worktreeDir)` + `Run(ctx, worktreeDir, planContent) -> RunOutput`。内置两种适配器：`vanilla`（claude -p）、`custom`（用户自定义 shell 命令）。通过 `adapter.Registry` 注册。

- **验证输出协议**：verify.sh 必须输出 `BENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5` —— 由 `collector.go` 通过正则解析。

### 正确性评分

L1（构建）是门控 —— 如果失败则得分为 0。否则：`0.20 * L2 + 0.10 * L3 + 0.70 * L4`。关键 VT 失败每次扣 0.1（最低为 0）。

### 任务结构

每个任务位于 `tasks/tier{1-4}/{task-name}/`，包含：
- `task.yaml` —— 元数据（id、tier、type、estimated_minutes）
- `plan.md` —— 提供给 AI 智能体的计划
- `repo/` —— 可编译的 Go 项目（必须是 git 仓库，用于创建 worktree）
- `verify/e2e_test.go.src` —— 真值端到端测试（`.src` 扩展名避免 Go 工具链在原地编译它）

任务通过 `filepath.Glob("tasks/tier*/*/task.yaml")` 发现 —— 无需索引文件。

### 嵌入模板

验证脚本模板和数据库 schema 通过 `//go:embed` 嵌入。修改 `internal/engine/templates/*.sh.tmpl` 或 `internal/store/schema.sql` 后需要重新构建。

### 配置与存储

- 配置文件：`~/.claude/workflow-bench/bench.yaml`（通过 `workflow-bench init` 创建）
- 结果数据库：`~/.claude/workflow-bench/results.db`（SQLite，WAL 模式）
- 指定 `--config` 时，数据库路径从配置文件所在目录派生（实现按配置位置完全隔离）

## 路线图背景

P1-P2（CLI、适配器、100 个任务、L1-L4 验证）已完成。P3+（LLM 评审、配对排名、稳定性评分）已实现。
