# Task: 实现 Actor 模型消息传递

## 目标
实现 Actor 模型：每个 Actor 是独立的并发单元，通过 Mailbox 接收消息，按顺序处理消息，通过 ActorRef 发送消息。

## 当前状态
- 只有空的接口定义

## 变更范围
- actor.go: 完整实现

## 具体要求
- REQ-1: Actor 接口：Receive(ctx ActorContext, msg Message)
- REQ-2: Message 包含 Type(string) 和 Payload(any)
- REQ-3: Mailbox 是有缓冲的消息队列（channel-based）
- REQ-4: ActorSystem 管理所有 Actor 的生命周期
- REQ-5: Spawn(name, actor) 创建并启动 Actor，返回 ActorRef
- REQ-6: ActorRef.Send(msg) 异步发送消息到 Actor 的 Mailbox
- REQ-7: Actor 按顺序处理 Mailbox 中的消息（单 goroutine 消费）
- REQ-8: ActorContext 提供 Self() ActorRef 和 System() *ActorSystem
- REQ-9: ActorContext 提供 Reply(msg Message) 回复发送者
- REQ-10: ActorSystem.Shutdown() 优雅关闭所有 Actor
- REQ-11: ActorSystem.Lookup(name) 通过名字查找 ActorRef
- REQ-12: 消息有发送者信息（From *ActorRef），用于 Reply

## Actor 模型核心原则
1. Actor 是最小并发单元
2. Actor 通过消息通信，不共享状态
3. Actor 按顺序处理消息
4. Actor 可以创建子 Actor

## 约束
- 纯 stdlib
- Mailbox 基于 buffered channel
- 每个 Actor 一个 goroutine

## 测试策略
- 发送消息并处理
- 消息顺序保证
- Actor 间通信
- Reply 回复
- System Shutdown
- Lookup 查找
- 并发安全
- 大量消息处理

## 不做什么
- 不实现 Actor 层级监督
- 不实现远程 Actor
- 不实现持久化
