# Task: 添加泛型 Min/Max 函数

## 目标
添加泛型 Min 和 Max 函数，支持所有 cmp.Ordered 类型。

## 变更范围
- strkit.go: 添加 Min 和 Max 函数

## 具体要求
- REQ-1: Min 返回两个值中较小的那个
- REQ-2: Max 返回两个值中较大的那个
- REQ-3: 支持 int、float64、string 等 cmp.Ordered 类型
- REQ-4: 两值相等时返回任一（确定性即可）

## 约束
- 使用 Go 1.22+ 泛型语法
- 使用 cmp.Ordered 约束（标准库 cmp 包）
- 不引入外部依赖

## 测试策略
- 验证 int 类型的 Min/Max
- 验证 float64 类型的 Min/Max
- 验证 string 类型的 Min/Max
- 验证相等值

## 不做什么
- 不添加可变参数版本
- 不添加 slice 版本的 Min/Max
