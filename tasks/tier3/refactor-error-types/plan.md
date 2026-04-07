# Task: 重构为类型化错误 AppError

## 目标
将分散的 http.Error 字符串重构为统一的 AppError 类型，包含错误码、消息和 HTTP 状态码。

## 变更范围
- handlers.go: 将 http.Error 替换为 AppError
- 新增 errors.go: AppError 类型定义和常见错误
- 新增 middleware.go: 错误处理中间件

## 具体要求
- REQ-1: 定义 AppError struct {Code string, Message string, HTTPStatus int}，实现 error 接口
- REQ-2: 定义常见错误变量：ErrNotFound, ErrBadRequest, ErrInvalidID, ErrInvalidBody, ErrTitleRequired
- REQ-3: 统一错误响应格式 {"error":{"code":"NOT_FOUND","message":"resource not found"}}
- REQ-4: Handler 返回 error，由中间件统一处理错误响应
- REQ-5: 定义 AppHandler 类型 = func(w, r) error，适配为 http.HandlerFunc
- REQ-6: 未知错误返回 500 {"error":{"code":"INTERNAL","message":"internal server error"}}
- REQ-7: 所有现有错误场景保持相同的 HTTP 状态码

## 约束
- setupRouter() 函数签名不变
- 错误响应格式统一
- 纯 stdlib

## 测试策略
- 验证 404 返回统一格式
- 验证 400 返回统一格式
- 验证正常请求不受影响

## 不做什么
- 不添加错误日志
- 不添加错误追踪
