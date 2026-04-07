# Task: 实现内存发布订阅

## 目标
实现内存 PubSub：Topic 隔离、异步投递、关闭清理。

## 变更范围
- pubsub.go: 需要实现的发布订阅框架

## 具体要求
- REQ-1: PubSub struct 管理 topic 和订阅者
- REQ-2: Subscribe(topic) <-chan Message 返回消息 channel
- REQ-3: Publish(topic, data) error 发布消息到所有订阅者
- REQ-4: Unsubscribe(topic, ch) 取消订阅
- REQ-5: Topic 隔离，消息只发给该 topic 的订阅者
- REQ-6: 异步投递，Publish 不阻塞
- REQ-7: Close() 关闭所有 channel，清理资源
- REQ-8: 订阅者 channel 有缓冲，慢消费者不阻塞其他订阅者

## 约束
- 纯 stdlib
- 并发安全

## 测试策略
- 验证发布和订阅
- 验证 topic 隔离
- 验证取消订阅
- 验证 Close 清理

## 不做什么
- 不实现持久化
- 不实现消息回溯
