# Task: 实现事件溯源模式

## 目标
当前 API 直接修改 Todo 状态（CRUD），没有变更历史。重构为事件溯源模式：所有状态变更记录为事件，状态通过回放事件重建。

## 当前问题
- 直接修改内存中的 todo slice
- 无法追溯变更历史
- 无法回放到任意时间点

## 变更范围
- events.go: 定义 Event 类型和 EventStore 接口
- aggregate.go: TodoAggregate 通过 Apply 事件变更状态
- handlers.go: 重写 handler 使用事件驱动
- main.go: setupRouter 注入 EventStore

## 具体要求
- REQ-1: 定义事件类型：TodoCreated, TodoUpdated, TodoDeleted, TodoCompleted
- REQ-2: Event 包含 ID, Type, AggregateID, Data(json), Timestamp, Version
- REQ-3: EventStore 接口：Append(event), GetEvents(aggregateID), GetAllEvents()
- REQ-4: InMemoryEventStore 实现，并发安全
- REQ-5: TodoAggregate.Apply(event) 根据事件类型更新状态
- REQ-6: Replay(events) 从事件流重建完整 TodoAggregate 状态
- REQ-7: 所有 HTTP handler 通过追加事件修改状态
- REQ-8: GET /events 返回完整事件流
- REQ-9: GET /events?aggregate_id=X 返回特定聚合的事件
- REQ-10: 乐观并发控制：Append 时检查 version，冲突返回 409
- REQ-11: 事件不可变 — 一旦写入不可修改或删除
- REQ-12: 基本 CRUD API 行为不变（对外兼容）

## 事件结构
```go
type Event struct {
    ID          string    `json:"id"`
    Type        string    `json:"type"`        // "TodoCreated", "TodoUpdated", etc.
    AggregateID string    `json:"aggregate_id"`
    Data        json.RawMessage `json:"data"`
    Timestamp   time.Time `json:"timestamp"`
    Version     int       `json:"version"`
}
```

## 约束
- setupRouter() 签名不变
- 纯 stdlib
- 对外 API 行为兼容（同样的 /todos CRUD）

## 测试策略
- 创建 todo 生成 TodoCreated 事件
- 更新 todo 生成 TodoUpdated 事件
- 删除 todo 生成 TodoDeleted 事件
- 回放事件重建正确状态
- 事件流 API 返回完整历史
- 乐观并发控制
- 事件不可变

## 不做什么
- 不实现快照（snapshot）
- 不实现事件 projection 到读模型
- 不实现事件发布/订阅
