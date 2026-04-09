# 评分系统

## 概述

workflow-bench 采用四维复合评分系统：

1. **正确性** (40%) -- 确定性 L1-L4 验证
2. **效率** (25%) -- Token 使用量和成本
3. **质量** (25%) -- LLM Judge Rubric 评分（7 个维度）
4. **稳定性** (10%) -- 多次运行的一致性

## 四层验证 (L1-L4)

每次运行都经过四层验证，由自动生成的 `verify.sh` 脚本执行。

### L1：构建

```bash
go build ./...
```

二进制门控。代码不能编译则 correctness = 0.0，后续检查不再运行。

### L2：单元测试

```bash
go test -json ./... -count=1 -race
```

使用竞态检测器运行任务的单元测试。使用 `go test -json` 精确计数（仅计算顶层测试函数，过滤子测试）。

### L3：静态分析

```bash
go vet ./...
staticcheck ./...   # 如已安装
gosec ./...          # 如已安装
```

统计 lint 问题数。每个问题从 L3 子分数中扣除 5%。

### L4：端到端测试

```bash
go test -json -run TestBenchE2E -count=1 -race ./...
```

运行任务 `verify/` 目录中的 E2E 测试。这些测试对工作流代理不可见，仅在验证阶段复制到 worktree。

## 正确性公式

```
if L1 == FAIL:
    correctness = 0.0
else:
    l2 = passed / total
    l3 = max(0, 1.0 - issues * 0.05)
    l4 = passed / total
    correctness = 0.20 * l2 + 0.10 * l3 + 0.70 * l4
```

### VT 扣分

关键 VT（Verification Target）失败会额外扣分：

```
correctness = max(0, correctness - 0.1 * critical_vt_fail_count)
```

支持的检测类型：`go build`、`unit test`、`e2e test case`、`go vet`、`race detector` 等。

## 效率评分

```
efficiency = 1.0 - min(1.0, cost_usd / cost_budget)
```

成本预算按难度等级设定（T1=$0.50，T2=$1.00，T3=$2.00，T4=$5.00）。成本使用统一的 `metrics.EstimateCost()` 函数计算，支持可配置的模型定价。

## LLM Judge -- Rubric 评分

配置 `judge.enabled: true` 后，每个完成的运行会由 LLM Judge 使用结构化 Rubric 评估。七个维度按 0-5 分评分：

| 维度 | 权重 | 评估内容 |
|------|------|---------|
| Correctness | 25% | 逻辑正确性、边界情况、隐藏 bug |
| Readability | 15% | 命名、结构、注释、控制流 |
| Simplicity | 15% | 不过度工程化、最简可行方案 |
| Robustness | 15% | 错误处理、资源管理、并发安全 |
| Minimality | 15% | 干净 diff、无无关修改 |
| Maintainability | 15% | 内聚、耦合、可扩展性 |
| Go Idioms | 补充 | Go 语言风格一致性 |

### 一致性验证

Judge 响应中包含布尔指标和数值评分。一致性检查会标记矛盾情况（如 5/6 布尔指标为正但分数为 2）。

### Judge 配置

```yaml
judge:
  enabled: true
  model: "claude-sonnet-4-20250514"
  input_price_per_mtok: 3.0
  output_price_per_mtok: 15.0
  repeat: 1
```

## 成对比较（Pairwise）

使用 `compare --pairwise` 时，LLM Judge 对两个 tag 的代码进行头对头比较：

1. 两个实现的 diff 同时提交给 Judge
2. Judge 在多个维度上评估
3. **位置偏差检测**：比较运行两次（交换顺序），标记结果是否一致

结果存储在 `pairwise_results` 表中。

## 稳定性评分

```
stability = pass_count / K
```

K 为 `runs_per_task`（需要 K >= 3 才有意义）。衡量工作流在多次运行中产出正确结果的一致性。

## 复合评分（四维）

```
final = 0.40 * correctness + 0.25 * efficiency + 0.25 * quality + 0.10 * stability
```

缺失维度的权重按比例重分配。

## 统计方法

### Wilson Score 置信区间

通过率以 95% Wilson Score 置信区间报告：

```
Pass Rate: 85.0% [72.3-93.1]
```

### 显著性检验

对比两个 tag 时，通过检查左右 Wilson CI 是否重叠判断统计显著性。不重叠标记为 `*`。

### 低样本量警告

总运行数 K < 5 时显示警告。

## 通过/失败判定

运行"通过"的条件：L1 构建成功 且 L4 E2E 全部通过。

## 安全否决

`gosec` 发现高/严重问题时，标记 `SECURITY_FAIL`（独立报告，不影响数值分数）。
