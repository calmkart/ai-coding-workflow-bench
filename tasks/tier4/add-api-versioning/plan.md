# Task: 添加 API 版本控制 (v1/v2)

## 目标
当前 API 只有一个版本，路由没有版本前缀。需要添加 /v1/ 和 /v2/ 版本路由，v2 支持新字段（priority, tags），同时保持 v1 向后兼容。

## 当前问题
- 路由没有版本前缀（/todos 而非 /v1/todos）
- Todo 结构体只有基础字段
- 无法在不破坏现有客户端的情况下添加新字段

## 变更范围
- main.go: setupRouter 配置 v1 和 v2 路由
- handlers.go: v1 handler 保持现有响应格式
- handlers_v2.go: v2 handler 支持新字段 priority(int) 和 tags([]string)
- store.go: 统一存储层，内部使用完整 Todo 结构

## 具体要求
- REQ-1: /v1/todos 端点保持现有行为和响应格式（id, title, done）
- REQ-2: /v2/todos 端点响应包含 priority(int, 0-5) 和 tags([]string)
- REQ-3: v1 和 v2 共享同一 TodoStore
- REQ-4: v1 创建的 todo 在 v2 查看时 priority=0, tags=[]
- REQ-5: v2 创建的带 priority/tags 的 todo 在 v1 查看时只显示 id/title/done
- REQ-6: /todos（无版本前缀）返回 301 重定向到 /v1/todos
- REQ-7: /v2/todos POST 验证 priority 范围 0-5
- REQ-8: /v2/todos 支持 ?priority=N 按优先级过滤
- REQ-9: /v2/todos 支持 ?tag=X 按标签过滤
- REQ-10: /health 不需要版本前缀
- REQ-11: v2 响应包含 api_version: "v2" 字段

## 约束
- setupRouter() 签名不变
- 纯 stdlib
- v1 响应格式绝对不能变

## 测试策略
- v1 CRUD 完全兼容
- v2 CRUD 支持新字段
- 跨版本数据一致性
- 版本前缀路由正确
- 无前缀重定向
- v2 过滤功能
- priority 验证

## 不做什么
- 不实现 Accept header 版本协商
- 不实现 v1 deprecation 警告
- 不实现超过 v2 的版本
