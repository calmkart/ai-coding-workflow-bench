# Task: 实现事件总线

## 目标
实现事件总线：Subscribe/Publish/Unsubscribe，支持多个 listener，按 topic 隔离。

## 变更范围
- eventbus.go: 需要实现的事件总线框架

## 具体要求
- REQ-1: EventBus struct 管理多个 topic 的订阅者
- REQ-2: Subscribe(topic, handler) SubscriptionID — 订阅事件
- REQ-3: Publish(topic, data any) — 发布事件给所有订阅者
- REQ-4: Unsubscribe(id) — 取消订阅
- REQ-5: 支持通配符 "*" 接收所有 topic
- REQ-6: 异步投递（goroutine），但 PublishSync 同步投递
- REQ-7: Close() 关闭总线，清理资源
- REQ-8: 并发安全

## 约束
- 纯 stdlib
- 不用外部消息库

## 测试策略
- 验证订阅和发布
- 验证取消订阅
- 验证多个 topic 隔离
- 验证并发安全

## 不做什么
- 不实现持久化
- 不实现分布式
