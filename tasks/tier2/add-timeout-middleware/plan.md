# Task: 添加请求超时中间件

## 目标
添加 context timeout 中间件，防止请求处理时间过长。

## 变更范围
- main.go: 添加 timeoutMiddleware 并在 setupRouter 中应用

## 具体要求
- REQ-1: 创建 timeoutMiddleware(timeout time.Duration) 中间件工厂
- REQ-2: 使用 http.TimeoutHandler 或 context.WithTimeout
- REQ-3: 默认超时 5 秒
- REQ-4: 超时返回 503 Service Unavailable

## 约束
- setupRouter() 函数签名不可更改
- 正常请求不受影响

## 测试策略
- 验证正常请求不超时
- 验证 slow 端点返回 503
- 验证响应体包含超时信息

## 不做什么
- 不添加请求级别的超时配置
- 不修改业务逻辑
