# Task: 实现信号量

## 目标
实现基于 channel 的信号量。

## 变更范围
- semaphore.go: 实现 Semaphore

## 具体要求
- REQ-1: NewSemaphore(n int) *Semaphore
- REQ-2: Acquire() 阻塞获取许可
- REQ-3: TryAcquire() bool 非阻塞获取
- REQ-4: Release() 释放许可
- REQ-5: 基于 buffered channel 实现

## 约束
- 不使用外部依赖
- 不使用 sync 包

## 测试策略
- 验证 Acquire/Release 基本流程
- 验证 TryAcquire 非阻塞
- 验证并发限制

## 不做什么
- 不添加超时获取
- 不添加动态调整
