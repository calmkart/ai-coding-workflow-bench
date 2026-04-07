# Task: 添加字符串截断函数

## 目标
添加 Truncate 函数，按单词边界截断字符串并在末尾添加 "..."。

## 变更范围
- strkit.go: 添加 Truncate 函数

## 具体要求
- REQ-1: 如果字符串长度 <= maxLen，原样返回
- REQ-2: 如果字符串长度 > maxLen，截断到最近的单词边界 + "..."
- REQ-3: 截断后总长度（含 "..."）不超过 maxLen
- REQ-4: maxLen < 4 时直接截断到 maxLen 长度（无 "..."）
- REQ-5: 空字符串返回空字符串

## 约束
- 函数签名: func Truncate(s string, maxLen int) string
- 使用标准库

## 测试策略
- 验证短字符串不截断
- 验证长字符串在单词边界截断
- 验证截断后加 "..."
- 验证空字符串
- 验证 maxLen 小于 4

## 不做什么
- 不处理 HTML 或其他标记语言
- 不做多语言分词
