# Task: 添加 /health 端点

## 目标
为 TODO API 添加 GET /health 健康检查端点。

## 变更范围
- main.go: 在 setupRouter() 中添加 /health 路由

## 具体要求
- REQ-1: GET /health 返回 HTTP 200
- REQ-2: 响应体为 JSON: {"status":"ok"}
- REQ-3: Content-Type 设为 application/json

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只添加 /health 端点，不做其他修改

## 测试策略
- 验证 GET /health 返回 200
- 验证响应体包含 {"status":"ok"}

## 不做什么
- 不添加其他端点
- 不修改现有 handler 逻辑
- 不添加中间件
