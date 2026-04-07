# Task: 实现 Saga 模式处理分布式事务

## 目标
当前 API 创建项目时需要多步操作（创建项目、创建关联 Todo、更新项目状态），但步骤间没有事务保证。实现 Saga 模式：步骤定义、顺序执行、失败补偿回滚。

## 当前问题
- 创建项目是多步操作但无事务
- 中间步骤失败后前面的步骤不回滚
- 项目可能处于部分创建的不一致状态

## 变更范围
- saga.go: Saga 引擎（SagaStep 接口、执行器、补偿逻辑）
- handlers.go: 原有 todo handler + 新的项目 handler 使用 Saga
- main.go: setupRouter 集成

## 具体要求
- REQ-1: SagaStep 接口有 Forward(ctx) error 和 Compensate(ctx) error 两个方法
- REQ-2: SagaStep 有 Name() string 方法标识步骤
- REQ-3: Saga.Execute 顺序执行所有步骤
- REQ-4: 任何步骤 Forward 失败，反向执行已完成步骤的 Compensate
- REQ-5: 补偿失败记录到 SagaResult.CompensationErrors（不中断其他补偿）
- REQ-6: SagaResult 包含：Success bool, FailedStep string, Error string, CompensationErrors []string, StepsCompleted []string
- REQ-7: POST /projects 使用 Saga 创建项目（步骤：创建项目 → 创建关联 todos → 更新项目状态为 active）
- REQ-8: GET /projects 列出所有项目
- REQ-9: GET /projects/{id} 获取项目详情（含关联 todo）
- REQ-10: 支持 Saga 日志：GET /sagas 返回执行历史
- REQ-11: Saga 支持 context 取消
- REQ-12: 每个步骤有超时（默认 5 秒）

## Project 结构
```go
type Project struct {
    ID     int      `json:"id"`
    Name   string   `json:"name"`
    Status string   `json:"status"` // "pending", "active", "failed"
    TodoIDs []int   `json:"todo_ids"`
}
```

## 约束
- setupRouter() 签名不变
- 纯 stdlib
- Saga 步骤同步顺序执行

## 测试策略
- 所有步骤成功 → 项目创建完成
- 中间步骤失败 → 补偿回滚之前的步骤
- 补偿也失败 → 记录补偿错误
- Saga 执行历史可查询
- context 取消中断执行
- 原有 todo CRUD 不受影响

## 不做什么
- 不实现异步 Saga（choreography）
- 不实现 Saga 持久化（仅内存）
- 不实现并行步骤执行
