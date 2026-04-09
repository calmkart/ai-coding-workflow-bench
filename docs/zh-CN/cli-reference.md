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
| `--workflow` | string | `vanilla` | Workflow adapter：`vanilla`（使用 `--bare` 纯模型基线）、`custom` 或自定义名称 |
| `--tasks` | string | *（必填）* | 任务选择器：`tier1`、`tier1/fix-handler-bug` 或 `all` |
| `--tag` | string | *（必填）* | 标记本次运行的 tag |
| `--runs` | int | 取配置值（默认 3） | 每个任务的运行次数 |
| `--plan` | string | | 计划文件覆盖路径（替换任务自带的 plan.md） |
| `--tasks-dir` | string | `tasks` | 任务目录路径 |
| `--parallel` | int | `1` | 并行运行的任务数 |
| `--keep-worktree` | bool | false | 运行后不删除 worktree（调试用） |
| `--shard` | string | | 分片索引/总数（如 `1/4`）用于分布式执行 |

**示例**：

```bash
# 使用 vanilla workflow 运行所有 tier 1 任务
workflow-bench run --workflow vanilla --tasks tier1 --runs 3 --tag baseline

# 使用多智能体 workflow（在 bench.yaml 中配置为 custom adapter）
workflow-bench run --workflow multi-agent --tasks tier1 --runs 1 --tag multi-agent

# 并行运行（同时 4 个任务）
workflow-bench run --tasks all --runs 1 --tag fast --parallel 4

# 分布式执行：第 1 个分片（共 4 个）
workflow-bench run --tasks all --runs 1 --tag distributed --shard 1/4

# 保留 worktree 用于调试
workflow-bench run --tasks tier1/fix-handler-bug --runs 1 --tag debug --keep-worktree
```

**注意**：
- 每次运行创建独立的 git worktree，原始任务仓库不会被修改。
- 如果相同 `(tag, workflow, task_id, run_number)` 的运行已完成，会自动跳过（断点续跑）。
- 每个任务的超时时间 = `estimated_minutes * timeout_multiplier`（默认 3 倍）。
- `--shard` 将任务列表均分；使用 `merge` 合并结果。

---

## workflow-bench report

**说明**：为指定 tag 的运行生成汇总报告。

**用法**：`workflow-bench report --tag <tag> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tag` | string | *（必填）* | 要生成报告的 tag |
| `--format` | string | `markdown` | 输出格式：`markdown`、`md` 或 `html` |

**示例**：

```bash
# 输出 Markdown 报告到 stdout
workflow-bench report --tag baseline

# 生成 HTML 报告
workflow-bench report --tag baseline --format html > report.html
```

---

## workflow-bench compare

**说明**：并排对比两个 tag 的基准测试结果。

**用法**：`workflow-bench compare --left <tag> --right <tag> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--left` | string | *（必填）* | 左侧 tag |
| `--right` | string | *（必填）* | 右侧 tag |
| `--format` | string | `markdown` | 输出格式：`markdown`、`md` 或 `html` |
| `--pairwise` | bool | false | 启用 LLM 成对比较（需配置 judge） |

**示例**：

```bash
# 对比 vanilla vs 多智能体
workflow-bench compare --left vanilla-tag --right v4-tag

# 带 LLM 成对比较
workflow-bench compare --left vanilla-tag --right v4-tag --pairwise

# 生成 HTML 对比报告
workflow-bench compare --left v1 --right v2 --format html > compare.html
```

---

## workflow-bench trend

**说明**：显示多个 tag 的指标变化趋势。

**用法**：`workflow-bench trend --tags <tag1,tag2,...> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tags` | string | *（必填）* | 逗号分隔的 tag 列表 |
| `--format` | string | `markdown` | 输出格式：`markdown`、`md` 或 `html` |

**示例**：

```bash
# 查看三个版本的趋势
workflow-bench trend --tags v1,v2,v3

# 生成 HTML 趋势图
workflow-bench trend --tags v1,v2,v3 --format html > trend.html
```

---

## workflow-bench export

**说明**：导出基准测试数据为 JSON 或 CSV。

**用法**：`workflow-bench export --tag <tag> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tag` | string | *（必填）* | 要导出的 tag |
| `--format` | string | `json` | 输出格式：`json` 或 `csv` |

---

## workflow-bench inspect

**说明**：查看特定运行的原始输出。

**用法**：`workflow-bench inspect --run-id <id>`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--run-id` | string | *（必填）* | 要查看的运行 ID |

---

## workflow-bench import

**说明**：从 git 提交历史创建任务。

**用法**：`workflow-bench import --repo <path> --commit <range> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--repo` | string | *（必填）* | Git 仓库路径 |
| `--commit` | string | *（必填）* | 提交范围（如 `abc123..def456`） |
| `--tier` | int | `0` | 覆盖难度等级（0 = 自动检测） |
| `--type` | string | | 覆盖任务类型（空 = 自动检测） |
| `--output` | string | | 输出目录（默认：`tasks/imported/<name>/`） |

---

## workflow-bench generate-variant

**说明**：生成重命名标识符的任务变体。

**用法**：`workflow-bench generate-variant --source <path> [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--source` | string | *（必填）* | 源任务目录路径 |
| `--output` | string | `<source>-variant` | 输出目录 |
| `--seed` | int64 | `0` | 随机种子（0 = 随机） |

---

## workflow-bench merge

**说明**：合并多个结果数据库（用于分片执行）。

**用法**：`workflow-bench merge --from <db1> --from <db2> --to <target>`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--from` | string[] | *（必填）* | 源数据库路径（可重复） |
| `--to` | string | *（必填）* | 目标数据库路径 |

---

## workflow-bench clean

**说明**：清理基准测试数据或孤立 worktree。

**用法**：`workflow-bench clean [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tag` | string | | 删除此 tag 的所有运行 |
| `--older-than` | string | | 删除早于此时间段的运行（如 `30d`、`24h`） |
| `--worktrees` | bool | false | 清理系统临时目录中的孤立 worktree 目录 |

**示例**：

```bash
# 删除特定 tag 的数据
workflow-bench clean --tag old-test

# 删除 30 天前的数据
workflow-bench clean --older-than 30d

# 清理孤立 worktree
workflow-bench clean --worktrees
```

---

## workflow-bench doctor

**说明**：检查环境前置条件。

**用法**：`workflow-bench doctor [flags]`

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tasks-dir` | string | `tasks` | 任务目录路径 |

---

## workflow-bench list tasks / workflows / tags

**说明**：列出可用的任务、workflow 或 tag。

---

## workflow-bench validate

**说明**：验证任务定义的正确性。

**选项**：

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--tasks` | string | `all` | 要验证的任务选择器 |
| `--tasks-dir` | string | `tasks` | 任务目录路径 |

---

## workflow-bench init

**说明**：初始化配置目录并创建默认 `bench.yaml`。

---

## workflow-bench version

**说明**：打印版本号。
