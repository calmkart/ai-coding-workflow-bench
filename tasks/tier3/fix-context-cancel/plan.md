# Task: 修复 context 取消不传播

## 目标
修复外层 context 取消不传播到内部 goroutine 的 bug。

## 变更范围
- coordinator.go: 修复 context 传播

## 具体要求
- REQ-1: Start(ctx) 使用传入的 ctx（不创建新的 background context）
- REQ-2: 所有内部 goroutine 使用派生的 context
- REQ-3: 外层 cancel 后所有 goroutine 应在合理时间内退出
- REQ-4: Worker 的 select 中必须有 ctx.Done() 分支
- REQ-5: Coordinator.Wait() 等待所有 goroutine 退出
- REQ-6: 修复后无 goroutine 泄漏
- REQ-7: 支持设置子任务超时（每个 worker 独立超时）

## 约束
- 保持现有功能不变
- 修复 context 传播，不改变接口

## 测试策略
- 验证 cancel 后所有 goroutine 退出
- 验证无 goroutine 泄漏
- 验证子任务超时

## 不做什么
- 不重写 coordinator
- 不添加新功能
