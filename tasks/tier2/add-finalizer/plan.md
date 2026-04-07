# Task: 添加 Finalizer 模式

## 目标
资源删除时执行清理逻辑（finalizer 模式）。

## 变更范围
- types.go: 添加 Finalizers 和 DeletionTimestamp 字段
- reconciler.go: 在 Reconcile 中实现 finalizer 逻辑

## 具体要求
- REQ-1: Resource 添加 Finalizers []string 和 DeletionTimestamp *time.Time
- REQ-2: 创建资源时添加 finalizer
- REQ-3: 删除时先执行 cleanup，再移除 finalizer
- REQ-4: 所有 finalizer 移除后才真正删除

## 约束
- 不使用外部依赖
- 保持简单模拟

## 测试策略
- 验证 finalizer 被添加
- 验证删除时 cleanup 被调用
- 验证所有 finalizer 移除后才删除

## 不做什么
- 不添加外部清理调用
- 不添加超时
