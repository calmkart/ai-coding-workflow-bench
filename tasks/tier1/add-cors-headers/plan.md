# Task: 添加 CORS 头支持

## 目标
为 TODO API 添加 CORS (Cross-Origin Resource Sharing) 支持，允许前端跨域调用。

## 变更范围
- main.go: 在 setupRouter() 中添加 CORS 中间件或在每个响应中添加 CORS 头

## 具体要求
- REQ-1: 所有响应添加 Access-Control-Allow-Origin: *
- REQ-2: 所有响应添加 Access-Control-Allow-Methods: GET, POST, DELETE, OPTIONS
- REQ-3: 所有响应添加 Access-Control-Allow-Headers: Content-Type
- REQ-4: OPTIONS 请求返回 204 No Content（preflight 响应）

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 使用标准库实现，不引入外部 CORS 库

## 测试策略
- 验证 GET 响应包含 CORS 头
- 验证 OPTIONS preflight 返回 204 和正确的 CORS 头
- 验证 POST 响应包含 CORS 头

## 不做什么
- 不限制特定域名（用 * 即可）
- 不添加 credentials 支持
- 不修改现有 handler 的业务逻辑
