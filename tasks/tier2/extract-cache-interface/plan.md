# Task: 提取 Cache 接口

## 目标
从具体的 MapCache 实现提取 Cache[K,V] 泛型接口。

## 变更范围
- cachebox.go: 提取 Cache 接口，MapCache 实现该接口

## 具体要求
- REQ-1: 定义 Cache[K comparable, V any] 接口
- REQ-2: 接口方法: Get(K) (V, bool), Set(K, V), Delete(K), Len() int
- REQ-3: MapCache 实现 Cache 接口
- REQ-4: NewMapCache 返回 Cache 接口类型

## 约束
- 不改变 MapCache 的行为
- 保持线程安全

## 测试策略
- 验证 MapCache 实现 Cache 接口
- 验证基本 CRUD 操作

## 不做什么
- 不添加 TTL 功能
- 不添加 LRU 淘汰
