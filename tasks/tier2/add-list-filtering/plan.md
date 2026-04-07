# Task: 添加列表过滤

## 目标
为 GET /todos 添加 ?done=true/false 查询参数过滤。

## 变更范围
- handlers.go: 在 listTodos 中添加过滤逻辑

## 具体要求
- REQ-1: 支持 ?done=true 只返回已完成
- REQ-2: 支持 ?done=false 只返回未完成
- REQ-3: 不传 done 参数返回全部
- REQ-4: done 参数无效值返回 400

## 约束
- setupRouter() 函数签名不可更改
- 分页逻辑在过滤之后应用

## 测试策略
- 验证 ?done=true 只返回已完成项
- 验证 ?done=false 只返回未完成项
- 验证无 done 参数返回全部
- 验证无效 done 值返回 400

## 不做什么
- 不添加其他过滤条件
- 不改变响应格式
