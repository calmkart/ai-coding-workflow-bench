# Task: 实现批处理器

## 目标
实现批处理器：攒够 N 条或超时 T 后批量处理。

## 变更范围
- batcher.go: 需要实现的批处理框架

## 具体要求
- REQ-1: BatchProcessor[T] 泛型批处理器
- REQ-2: Add(item T) 添加项目到批次
- REQ-3: 达到 batchSize 自动触发 flush
- REQ-4: 达到 flushInterval 超时也触发 flush
- REQ-5: Flush() 手动触发处理
- REQ-6: Close() 处理剩余项目并关闭
- REQ-7: process 回调接收批次切片
- REQ-8: 并发安全，多个 goroutine 可同时 Add

## 约束
- 纯 stdlib + 泛型
- 不丢失数据

## 测试策略
- 验证按大小触发
- 验证按超时触发
- 验证 Close 处理剩余
- 验证并发安全

## 不做什么
- 不实现重试
- 不实现背压
