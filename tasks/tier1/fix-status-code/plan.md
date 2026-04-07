# Task: 修复 POST 返回状态码

## 目标
修复 POST /todos 端点的 HTTP 状态码：创建成功后应返回 201 Created，当前实现返回 200 OK。

## 变更范围
- handlers.go: 修复 createTodo handler 的返回状态码

## 具体要求
- REQ-1: POST /todos 成功创建后返回 HTTP 201 Created
- REQ-2: 响应体仍然返回创建的 todo JSON 对象
- REQ-3: 错误情况（无效 body）仍返回 400

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只修改 createTodo 的状态码逻辑

## 测试策略
- 验证 POST /todos 返回 201
- 验证响应体包含创建的 todo
- 验证无效 body 返回 400

## 不做什么
- 不改变 API 路径或响应格式
- 不添加新的端点
- 不重构其他 handler
