# Task: 修复 channel 死锁

## 目标
修复 unbuffered channel 导致的死锁问题。

## 变更范围
- pipeline.go: 修复 channel 使用

## 具体要求
- REQ-1: Process 函数不死锁
- REQ-2: 所有 items 都被处理
- REQ-3: 结果顺序可以不同于输入
- REQ-4: 使用 goroutine 正确分离 producer 和 consumer

## 约束
- 保持并发处理逻辑
- 不改变函数签名

## 测试策略
- 验证不死锁（有超时）
- 验证所有结果返回
- 验证结果正确

## 不做什么
- 不改变处理逻辑
- 不添加 context
