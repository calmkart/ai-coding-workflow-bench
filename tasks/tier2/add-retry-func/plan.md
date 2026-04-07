# Task: 实现重试函数

## 目标
实现通用的重试函数，带指数退避。

## 变更范围
- retry.go: 实现 Retry 函数

## 具体要求
- REQ-1: Retry(fn, maxAttempts, baseDelay) error
- REQ-2: 指数退避: delay = baseDelay * 2^attempt
- REQ-3: 成功时立即返回 nil
- REQ-4: 所有尝试失败后返回最后一个错误
- REQ-5: maxAttempts < 1 时 panic

## 约束
- 不使用外部依赖
- 不添加 jitter

## 测试策略
- 验证首次成功
- 验证第 N 次成功
- 验证全部失败返回最后错误
- 验证尝试次数正确

## 不做什么
- 不添加 context 取消
- 不添加 jitter
