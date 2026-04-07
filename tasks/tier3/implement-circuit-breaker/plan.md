# Task: 实现熔断器模式

## 目标
实现熔断器（Circuit Breaker）模式：Closed -> Open -> HalfOpen 状态机。

## 变更范围
- breaker.go: 需要完善的熔断器框架

## 具体要求
- REQ-1: 三种状态：Closed（正常）、Open（熔断）、HalfOpen（试探）
- REQ-2: Closed 状态：执行操作，记录失败次数，连续失败达阈值转 Open
- REQ-3: Open 状态：直接返回 ErrCircuitOpen，不执行操作
- REQ-4: Open 超时后转 HalfOpen，允许一次试探
- REQ-5: HalfOpen：试探成功转 Closed，失败转回 Open
- REQ-6: Options{MaxFailures int, Timeout time.Duration, HalfOpenMaxCalls int}
- REQ-7: State() 返回当前状态
- REQ-8: Reset() 手动重置为 Closed
- REQ-9: 并发安全

## 约束
- 纯 stdlib
- 不用外部熔断库

## 测试策略
- 验证 Closed 正常执行
- 验证失败达阈值后 Open
- 验证 Open 超时后 HalfOpen
- 验证 HalfOpen 恢复

## 不做什么
- 不实现滑动窗口
- 不实现半开放比例控制
