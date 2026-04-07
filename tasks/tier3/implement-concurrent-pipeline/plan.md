# Task: 实现三阶段并发管道

## 目标
实现 producer -> transform -> consumer 三阶段 channel 管道。

## 变更范围
- pipeline.go: 需要实现的管道框架

## 具体要求
- REQ-1: Producer goroutine 生成数据发到 channel
- REQ-2: Transform stage 从 channel 读取、转换、发到下一个 channel
- REQ-3: Consumer stage 从 channel 读取、处理、收集结果
- REQ-4: Context 取消时所有 stage 停止
- REQ-5: Transform 错误传播到结果
- REQ-6: 支持配置 transform 并发数
- REQ-7: 正确关闭 channel（生产者关闭自己的输出 channel）
- REQ-8: 无 goroutine 泄漏

## 约束
- 纯 stdlib
- 泛型实现

## 测试策略
- 验证数据完整流过三阶段
- 验证 context 取消
- 验证并发 transform
- 验证错误传播

## 不做什么
- 不实现背压
- 不实现动态扩缩
