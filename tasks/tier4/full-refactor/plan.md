# Task: 完整重构 God Handler

## 目标
当前项目是一个 500+ 行的 God Handler：所有逻辑（路由、业务、数据、配置、错误处理）都在一个文件中。需要完整重构为分层架构。

## 当前问题
- 所有代码在一个 handler.go 文件中（500+ 行）
- 业务逻辑与 HTTP 处理混合
- 硬编码配置（端口、分页大小等）
- 错误处理不一致（有些返回 JSON，有些返回纯文本）
- 无中间件（无 logging、无 recovery、无 request-id）
- 无类型化错误
- panic 无恢复

## 变更范围
将 handler.go 重构为：
- config.go: 配置管理
- errors.go: 类型化错误定义
- repository.go: 数据存储层（TodoRepository 接口 + 实现）
- service.go: 业务逻辑层（TodoService）
- handlers.go: HTTP handler 层（只做请求解析、调用 service、响应格式化）
- middleware.go: 中间件（logging, recovery, request-id）
- main.go: 组装和启动

## 具体要求
- REQ-1: 分离为 repository/service/handler 三层
- REQ-2: TodoRepository 接口：Create/Get/List/Update/Delete
- REQ-3: TodoService 调用 repository，包含业务逻辑（验证等）
- REQ-4: Handler 只做 HTTP 相关工作
- REQ-5: 实现 AppError 类型化错误（Code, Message, HTTPStatus）
- REQ-6: 添加 Recovery 中间件（panic 恢复为 500）
- REQ-7: 添加 Logging 中间件（记录请求方法、路径、耗时、状态码）
- REQ-8: 添加 RequestID 中间件（生成 X-Request-ID header）
- REQ-9: 提取配置到 Config struct（Port, PageSize, MaxTitleLength）
- REQ-10: 所有错误返回统一 JSON 格式 {"error": "message", "code": "NOT_FOUND"}
- REQ-11: API 行为完全不变（同样的路由、请求、响应）
- REQ-12: setupRouter() 签名不变

## 约束
- setupRouter() 签名不变
- 纯 stdlib
- 外部 API 行为完全兼容
- 重构不能改变功能

## 测试策略
- 所有 CRUD 操作行为不变
- 错误返回统一 JSON 格式
- Recovery 中间件拦截 panic
- RequestID 在响应头中
- 分页工作正确
- 并发安全

## 不做什么
- 不添加新功能
- 不改变 API 路由
- 不引入依赖注入框架
