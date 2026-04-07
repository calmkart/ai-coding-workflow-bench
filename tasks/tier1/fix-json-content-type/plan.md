# Task: 修复 JSON Content-Type 头

## 目标
为 TODO API 的所有 JSON 响应添加正确的 Content-Type: application/json 头。

## 变更范围
- handlers.go: 为缺少 Content-Type 的 handler 添加响应头

## 具体要求
- REQ-1: GET /todos 响应包含 Content-Type: application/json
- REQ-2: GET /todos/{id} 响应包含 Content-Type: application/json
- REQ-3: POST /todos 响应包含 Content-Type: application/json
- REQ-4: 错误响应不需要修改（http.Error 自动设置 text/plain）

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只添加 Content-Type 头，不做其他修改

## 测试策略
- 验证 GET /todos 响应头包含 Content-Type: application/json
- 验证 GET /todos/{id} 响应头包含 Content-Type: application/json
- 验证 POST /todos 响应头包含 Content-Type: application/json

## 不做什么
- 不改变 API 路径或响应体格式
- 不添加新的端点
- 不添加中间件
