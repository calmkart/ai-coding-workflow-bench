# Task: 添加批量创建端点

## 目标
添加 POST /todos/bulk 端点，一次创建多个 todo。

## 变更范围
- handlers.go: 添加 bulkCreateTodos handler
- main.go: 注册新路由

## 具体要求
- REQ-1: POST /todos/bulk 接受 JSON 数组
- REQ-2: 返回 201 Created 和创建的 todo 数组
- REQ-3: 空数组返回 400
- REQ-4: 数组中任一 todo 无效（如空 title）返回 400，不创建任何记录
- REQ-5: 最多一次创建 100 个

## 约束
- setupRouter() 函数签名不可更改
- 现有端点不变

## 测试策略
- 验证批量创建返回 201
- 验证创建的 todo 都有 ID
- 验证空数组返回 400
- 验证超过 100 个返回 400

## 不做什么
- 不添加批量删除或更新
- 不改变单个创建端点
