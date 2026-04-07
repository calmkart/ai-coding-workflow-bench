# Task: 迁移到 Conditions 数组

## 目标
从 Phase 字符串状态迁移到 Conditions 数组（type/status/reason/message/lastTransitionTime）。

## 变更范围
- types.go: 添加 Condition 类型
- reconciler.go: 使用 Conditions 而非 Phase

## 具体要求
- REQ-1: Condition struct {Type, Status, Reason, Message string, LastTransitionTime time.Time}
- REQ-2: ResourceStatus.Conditions []Condition 字段
- REQ-3: SetCondition(status, condition) 设置或更新条件
- REQ-4: GetCondition(status, type) *Condition 获取条件
- REQ-5: IsConditionTrue(status, type) bool 检查条件是否为 True
- REQ-6: Condition Types: "Ready", "Available", "Progressing"
- REQ-7: LastTransitionTime 只在 Status 变更时更新
- REQ-8: 保留 Phase 字段但从 Conditions 派生

## 约束
- Reconciler 接口不变
- 纯 stdlib

## 测试策略
- 验证条件设置和获取
- 验证 transition time 只在变更时更新
- 验证 Phase 从 Conditions 派生

## 不做什么
- 不改变 reconcile 逻辑
- 不添加自定义 condition types
