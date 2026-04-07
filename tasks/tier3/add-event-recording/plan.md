# Task: 添加事件记录器

## 目标
添加 EventRecorder 接口记录 reconcile 过程中的事件。

## 变更范围
- reconciler.go: 注入 EventRecorder，在关键步骤记录事件

## 具体要求
- REQ-1: Event struct {ResourceName, Type, Reason, Message string, Timestamp time.Time}
- REQ-2: EventRecorder 接口 {Record(Event), Events(resourceName) []Event}
- REQ-3: InMemoryEventRecorder 实现（用于测试）
- REQ-4: Normal/Warning 两种事件类型
- REQ-5: Reconciler 在以下时机记录事件：开始 reconcile、状态变更、完成、错误
- REQ-6: 事件包含资源名称，可按资源过滤
- REQ-7: 保留最近 100 条事件（per resource）
- REQ-8: 并发安全

## 约束
- Reconciler 接口不变
- 纯 stdlib

## 测试策略
- 验证事件记录
- 验证事件按资源过滤
- 验证事件类型正确

## 不做什么
- 不实现持久化
- 不实现事件聚合
