# Task: 添加分页响应头

## 目标
为 GET /todos 添加分页相关的响应头。

## 变更范围
- handlers.go: 在 listTodos 中添加响应头

## 具体要求
- REQ-1: 添加 X-Total-Count 头（总记录数）
- REQ-2: 添加 X-Page 头（当前页码）
- REQ-3: 添加 X-Page-Size 头（每页大小）
- REQ-4: 这些头应在 JSON 响应体之前设置

## 约束
- setupRouter() 函数签名不可更改
- 分页逻辑不变

## 测试策略
- 验证响应包含 X-Total-Count 头
- 验证 X-Total-Count 值正确
- 验证其他分页头正确

## 不做什么
- 不改变分页逻辑
- 不添加 Link 头（简化版本）
