# Task: 实现完整的 Operator

## 目标
实现完整的 Kubernetes-style Operator：资源定义（Spec/Status/Conditions）、Reconciler 状态机、Status 管理、事件记录。

## 当前状态
- 有基本的 Reconciler 骨架但不完整
- 没有 Conditions
- 没有事件记录
- 状态转换逻辑简陋

## 变更范围
- types.go: 完整的资源定义
- reconciler.go: 完整的 Reconcile 逻辑
- events.go: EventRecorder 实现
- store.go: 资源存储

## 具体要求
- REQ-1: Resource 包含 Metadata(Name, Generation, Labels), Spec, Status
- REQ-2: Status 包含 Phase, Conditions[], ObservedGeneration, ReadyReplicas
- REQ-3: Condition 包含 Type, Status(True/False/Unknown), LastTransitionTime, Reason, Message
- REQ-4: 预定义 Condition 类型：Ready, Available, Progressing, Degraded
- REQ-5: Reconciler 实现完整状态机：
  - Pending → Progressing (开始创建副本)
  - Progressing → Running (所有副本就绪)
  - Running → Degraded (副本不足)
  - 任意 → Failed (不可恢复错误)
- REQ-6: Reconcile 返回 Result{Requeue, RequeueAfter}
- REQ-7: EventRecorder 记录事件：Normal/Warning 类型，包含 reason + message
- REQ-8: Status 更新时设置 ObservedGeneration = Metadata.Generation
- REQ-9: Condition 更新时只有状态变化才更新 LastTransitionTime
- REQ-10: Store 支持 CRUD + List + Watch(callback)
- REQ-11: Runner 运行 reconcile 循环，支持 RequeueAfter 延迟
- REQ-12: 所有操作并发安全

## 状态机
```
Pending → Progressing → Running
                     ↓         ↓
                   Failed   Degraded
```

## 约束
- 纯 stdlib
- 模拟 Kubernetes 风格但不需要真实 API

## 测试策略
- 资源创建触发 reconcile
- 状态机完整流转
- Conditions 正确更新
- 事件记录
- 副本不足降级
- 不可恢复错误
- ObservedGeneration 正确
- 并发安全

## 不做什么
- 不实现真正的 Kubernetes API
- 不实现 Admission Webhook
- 不实现 leader election
