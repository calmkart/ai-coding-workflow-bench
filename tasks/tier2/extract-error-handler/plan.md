# Task: 提取统一错误处理

## 目标
提取 errorResponse 函数，统一所有错误响应为 JSON 格式。

## 变更范围
- handlers.go: 添加 errorResponse 函数，替换所有 http.Error 调用

## 具体要求
- REQ-1: 创建 errorResponse(w, message, statusCode) 函数
- REQ-2: 错误响应格式: {"error": "message"}
- REQ-3: 设置 Content-Type: application/json
- REQ-4: 替换所有 handler 中的 http.Error 调用

## 约束
- setupRouter() 函数签名不可更改
- 错误的 HTTP 状态码不变

## 测试策略
- 验证 404 响应是 JSON 格式
- 验证 400 响应是 JSON 格式
- 验证响应有 Content-Type: application/json

## 不做什么
- 不改变成功响应格式
- 不添加错误码或详细堆栈
