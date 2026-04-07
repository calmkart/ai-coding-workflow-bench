# Task: 提取 Logger 接口

## 目标
从硬编码的 fmt.Printf 提取 Logger 接口并注入。

## 变更范围
- processor.go: 定义 Logger 接口，Processor 接受 Logger

## 具体要求
- REQ-1: 定义 Logger 接口: Info(msg, args...), Error(msg, args...)
- REQ-2: 提供 StdLogger 默认实现（使用 log 标准库）
- REQ-3: Processor 通过构造函数接受 Logger
- REQ-4: 替换所有 fmt.Printf 为 Logger 调用

## 约束
- 不使用外部依赖
- Processor 功能不变

## 测试策略
- 验证 Logger 接口可替换
- 验证 StdLogger 正确输出
- 验证 Processor 使用注入的 Logger

## 不做什么
- 不添加日志级别（只有 Info/Error）
- 不添加结构化日志
