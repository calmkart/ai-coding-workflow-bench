# Task: 实现简单工作流引擎

## 目标
实现简单的工作流引擎：解析 YAML 格式的工作流定义，支持步骤的顺序和并行执行，收集结果。

## 当前状态
- 只有空的结构定义

## 变更范围
- workflow.go: 工作流定义和解析
- executor.go: 工作流执行引擎
- main.go: CLI 入口

## 具体要求
- REQ-1: 自定义简单 YAML-like 格式解析（key: value, 缩进表示层级）
- REQ-2: Workflow 包含 name, steps 列表
- REQ-3: Step 包含 name, command(shell 命令字符串，模拟执行), depends_on(依赖步骤名), parallel(bool)
- REQ-4: 顺序执行：按定义顺序执行 steps
- REQ-5: 并行执行：标记为 parallel 的连续步骤同时执行
- REQ-6: 依赖：step A depends_on step B，则 B 必须在 A 之前完成
- REQ-7: Execute 返回 WorkflowResult：每个步骤的执行状态、耗时、输出
- REQ-8: 步骤失败时（command 返回错误），后续依赖步骤跳过
- REQ-9: context 取消时停止执行
- REQ-10: WorkflowResult.Success 为 true 当所有步骤成功
- REQ-11: 支持步骤超时（timeout 字段，秒）
- REQ-12: 步骤执行使用可注入的 StepRunner 接口（便于测试）

## 工作流格式（简化 YAML）
```
name: deploy
steps:
  - name: build
    command: go build ./...
    timeout: 60
  - name: test
    command: go test ./...
    depends_on: build
    timeout: 120
  - name: lint
    command: golint ./...
    depends_on: build
    parallel: true
  - name: deploy
    command: deploy.sh
    depends_on: test
```

## 约束
- 纯 stdlib
- 不用 YAML 库（手写简单解析器）
- 步骤 command 通过 StepRunner 接口执行（测试用 mock）

## 测试策略
- 解析工作流定义
- 顺序执行步骤
- 并行步骤同时执行
- 依赖关系正确
- 步骤失败跳过依赖
- 超时处理
- context 取消
- 结果收集完整

## 不做什么
- 不实现条件分支
- 不实现循环
- 不实现变量替换
