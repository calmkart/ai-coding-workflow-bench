# Task: 提取 logging 中间件

## 目标
从每个 handler 中提取重复的 log.Printf 调用，创建一个统一的 logging 中间件。

## 变更范围
- main.go: 添加 loggingMiddleware 并在 setupRouter 中应用
- handlers.go: 移除 handler 中的重复 log.Printf 调用

## 具体要求
- REQ-1: 创建 loggingMiddleware(next http.Handler) http.Handler
- REQ-2: 中间件记录请求方法、路径、状态码和耗时
- REQ-3: 移除 handler 内的所有 log.Printf 调用
- REQ-4: setupRouter() 返回被中间件包装的 handler

## 约束
- setupRouter() 函数签名不可更改
- 所有现有端点功能不变

## 测试策略
- 验证所有端点功能正常
- 验证响应中不包含日志相关副作用

## 不做什么
- 不添加新端点
- 不改变响应格式
