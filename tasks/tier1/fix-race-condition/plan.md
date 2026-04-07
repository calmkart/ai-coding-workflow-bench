# Task: 修复计数器竞态条件

## 目标
修复 Counter struct 的并发安全问题：Inc 和 Get 方法无锁保护，并发调用导致数据竞态。

## 变更范围
- counter.go: 为 Counter 添加 sync.Mutex 或使用 sync/atomic

## 具体要求
- REQ-1: Inc 方法并发安全
- REQ-2: Get 方法并发安全
- REQ-3: 并发 Inc 后 Get 返回正确的计数值
- REQ-4: go test -race 不报错

## 约束
- Counter 公开方法签名不可更改
- 使用标准库（sync.Mutex 或 sync/atomic）

## 测试策略
- 用 -race flag 测试并发 Inc + Get
- 验证并发 Inc 后计数正确

## 不做什么
- 不添加新方法
- 不改变 Counter 的公开接口
