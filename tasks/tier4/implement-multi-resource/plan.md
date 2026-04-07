# Task: 实现多资源协调器（A->B->C 依赖链）

## 目标
当前 Reconciler 只能处理单一资源类型。扩展为多资源协调器：资源 A 创建后自动创建依赖资源 B，B 创建后自动创建 C，形成依赖链。

## 当前状态
- Reconciler 只处理单资源
- 没有跨资源依赖

## 变更范围
- types.go: 多种资源类型定义
- store.go: 支持多种资源的存储
- graph.go: ResourceGraph 定义依赖关系
- multi.go: MultiReconciler 实现

## 具体要求
- REQ-1: 三种资源类型：Application, Service, Endpoint
- REQ-2: 依赖关系：Application → Service → Endpoint (A 需要 B，B 需要 C)
- REQ-3: ResourceGraph 定义资源间的依赖关系
- REQ-4: MultiReconciler.Reconcile 按依赖顺序处理
- REQ-5: 创建 Application 时自动创建对应的 Service
- REQ-6: 创建 Service 时自动创建对应的 Endpoint
- REQ-7: 所有子资源设置 OwnerReference 指向父资源
- REQ-8: 父资源状态依赖子资源：Application Ready 当 Service Ready，Service Ready 当 Endpoint Ready
- REQ-9: 子资源变化触发父资源重新 reconcile
- REQ-10: 删除父资源级联删除子资源
- REQ-11: 子资源 Spec 变化时更新（而非重新创建）
- REQ-12: 整个 reconcile 过程并发安全

## 依赖链
```
Application (A)
  └── Service (B)
        └── Endpoint (C)

状态传播：C Ready → B Ready → A Ready
删除级联：Delete A → Delete B → Delete C
```

## 约束
- 纯 stdlib
- 模拟 Kubernetes 风格

## 测试策略
- 创建 Application 级联创建 Service 和 Endpoint
- 状态向上传播
- 删除级联
- 子资源 Spec 变化触发更新
- 多次 reconcile 幂等
- 并发安全

## 不做什么
- 不实现真正的 Kubernetes API
- 不超过 3 级依赖
- 不实现循环依赖检测
