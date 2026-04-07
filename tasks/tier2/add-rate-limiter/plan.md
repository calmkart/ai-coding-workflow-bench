# Task: 实现令牌桶限速器

## 目标
实现令牌桶 RateLimiter。

## 变更范围
- ratelimiter.go: 实现 RateLimiter

## 具体要求
- REQ-1: NewRateLimiter(rate float64, burst int)
- REQ-2: Allow() bool 非阻塞检查
- REQ-3: Wait() 阻塞等待获取令牌
- REQ-4: rate 表示每秒生成的令牌数
- REQ-5: burst 表示桶容量（最大突发）

## 约束
- 不使用外部依赖
- 使用 time 包实现令牌补充

## 测试策略
- 验证 burst 允许突发
- 验证超过速率被拒绝
- 验证令牌补充

## 不做什么
- 不添加 context 支持
- 不添加动态调整
