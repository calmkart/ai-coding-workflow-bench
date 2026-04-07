# Task: 添加 Owner Reference 和级联删除

## 目标
添加 OwnerReference 管理，子资源创建时设置 owner，父资源删除时级联删除子资源。

## 变更范围
- types.go: 添加 OwnerReference 结构
- reconciler.go: 创建子资源时设置 owner，删除时级联

## 具体要求
- REQ-1: OwnerReference struct {Name, Kind string}
- REQ-2: Resource.OwnerRef *OwnerReference 字段
- REQ-3: SetOwner(child, parent) 设置 owner reference
- REQ-4: GetOwned(store, parent) 获取所有子资源
- REQ-5: DeleteWithCascade(store, name) 删除资源及其所有子资源
- REQ-6: 子资源创建时自动设置 owner
- REQ-7: IsOwnedBy(child, parent) 检查 owner 关系

## 约束
- Reconciler 接口不变
- 纯 stdlib

## 测试策略
- 验证 owner reference 设置
- 验证级联删除
- 验证孤儿资源检测

## 不做什么
- 不实现垃圾收集器
- 不实现多层级联
