# Task: 添加表格输出

## 目标
将 list 命令的输出改为 tabwriter 表格格式。

## 变更范围
- commands.go: 使用 text/tabwriter 格式化输出

## 具体要求
- REQ-1: 使用 text/tabwriter 对齐列
- REQ-2: 表头: ID, STATUS, TITLE
- REQ-3: STATUS 列显示 "done" 或 "pending"
- REQ-4: 列之间用 tab 分隔

## 约束
- 使用标准库 text/tabwriter
- 不改变数据结构

## 测试策略
- 验证输出包含表头
- 验证列对齐
- 验证 status 显示正确

## 不做什么
- 不添加颜色
- 不添加边框
