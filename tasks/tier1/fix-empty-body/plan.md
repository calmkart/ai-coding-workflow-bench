# Task: 修复空 body 导致的 panic

## 目标
修复 POST /todos 端点：当请求体为空时服务器 panic（nil pointer dereference），应返回 400 Bad Request。

## 变更范围
- handlers.go: 修复 createTodo handler 的 body 解析

## 具体要求
- REQ-1: 空 body 返回 400 Bad Request
- REQ-2: 无效 JSON 返回 400 Bad Request
- REQ-3: 缺少 title 字段的有效 JSON (如 {}) 仍能正常处理（title 为空字符串）
- REQ-4: 正常 JSON body 继续正常工作

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只修复 body 解析逻辑

## 测试策略
- 验证空 body 返回 400
- 验证 nil body 返回 400
- 验证无效 JSON 返回 400
- 验证有效 JSON 返回 201

## 不做什么
- 不改变 API 路径或响应格式
- 不添加新的端点
- 不添加 body 大小限制
