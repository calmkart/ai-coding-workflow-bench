# Task: 修复 DELETE 不存在资源的状态码

## 目标
修复 DELETE /todos/{id} 端点：当删除不存在的 todo 时应返回 404 Not Found，当前实现返回 500 Internal Server Error。

## 变更范围
- handlers.go: 修复 deleteTodo handler 的错误处理

## 具体要求
- REQ-1: DELETE 不存在的 ID 返回 404 Not Found
- REQ-2: DELETE 存在的 ID 仍返回 204 No Content
- REQ-3: DELETE 无效 ID（非数字）仍返回 400 Bad Request

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只修改 deleteTodo 的错误处理逻辑

## 测试策略
- 验证删除不存在的 ID 返回 404
- 验证删除存在的 ID 返回 204
- 验证删除无效 ID 返回 400

## 不做什么
- 不改变 API 路径或响应格式
- 不添加新的端点
- 不修改其他 handler
