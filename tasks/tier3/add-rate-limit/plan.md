# Task: 添加 per-IP 令牌桶限速中间件

## 目标
添加 per-IP 的令牌桶限速中间件。每个 IP 有独立的令牌桶，超限返回 429。

## 变更范围
- main.go: 在 setupRouter 中使用限速中间件
- 新增 ratelimit.go: 令牌桶和限速中间件实现

## 具体要求
- REQ-1: 实现 TokenBucket struct {rate, burst, tokens, lastRefill}
- REQ-2: 实现 RateLimiter struct 维护 per-IP 的 TokenBucket
- REQ-3: 实现 RateLimitMiddleware(limiter, next) http.Handler
- REQ-4: 超限返回 429 {"error":"rate limit exceeded"} + Retry-After header
- REQ-5: 正常请求添加 X-RateLimit-Remaining header
- REQ-6: IP 从 r.RemoteAddr 提取（去掉端口）
- REQ-7: 默认 10 req/s，burst 20
- REQ-8: 过期清理：超过 5 分钟无请求的 IP 清除其桶

## 约束
- setupRouter() 函数签名不变
- 纯 stdlib，不用外部限速库
- 并发安全

## 测试策略
- 验证正常速率请求通过
- 验证超限返回 429
- 验证 Retry-After header
- 验证不同 IP 独立限速

## 不做什么
- 不实现分布式限速
- 不实现按路径限速
