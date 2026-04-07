# Task: 修复 WaitGroup 泄漏

## 目标
修复 goroutine 中 panic 导致 WaitGroup 永不 Done。

## 变更范围
- processor.go: 修复 WaitGroup 使用

## 具体要求
- REQ-1: 使用 defer wg.Done() 确保 Done 被调用
- REQ-2: recover panic 并记录到错误中
- REQ-3: 返回所有成功结果和第一个错误
- REQ-4: ProcessAll 不永久阻塞

## 约束
- 保持并发处理
- 不改变函数签名

## 测试策略
- 验证不死锁
- 验证 panic 被恢复
- 验证正常项处理成功

## 不做什么
- 不添加 context
- 不改变处理逻辑
