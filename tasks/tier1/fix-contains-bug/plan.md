# Task: 修复 ContainsAny 空候选列表 bug

## 目标
修复 ContainsAny 函数：当 candidates 为空 slice 时应返回 false，当前实现错误返回 true。

## 变更范围
- strkit.go: 修复 ContainsAny 函数的边界条件

## 具体要求
- REQ-1: candidates 为空 slice 时返回 false
- REQ-2: candidates 为 nil 时返回 false
- REQ-3: s 为空字符串且 candidates 包含空字符串时返回 true
- REQ-4: 正常匹配逻辑不受影响

## 约束
- ContainsAny 函数签名不可更改

## 测试策略
- 验证空 candidates 返回 false
- 验证 nil candidates 返回 false
- 验证正常匹配返回 true
- 验证不匹配返回 false

## 不做什么
- 不添加新函数
- 不修改其他函数
