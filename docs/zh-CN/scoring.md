# 评分系统

## 概述

workflow-bench 采用多层评分系统。当前阶段（P1）仅实现了正确性评分。后续阶段将添加效率、代码质量（LLM Judge）和稳定性评分，最终合成综合分数。

## 四层验证 (L1-L4)

每次运行都经过四层验证，由自动生成的 `verify.sh` 脚本执行。

### L1：构建

```bash
go build ./...
```

二进制门控。代码无法编译则 correctness = 0.0，不再执行后续检查。

### L2：单元测试

```bash
go test ./... -count=1 -race -v
```

运行任务自带的单元测试并启用竞态检测器，统计通过和失败的测试数。

### L3：静态分析

```bash
go vet ./...
staticcheck ./...   # 如已安装
gosec ./...          # 如已安装
```

统计 lint 问题数。每个问题扣除 L3 子分数的 5%。仅 `go vet` 为必需；`staticcheck` 和 `gosec` 在可用时使用。

### L4：端到端测试

```bash
go test -v -run TestBenchE2E -count=1 -race ./...
```

运行任务 `verify/` 目录中的 E2E 测试（真值）。workflow 智能体看不到这些测试——它们仅在验证阶段被复制到 worktree 中。

对于 `http-server` 类型任务，E2E 测试使用 `httptest` 并调用 `setupRouter()` 来测试 API。

## 正确性评分公式

正确性分数是 L2、L3、L4 结果的加权组合，以 L1 为门控：

```
if L1 == FAIL:
    correctness = 0.0
else:
    l2_score = ut_passed / ut_total         # 0.0-1.0
    l3_score = max(0, 1.0 - issues * 0.05)  # 每个问题扣 5%
    l4_score = e2e_passed / e2e_total       # 0.0-1.0

    correctness = 0.20 * l2_score + 0.10 * l3_score + 0.70 * l4_score
```

**边界情况**：
- `ut_total == 0`（无单元测试）时，`l2_score = 1.0`
- `e2e_total == 0`（无 E2E 测试）时，`l4_score = 1.0`

### 计算示例

**满分运行**：L1=PASS, L2=8/8, L3=0 个问题, L4=5/5

```
l2 = 8/8 = 1.0
l3 = max(0, 1.0 - 0*0.05) = 1.0
l4 = 5/5 = 1.0
correctness = 0.20*1.0 + 0.10*1.0 + 0.70*1.0 = 1.00
```

**部分通过**：L1=PASS, L2=6/8, L3=2 个问题, L4=3/5

```
l2 = 6/8 = 0.75
l3 = max(0, 1.0 - 2*0.05) = 0.90
l4 = 3/5 = 0.60
correctness = 0.20*0.75 + 0.10*0.90 + 0.70*0.60 = 0.15 + 0.09 + 0.42 = 0.66
```

**构建失败**：L1=FAIL

```
correctness = 0.0
```

## Verification Target (VT)

任务可定义 verification target——智能体在重构过程中可能引入的已知陷阱。VT 横跨 9 个类别（共定义 78 种模式）：

| 类别 | 示例 |
|------|------|
| Concurrency | goroutine 泄漏、数据竞争、死锁 |
| Error handling | 错误吞没、nil interface 陷阱 |
| Memory/Resources | 未关闭 HTTP body、文件描述符泄漏 |
| Interface/Types | nil interface 陷阱、type assertion panic |
| Package/Dependencies | 循环导入、init() 顺序依赖 |
| HTTP | 缺少 server timeout、handler panic 恢复 |
| Distributed | 无退避重试、非幂等重试 |
| K8s Operator | 无限 reconcile、finalizer 未清理 |
| Testing | 测试污染、时间依赖测试 |

### VT 扣分（计划中）

VT 检测实现后，关键 VT 失败会从正确性分数中扣除：

```
correctness = max(0, correctness - 0.1 * critical_vt_fail_count)
```

每个关键 VT 失败扣 0.1 分（10%），下限为 0。

## 通过/失败判定

运行被判定为"通过"的条件：
- L1 构建成功 **且**
- 所有 L4 E2E 测试通过（L4 passed == L4 total，且 total > 0）

该结果用于报告中的通过率指标。

## 未来评分维度（P2+）

### 效率评分

```
efficiency = 1.0 - min(1.0, cost_usd / cost_budget)
```

其中 `cost_budget` 按等级设定（T1=$0.50, T2=$1.00, T3=$2.00, T4=$5.00）。如无 token 数据，效率评分为 N/A 并从综合分数中排除。

### LLM 质量评分

由 LLM Judge 通过 Rubric 评估的六个维度（每维度 0-5 分）：

| 维度 | 权重 | 评估内容 |
|------|------|----------|
| Correctness | 25% | 逻辑正确性、边界情况、隐藏缺陷 |
| Readability | 15% | 命名、结构、注释、控制流 |
| Simplicity | 15% | 无过度工程、最简可用方案 |
| Robustness | 15% | 错误处理、资源管理、并发安全 |
| Minimality | 15% | 干净的 diff、无无关变更、范围合理 |
| Maintainability | 15% | 内聚性、耦合度、可扩展性、模式一致性 |

另有两个补充维度单独报告：
- **Go Idioms**：Go 语言风格一致性
- **Workflow Private**：计划遵循度（根据计划内容动态生成）

### 稳定性评分

```
stability = pass_count / K
```

其中 K 为 `runs_per_task`（K >= 3 才有意义）。

### 综合评分

```
final = 0.40 * correctness + 0.25 * efficiency + 0.25 * quality + 0.10 * stability
```

当某维度为 N/A 时，其权重按比例重新分配给其余维度。

**安全否决**：如果 `gosec` 发现 High/Critical 问题，运行标记为 `SECURITY_FAIL`（单独报告，不影响数值评分）。
