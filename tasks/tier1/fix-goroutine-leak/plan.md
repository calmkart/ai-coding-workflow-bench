# Task: 修复 goroutine 泄漏

## 目标
修复 worker pool 的 Stop 方法：当前 Stop 不会通知 worker goroutine 退出，导致 goroutine 泄漏。

## 变更范围
- pool.go: 添加 context 或 close channel 机制使 Stop 能终止 worker

## 具体要求
- REQ-1: Stop 调用后所有 worker goroutine 退出
- REQ-2: Stop 后 Submit 不应 panic
- REQ-3: Start 启动指定数量的 worker

## 约束
- Pool struct 的公开方法签名不可更改
- 使用标准库（context 或 close channel）

## 测试策略
- 验证 Stop 后 goroutine 计数回到基线
- 验证正常 Submit 执行工作
- 验证 Stop 后再 Submit 不 panic

## 不做什么
- 不添加新的公开方法
- 不改变任务执行语义
