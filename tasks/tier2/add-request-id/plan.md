# Task: 添加请求 ID 中间件

## 目标
添加中间件为每个请求生成唯一的 UUID 放入 X-Request-ID 响应头。

## 变更范围
- main.go: 添加 requestIDMiddleware 并在 setupRouter 中应用

## 具体要求
- REQ-1: 创建 requestIDMiddleware(next http.Handler) http.Handler
- REQ-2: 生成 UUID v4 格式的请求 ID（使用 crypto/rand）
- REQ-3: 在响应头中设置 X-Request-ID
- REQ-4: 如果请求已包含 X-Request-ID 头，使用请求中的值

## 约束
- setupRouter() 函数签名不可更改
- 不使用外部依赖（crypto/rand 生成 UUID）

## 测试策略
- 验证每个响应都有 X-Request-ID
- 验证 ID 格式是 UUID v4
- 验证传入的 X-Request-ID 被保留

## 不做什么
- 不添加日志记录
- 不修改响应体
