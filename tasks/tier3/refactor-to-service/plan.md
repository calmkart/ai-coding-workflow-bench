# Task: 重构为 Handler/Service/Repository 三层架构

## 目标
将巨型 handler 拆分为三层架构：Handler（HTTP 处理 + 序列化）→ Service（业务逻辑 + 验证）→ Repository（数据访问）。

## 变更范围
- handlers.go: 当前的巨型 handler，需要拆分
- 新增 service.go: TodoService 接口和实现
- 新增 repository.go: TodoRepository 接口和实现

## 具体要求
- REQ-1: 定义 TodoRepository 接口 { List(offset, limit) ([]Todo, int); GetByID(id) (*Todo, error); Create(todo) (*Todo, error); Delete(id) error }
- REQ-2: 定义 TodoService 接口 { ListTodos(page, pageSize) ([]Todo, int, error); GetTodo(id) (*Todo, error); CreateTodo(title) (*Todo, error); DeleteTodo(id) error }
- REQ-3: Handler 层只负责 HTTP 解析、调用 Service、JSON 响应
- REQ-4: Service 层负责参数验证和业务逻辑
- REQ-5: Repository 层负责数据存储和检索
- REQ-6: setupRouter() 签名不变，内部组装三层
- REQ-7: 所有 API 行为不变（路径、状态码、响应格式）

## 约束
- setupRouter() 函数签名不可更改
- 所有现有 API 端点行为不变
- 纯 stdlib，不用外部框架

## 测试策略
- 验证所有 CRUD 操作正常
- 验证分页正常
- 验证错误响应不变

## 不做什么
- 不添加新端点
- 不改变数据模型
- 不添加数据库支持
