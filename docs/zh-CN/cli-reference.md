# CLI 命令参考

## 全局选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--config` | string | `~/.claude/workflow-bench/bench.yaml` | 配置文件路径 |
| `-v, --verbose` | bool | false | 启用详细（debug）日志 |

---

## workflow-bench run

**说明**：使用指定 workflow adapter 对选定任务执行基准测试。

**用法**：`workflow-bench run --tasks <selector> --tag <tag> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--workflow` | string | `vanilla` | Workflow adapter：`vanilla` 或自定义名称 |
| `--tasks` | string | *（必填）* | 任务选择器：`tier1`、`tier1/fix-handler-bug` 或 `all` |
| `--tag` | string | *（必填）* | 标记本次运行的 tag |
| `--runs` | int | 取配置值（默认 3） | 每个任务的运行次数 |
| `--plan` | string | | 计划文件覆盖路径（替换任务自带的 plan.md） |
| `--tasks-dir` | string | `tasks` | 任务目录路径 |

**示例**：

```bash
# 用 vanilla workflow 运行所有 tier 1 任务
workflow-bench run --workflow vanilla --tasks tier1 --runs 3 --tag baseline

# 用自定义 workflow 运行（在 bench.yaml 中定义）
workflow-bench run --workflow my-workflow --tasks tier1 --runs 1 --tag custom-test

# 运行单个任务
workflow-bench run --tasks tier1/fix-handler-bug --runs 1 --tag quick-test

# 使用自定义计划
workflow-bench run --tasks tier1/fix-handler-bug --plan ./my-plan.md --tag custom-plan
```

**备注**：
- 每次运行会创建独立的 git worktree，不会修改原始任务仓库。
- 如果相同的 `(tag, workflow, task_id, run_number)` 已完成，会自动跳过（断点续跑）。
- 超时时间 = `estimated_minutes * timeout_multiplier`（默认 3 倍）。
- `--plan` 仅覆盖当前运行的计划，不修改磁盘上的文件。

---

## workflow-bench report

**说明**：为指定 tag 的基准测试运行生成 Markdown 汇总报告。

**用法**：`workflow-bench report --tag <tag> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tag` | string | *（必填）* | 要生成报告的 tag |
| `--format` | string | `markdown` | 输出格式（`markdown` 或 `md`） |

**示例**：

```bash
# 输出到终端
workflow-bench report --tag baseline

# 保存到文件
workflow-bench report --tag baseline > report.md
```

**备注**：
- 输出包含总体通过率、平均正确性分数和逐任务 L1-L4 明细。
- 同一任务有多次运行时，结果按任务分组并汇总通过次数。

---

## workflow-bench list tasks

**说明**：列出从任务目录发现的所有可用任务。

**用法**：`workflow-bench list tasks [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tasks-dir` | string | `tasks` | 任务目录路径 |

**示例**：

```bash
workflow-bench list tasks
```

**输出**：

```
ID                             TIER  TYPE            EST
tier1/fix-handler-bug          T1    http-server     5m
tier1/add-health-check         T1    http-server     5m

2 tasks found
```

---

## workflow-bench list workflows

**说明**：列出可用的 workflow adapter。

**用法**：`workflow-bench list workflows`

**示例**：

```bash
workflow-bench list workflows
```

**输出**：

```
Available workflows:
  vanilla    - Claude CLI direct execution
  custom     - User-defined command execution
```

---

## workflow-bench list tags

**说明**：列出所有 tag 及其运行次数、日期和 workflow。

**用法**：`workflow-bench list tags`

**示例**：

```bash
workflow-bench list tags
```

**输出**：

```
TAG                       RUNS   DATE                 WORKFLOW
baseline                  6      2026-03-30 14:00     vanilla
experiment-1                   3      2026-03-31 10:30     vanilla
```

---

## workflow-bench validate

**说明**：验证任务定义的正确性。检查 task.yaml 字段、repo 结构、构建、plan.md 和 E2E 测试文件。

**用法**：`workflow-bench validate [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tasks` | string | `all` | 要验证的任务选择器 |
| `--tasks-dir` | string | `tasks` | 任务目录路径 |

**验证项目**：
- `id` 非空
- `tier` 为 1-4
- `type` 非空
- `repo/` 目录存在
- `repo/go.mod` 存在
- `go build ./...` 在 repo/ 中通过（30 秒超时）
- `plan.md` 存在且非空
- `verify/e2e_test.go` 或 `verify/e2e_test.go.src` 存在
- `estimated_minutes` > 0

**示例**：

```bash
# 验证所有任务
workflow-bench validate

# 仅验证 tier 1 任务，启用详细输出
workflow-bench validate --tasks tier1 -v
```

**备注**：
- 使用 `-v`（详细模式）时，每项检查会输出 OK/FAIL 及详情。
- 任何任务验证失败时退出码非零。

---

## workflow-bench init

**说明**：初始化配置目录 `~/.claude/workflow-bench/` 并创建默认 `bench.yaml`（如不存在）。

**用法**：`workflow-bench init`

**示例**：

```bash
workflow-bench init
```

**输出**：

```
Initialized: /Users/you/.claude/workflow-bench
Config: /Users/you/.claude/workflow-bench/bench.yaml
```

**备注**：
- 可多次运行。如果 `bench.yaml` 已存在，不会覆盖。

---

## workflow-bench version

**说明**：打印版本号。

**用法**：`workflow-bench version`

**示例**：

```bash
workflow-bench version
```

**输出**：

```
workflow-bench 0.1.0
```
