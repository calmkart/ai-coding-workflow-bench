# Task: 修复查询参数解析 panic

## 目标
修复 GET /todos 端点：当 page_size 参数为非法值（如 "abc"）时服务器 panic，应返回 400 Bad Request。

## 变更范围
- handlers.go: 修复 listTodos handler 的查询参数解析

## 具体要求
- REQ-1: page_size 为非数字字符串时返回 400 Bad Request
- REQ-2: page 为非数字字符串时返回 400 Bad Request
- REQ-3: 合法参数仍正常工作
- REQ-4: 不传参数时使用默认值（page=1, page_size=10）

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只修复参数解析逻辑

## 测试策略
- 验证 page_size=abc 返回 400
- 验证 page=abc 返回 400
- 验证 page_size=10 正常工作
- 验证无参数返回默认结果

## 不做什么
- 不改变 API 路径或响应格式
- 不添加新的端点
- 不添加分页逻辑（已有）
