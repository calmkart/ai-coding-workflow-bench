# Task: 添加请求体验证

## 目标
为 POST /todos 添加请求验证，拒绝无效输入。

## 变更范围
- handlers.go: 在 createTodo 中添加验证逻辑

## 具体要求
- REQ-1: title 不能为空字符串或纯空白
- REQ-2: title 长度不能超过 200 字符
- REQ-3: 验证失败返回 400 Bad Request 和 JSON 错误信息
- REQ-4: 错误响应格式: {"error": "validation failed", "details": ["title is required"]}

## 约束
- setupRouter() 函数签名不可更改
- 其他端点功能不变

## 测试策略
- 验证空 title 返回 400
- 验证纯空白 title 返回 400
- 验证超长 title 返回 400
- 验证正常 title 返回 201

## 不做什么
- 不添加其他字段的验证
- 不改变成功响应格式
