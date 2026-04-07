# 配置参考

## 概述

workflow-bench 使用 YAML 配置文件定义 workflow 和默认参数。运行 `workflow-bench init` 创建默认配置。

## 文件位置

| 路径 | 说明 |
|------|------|
| `~/.claude/workflow-bench/bench.yaml` | 默认配置文件 |
| `~/.claude/workflow-bench/results.db` | SQLite 数据库 |
| `~/.claude/workflow-bench/` | workflow-bench 所有数据的主目录 |

通过 `--config /path/to/bench.yaml` 指定其他配置文件。

如果配置文件不存在且未指定 `--config`，workflow-bench 使用内置默认值。

## bench.yaml 完整参考

```yaml
# Workflow 定义
workflows:
  vanilla:
    adapter: vanilla               # 直接调用 Claude CLI

  v4-claude:
    adapter: v4-claude             # 通过 --agent manager 的多智能体工作流
    agents_dir: "~/.claude/agents"

  my-workflow:
    adapter: custom                # 用户自定义命令
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"

# 默认运行参数
defaults:
  runs_per_task: 3                 # 每个任务的运行次数（可被 --runs 覆盖）
  timeout_multiplier: 3            # 超时 = estimated_minutes * 此倍数

  # P2+ 计划：
  # model: "claude-sonnet-4-20250514"
  # cost_budget:
  #   tier1: 0.50
  #   tier2: 1.00
  #   tier3: 2.00
  #   tier4: 5.00

# LLM Judge 设置（P3+ 计划）
# judge:
#   model: "claude-sonnet-4-20250514"
#   ensemble_model: "gpt-4o"
#   pairwise_mode: "compact"       # compact | full
#   enable_ensemble: false

# 导入设置（P6+ 计划）
# import:
#   time_window_hours: 72
#   jaccard_threshold: 0.3
```

## 字段参考

### workflows

workflow 名称到配置的映射。每个条目定义一个命名 workflow，通过 `--workflow` 选择。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `adapter` | string | 是 | 使用的 adapter 实现名称 |

当前可用 adapter：
- `vanilla` -- 直接使用计划内容运行 `claude -p`
- `v4-claude` -- 通过 `claude --agent manager` 运行多智能体工作流
- `custom` -- 用户自定义命令执行（详见下文）

### defaults

基准测试运行的默认参数。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `runs_per_task` | int | 3 | 每个任务的运行次数（每个 workflow）。可被 `--runs` 覆盖。 |
| `timeout_multiplier` | int | 3 | 乘以 task.yaml 的 `estimated_minutes` 得到单次运行超时。 |

### defaults.cost_budget（计划中）

按难度等级的成本预算，用于效率评分归一化。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `tier1` | float | 0.50 | T1 任务的最大预期 USD 成本 |
| `tier2` | float | 1.00 | T2 任务的最大预期 USD 成本 |
| `tier3` | float | 2.00 | T3 任务的最大预期 USD 成本 |
| `tier4` | float | 5.00 | T4 任务的最大预期 USD 成本 |

### judge（计划中）

LLM Judge 代码质量评分配置。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `model` | string | `claude-sonnet-4-20250514` | Rubric 和 Pairwise 评估使用的模型 |
| `ensemble_model` | string | `gpt-4o` | 集成评估的第二个模型 |
| `pairwise_mode` | string | `compact` | `compact`（14 次调用）或 `full`（最多 50 次调用） |
| `enable_ensemble` | bool | false | 为 T3-T4 任务启用多模型集成 |

## 环境变量

| 变量 | 必填 | 说明 |
|------|------|------|
| `ANTHROPIC_API_KEY` | 是 | Claude CLI 和 LLM Judge 的 API 密钥 |
| `OPENAI_API_KEY` | 否 | 仅在 `judge.enable_ensemble` 为 true 时需要 |
| `HOME` | 是 | 用于定位 `~/.claude/workflow-bench/` |

## 配置解析顺序

1. `--config` 显式指定
2. 默认路径：`~/.claude/workflow-bench/bench.yaml`
3. 内置默认值（无配置文件时）

配置值与默认值合并：如果你的配置只指定了 `workflows`，`defaults` 部分仍使用内置值。

## 何时使用哪个 Adapter

| Adapter | 适用场景 | 示例 |
|---------|----------|------|
| `vanilla` | 使用 Claude CLI 直接执行的基线测试 | `claude -p` 加计划内容 |
| `v4-claude` | 带 manager 编排的多智能体工作流 | `claude --agent manager` |
| `custom` | 任何其他工具、包装脚本或自定义配置 | Aider、Cursor、自定义脚本 |

建议先用 `vanilla` 作为基线，再与 `v4-claude` 或 `custom` 工作流对比。

## Custom Adapter（自定义命令）

`custom` adapter 允许你通过 bench.yaml 配置任意命令作为 workflow，无需编写 Go 代码：

```yaml
workflows:
  my-workflow:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"
      - "mkdir -p .planning/manager"
```

### Custom Adapter 字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `entry_command` | string | 是 | 主执行命令（通过 `bash -c` 运行） |
| `setup_commands` | string 列表 | 否 | 在 entry_command 之前按顺序执行的准备命令 |

### 环境变量（adapter 自动设置）

以下环境变量在 `entry_command` 中可用：

| 变量 | 说明 |
|------|------|
| `BENCH_REPO_DIR` | worktree 目录的绝对路径 |
| `BENCH_PLAN_FILE` | 计划文件的绝对路径（`.bench-plan.md`） |
| `BENCH_PLAN_PROMPT` | 便捷提示词：`"Read the plan from <BENCH_PLAN_FILE> and implement it."` |

### stdout JSON（可选）

如果 `entry_command` 输出包含 `usage` 字段的 JSON，则会提取 token 数据：

```json
{"usage": {"input_tokens": 100, "output_tokens": 50}, "tool_uses": 5}
```

如果输出不是有效 JSON 或缺少 `usage`，token 数据标记为 N/A。

### 示例

**示例 1：带自定义 agent 的 Claude CLI**

```yaml
workflows:
  my-agents:
    adapter: custom
    entry_command: |
      claude -p "$BENCH_PLAN_PROMPT" --output-format json
    setup_commands:
      - "cp -r ~/my-agents/ .claude/agents/"
      - "mkdir -p .planning/manager"
```

**示例 2：Shell 脚本包装**

```yaml
workflows:
  my-script:
    adapter: custom
    entry_command: |
      ~/scripts/run-coding-agent.sh "$BENCH_PLAN_FILE" "$BENCH_REPO_DIR"
```

```bash
# 使用自定义 workflow 运行
workflow-bench run --workflow my-agents --tasks tier1 --runs 1 --tag test
```

## 添加新 Adapter（Go）

通过 Go 代码添加 workflow adapter：

1. 在 `internal/adapter/` 中实现 `Adapter` 接口
2. 在 `adapter.Registry` 中注册
3. 在 `bench.yaml` 中添加 workflow 条目
4. 通过 `--workflow <name>` 运行
