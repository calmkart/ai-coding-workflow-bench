# Task: 实现 Fan-Out

## 目标
实现 FanOut 并发处理函数。

## 变更范围
- fanout.go: 实现 FanOut 函数

## 具体要求
- REQ-1: FanOut[T, R](input []T, workers int, fn func(T) R) []R
- REQ-2: 使用 workers 个 goroutine 并发处理
- REQ-3: 结果保持与输入相同的顺序
- REQ-4: 所有输入都被处理

## 约束
- 不使用外部依赖
- 结果有序

## 测试策略
- 验证结果正确
- 验证顺序保持
- 验证并发执行

## 不做什么
- 不添加错误处理
- 不添加 context 取消
