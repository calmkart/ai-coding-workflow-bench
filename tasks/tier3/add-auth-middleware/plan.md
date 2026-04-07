# Task: 添加 Bearer Token 认证中间件

## 目标
添加简单的 Bearer token 认证中间件，保护所有 CRUD 端点。Health 端点不需要认证。

## 变更范围
- main.go: 在 setupRouter 中添加中间件
- 新增 middleware.go: 认证中间件

## 具体要求
- REQ-1: 实现 AuthMiddleware(validTokens map[string]string, next http.Handler) http.Handler
- REQ-2: 从 Authorization header 解析 Bearer token
- REQ-3: 无 Authorization header 返回 401
- REQ-4: 无效 token 返回 401 {"error":"unauthorized"}
- REQ-5: 有效 token 将用户信息存入 context，传给下游
- REQ-6: GET /health 不需要认证
- REQ-7: 支持配置多个有效 token（token -> username 映射）
- REQ-8: POST/GET/PUT/DELETE /todos* 都需要认证

## 约束
- setupRouter() 函数签名不变
- 纯 stdlib，不用 JWT 库
- token 是简单字符串比较

## 测试策略
- 验证无 token 返回 401
- 验证无效 token 返回 401
- 验证有效 token 可以访问
- 验证 health 不需要 token

## 不做什么
- 不实现 JWT
- 不实现 OAuth
- 不持久化 token
