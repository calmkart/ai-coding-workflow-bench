# Task: 将 interface{} 重构为泛型

## 目标
将使用 interface{} 的集合工具函数重构为泛型版本，获得编译期类型安全。

## 变更范围
- collections.go: 将 interface{} 函数重构为泛型

## 具体要求
- REQ-1: Map[T, U](slice []T, fn func(T) U) []U
- REQ-2: Filter[T](slice []T, fn func(T) bool) []T
- REQ-3: Reduce[T, U](slice []T, init U, fn func(U, T) U) U
- REQ-4: Contains[T comparable](slice []T, elem T) bool
- REQ-5: Unique[T comparable](slice []T) []T
- REQ-6: GroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T
- REQ-7: 所有函数返回新切片，不修改输入
- REQ-8: nil 输入返回空切片（非 nil）

## 约束
- 纯 stdlib
- Go 1.22 泛型
- 保留旧函数名但改为泛型签名

## 测试策略
- 验证类型安全（编译期检查）
- 验证各函数正确性
- 验证 nil/empty 边界

## 不做什么
- 不添加并发版本
- 不添加排序
