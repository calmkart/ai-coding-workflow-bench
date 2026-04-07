# Task: 拆分巨型 Reconcile 为子 reconciler

## 目标
将巨型 Reconcile() 拆分为 validateSpec、ensureResources、updateStatus 三个独立的子 reconciler。

## 变更范围
- reconciler.go: 拆分巨型 Reconcile
- types.go: 可能需要扩展类型

## 具体要求
- REQ-1: SubReconciler 接口 { Reconcile(ctx, resource) (Result, error) }
- REQ-2: ValidateSpecReconciler 验证 resource spec 合法性
- REQ-3: EnsureResourcesReconciler 确保子资源存在且状态正确
- REQ-4: UpdateStatusReconciler 更新 resource status
- REQ-5: CompositeReconciler 按顺序调用子 reconciler
- REQ-6: 任一子 reconciler 返回 Requeue 或 error，停止后续
- REQ-7: 保持现有 Reconciler 接口兼容

## 约束
- 保持 Reconciler 接口签名不变
- 纯 stdlib

## 测试策略
- 验证各子 reconciler 独立正确
- 验证组合后行为与原始一致
- 验证错误传播

## 不做什么
- 不添加新功能
- 不改变状态转换逻辑
