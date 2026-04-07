# Task: 实现泛型 Result 类型

## 目标
实现 Result[T] 泛型类型，提供 Map/FlatMap 等函数式操作。

## 变更范围
- result.go: 实现 Result[T] 类型和方法

## 具体要求
- REQ-1: Result[T]{Value T, Err error}
- REQ-2: Ok[T](v) 创建成功结果
- REQ-3: Fail[T](err) 创建失败结果
- REQ-4: IsOk() bool, IsErr() bool 判断
- REQ-5: Map(fn func(T) T) Result[T] 变换（失败时短路）
- REQ-6: FlatMap(fn func(T) Result[T]) Result[T] 链式变换
- REQ-7: Unwrap() T（失败时 panic）
- REQ-8: UnwrapOr(def T) T（失败时返回默认值）

## 约束
- 不使用外部依赖

## 测试策略
- 验证 Ok/Fail 创建
- 验证 Map 变换
- 验证 FlatMap 链式
- 验证 Unwrap panic

## 不做什么
- 不添加 async 操作
