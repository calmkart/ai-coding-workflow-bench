# Task: 实现泛型对象池

## 目标
实现泛型对象池 Pool[T]：Get/Put，并发安全，可配置最大池大小，提供使用统计。

## 变更范围
- pool.go: 需要实现的对象池框架

## 具体要求
- REQ-1: Pool[T] 使用 channel 实现有界池
- REQ-2: NewPool(factory, opts) 创建池，factory 用于创建新对象
- REQ-3: Get() T — 从池取对象，池空时用 factory 创建
- REQ-4: Put(obj T) — 归还对象到池，池满时丢弃
- REQ-5: PoolOption: WithMaxSize(n), WithResetFunc(fn)
- REQ-6: Reset 函数在 Put 时调用，清理对象状态
- REQ-7: Stats() PoolStats {Gets, Puts, News, Discards}
- REQ-8: 并发安全

## 约束
- 纯 stdlib + Go 泛型
- 不使用 sync.Pool（自行实现）

## 测试策略
- 验证 Get/Put 基本流程
- 验证池满丢弃
- 验证并发安全
- 验证统计准确

## 不做什么
- 不实现对象过期
- 不实现池缩容
