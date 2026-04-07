# Task: 添加重试退避

## 目标
为 reconciler 失败重试添加指数退避策略。

## 变更范围
- reconciler.go: 添加 Backoff 类型和退避逻辑

## 具体要求
- REQ-1: Backoff{InitialDelay, MaxDelay, Factor} 类型
- REQ-2: 连续失败时延迟指数增长
- REQ-3: 成功时重置退避
- REQ-4: 延迟不超过 MaxDelay

## 约束
- 不使用外部依赖
- 不添加 jitter

## 测试策略
- 验证退避延迟递增
- 验证不超过最大值
- 验证成功后重置

## 不做什么
- 不添加 jitter
- 不添加自定义策略接口
