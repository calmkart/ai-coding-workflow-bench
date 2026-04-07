# Task: 修复分页 off-by-one bug

## 目标
修复 GET /todos 端点的分页逻辑：请求 page=2 时应该返回第二页的数据，当前实现返回了第三页。

## 变更范围
- handlers.go: 修复 listTodos handler 中的分页 offset 计算

## 具体要求
- REQ-1: offset 计算从 `page * pageSize` 改为 `(page - 1) * pageSize`
- REQ-2: page 参数从 1 开始计数（page=1 是第一页）
- REQ-3: page=0 或 page<0 时返回 400 Bad Request
- REQ-4: page 超出总页数时返回空列表（非 404）

## 约束
- setupRouter() 函数签名不可更改
- 所有现有测试必须继续通过
- 只修改分页逻辑，不做其他重构

## 测试策略
- 验证 page=1 返回前 10 条
- 验证 page=2 返回第 11-20 条
- 验证 page=0 返回 400
- 验证 page 超出范围返回空列表

## 不做什么
- 不改变 API 路径或响应格式
- 不添加新的端点
- 不重构其他 handler
