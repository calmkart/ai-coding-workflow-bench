# Task: 修复 Dedup 空 slice panic

## 目标
修复 Dedup 函数：当输入为空 slice 或 nil 时 panic（index out of range），应返回空 slice。

## 变更范围
- strkit.go: 修复 Dedup 函数的空 slice 处理

## 具体要求
- REQ-1: nil 输入返回空（nil 或空 slice）
- REQ-2: 空 slice 输入返回空 slice
- REQ-3: 单元素 slice 原样返回
- REQ-4: 有重复的 slice 正确去重
- REQ-5: 无重复的 slice 原样返回
- REQ-6: 保持元素原始顺序

## 约束
- Dedup 函数签名不可更改

## 测试策略
- 验证 nil 输入
- 验证空 slice
- 验证单元素
- 验证有重复
- 验证无重复

## 不做什么
- 不改变返回顺序逻辑
- 不添加新函数
