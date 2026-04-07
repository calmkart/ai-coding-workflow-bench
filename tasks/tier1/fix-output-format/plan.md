# Task: 修复列表输出对齐

## 目标
修复 formatTasks 函数：当前使用简单 \t 拼接导致不同长度的字段不对齐，应使用 tabwriter 或 fmt.Sprintf 对齐。

## 变更范围
- format.go: 修复 formatTasks 函数使用 tabwriter

## 具体要求
- REQ-1: ID、标题、状态列对齐
- REQ-2: 包含表头行
- REQ-3: 使用 tabwriter 或固定宽度 fmt.Sprintf 格式化

## 约束
- formatTasks 函数签名不可更改
- Task struct 不可更改

## 测试策略
- 验证输出包含表头
- 验证列对齐（相同列的内容在相同位置开始）
- 验证不同长度的标题仍然对齐

## 不做什么
- 不修改 Task struct
- 不改变输出的信息内容
