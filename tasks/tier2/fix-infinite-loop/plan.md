# Task: 修复无限 reconcile 循环

## 目标
修复 Reconcile 返回 Requeue:true 时无退避导致的无限快速循环。

## 变更范围
- reconciler.go: 添加 backoff 逻辑
- types.go: Result 添加 RequeueAfter 字段

## 具体要求
- REQ-1: Result 添加 RequeueAfter time.Duration 字段
- REQ-2: Runner 在 Requeue 时使用 RequeueAfter 延迟
- REQ-3: 如果 Requeue 但 RequeueAfter 为 0，使用默认 1 秒
- REQ-4: 错误时使用指数退避（最大 30 秒）

## 约束
- 不使用外部依赖
- 不改变 Reconciler 接口

## 测试策略
- 验证 requeue 有延迟
- 验证不会无限快速循环
- 验证正常完成时不 requeue

## 不做什么
- 不添加 jitter
- 不添加 maxRetries
