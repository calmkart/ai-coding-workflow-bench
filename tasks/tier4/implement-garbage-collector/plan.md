# Task: 实现 Owner Reference 级联删除 GC

## 目标
资源通过 OwnerReference 建立 parent-child 关系，但删除 parent 不会自动删除 children（产生孤儿）。实现 GC 控制器自动级联删除。

## 当前状态
- 资源有 OwnerReferences 字段但删除不级联
- 孤儿资源会累积

## 变更范围
- types.go: OwnerReference 类型
- store.go: 资源存储
- gc.go: GarbageCollector 实现

## 具体要求
- REQ-1: OwnerReference 包含 Name, Kind, UID
- REQ-2: Resource 包含 Metadata.OwnerReferences []OwnerReference
- REQ-3: Resource 包含 Metadata.UID（唯一标识）
- REQ-4: Resource 包含 Metadata.DeletionTimestamp（nil 表示未删除）
- REQ-5: GarbageCollector.CollectOnce() 扫描并清理：
  - 找出所有 owner 已不存在的资源
  - 级联删除这些孤儿资源
  - 返回删除的资源数
- REQ-6: 级联删除是递归的（A→B→C，删除 A 最终删除 B 和 C）
- REQ-7: GarbageCollector.Run(ctx) 定期运行 CollectOnce
- REQ-8: Store.Delete 设置 DeletionTimestamp 而非立即删除
- REQ-9: Store.Purge 真正删除已标记且无 dependents 的资源
- REQ-10: 资源的 Finalizers 必须为空才能被 GC 清理
- REQ-11: Store 支持按 owner 查找 dependents
- REQ-12: GC 并发安全

## GC 流程
```
1. 扫描所有资源
2. 对每个有 OwnerReference 的资源：
   a. 检查 owner 是否存在
   b. 如果 owner 不存在 → 标记 dependent 为删除
3. 对每个标记删除且无 dependents 的资源：
   a. 如果 Finalizers 为空 → Purge
4. 重复直到无新删除
```

## 约束
- 纯 stdlib
- 内存存储

## 测试策略
- 删除 parent 后 child 被清理
- 多级级联删除
- 有 finalizer 的资源不被删除
- 循环引用处理
- 无 owner 的资源不受影响
- 并发安全

## 不做什么
- 不实现背景删除（background deletion）vs 前台删除
- 不实现 orphan 策略（只实现 cascade）
- 不实现持久化
