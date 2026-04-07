# Task: 添加事务性批量操作

## 目标
添加 POST /todos/batch 端点，支持在单个请求中执行批量 create/update/delete，操作全部成功或全部回滚。

## 变更范围
- main.go: 添加 /todos/batch 路由
- handlers.go: 添加批量操作 handler

## 具体要求
- REQ-1: POST /todos/batch 接受操作数组 {"operations":[{"action":"create","data":{...}}, ...]}
- REQ-2: 支持 action: "create"（需要 title）, "update"（需要 id + 字段）, "delete"（需要 id）
- REQ-3: 事务性：任一操作验证失败，全部不执行
- REQ-4: 成功返回 200 和所有操作结果
- REQ-5: 验证失败返回 400，指明哪个操作失败
- REQ-6: 最大批量大小 100
- REQ-7: 操作按数组顺序执行

## 约束
- setupRouter() 函数签名不变
- 纯 stdlib
- 批量操作在单次锁内完成

## 测试策略
- 验证批量创建
- 验证批量混合操作
- 验证失败时回滚
- 验证最大批量限制

## 不做什么
- 不实现异步批量
- 不实现部分成功模式
