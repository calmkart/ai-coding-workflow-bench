# Task: 添加 SSE 实时变更通知

## 目标
添加 Server-Sent Events (SSE) 端点 GET /events，当 todo 发生 CRUD 变更时推送事件。

## 变更范围
- main.go: 添加 /events 路由
- handlers.go: 在 CRUD 操作中发送事件
- 新增 broadcaster.go: 事件广播器

## 具体要求
- REQ-1: 实现 Broadcaster struct，管理 SSE 客户端连接
- REQ-2: GET /events 端点，返回 text/event-stream
- REQ-3: 创建 todo 时发送 event: created，data 为 JSON
- REQ-4: 更新 todo 时发送 event: updated
- REQ-5: 删除 todo 时发送 event: deleted
- REQ-6: 客户端断开连接时清理
- REQ-7: 支持多个并发 SSE 客户端
- REQ-8: 事件格式符合 SSE 规范（event: type\ndata: json\n\n）

## 约束
- setupRouter() 函数签名不变
- 纯 stdlib（net/http 支持 SSE）
- 不依赖外部 WebSocket 库

## 测试策略
- 验证 /events 返回 text/event-stream
- 验证创建 todo 后 SSE 收到事件
- 验证事件格式正确

## 不做什么
- 不实现 WebSocket
- 不实现事件持久化
- 不实现重连 ID
