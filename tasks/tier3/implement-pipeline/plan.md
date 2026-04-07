# Task: 实现类型安全 Pipeline

## 目标
实现类型安全的 Pipeline[I,O]：Stage 链式调用，自动错误传播。

## 变更范围
- pipeline.go: 需要实现的管道框架

## 具体要求
- REQ-1: Pipeline[I, O] 表示从 I 到 O 的转换管道
- REQ-2: NewPipeline(fn) 创建单 stage 管道
- REQ-3: Then 函数连接两个 Pipeline：Pipeline[A,B].Then(Pipeline[B,C]) -> Pipeline[A,C]
- REQ-4: Execute(input I) (O, error) 执行管道
- REQ-5: 任何 stage 错误立即停止并返回
- REQ-6: Parallel(pipelines...) 并行执行多个管道，收集结果
- REQ-7: WithRetry(n) 包装 stage 添加重试

## 约束
- 纯 stdlib + Go 泛型
- 类型安全，无 interface{} 或反射

## 测试策略
- 验证单 stage 管道
- 验证链式 stage
- 验证错误传播
- 验证并行执行

## 不做什么
- 不实现并发管道
- 不实现背压
