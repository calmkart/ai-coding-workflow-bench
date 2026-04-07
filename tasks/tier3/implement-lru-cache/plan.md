# Task: 实现完整 LRU Cache

## 目标
实现完整的 LRU Cache：Get/Set/Delete/Len，支持容量限制、LRU 淘汰策略、TTL 过期和并发安全。

## 变更范围
- cache.go: 当前简单实现，需要完善

## 具体要求
- REQ-1: LRUCache[K, V] 使用双向链表 + map 实现 O(1) 操作
- REQ-2: Get(key) (V, bool) — 存在则返回并移到最近使用
- REQ-3: Set(key, value) — 设置并移到最近使用，超容量淘汰最久未使用
- REQ-4: Delete(key) bool — 删除指定键
- REQ-5: Len() int — 返回当前元素数
- REQ-6: TTL 支持：过期的元素自动不返回
- REQ-7: 并发安全（sync.RWMutex）
- REQ-8: Keys() []K — 返回所有有效键（未过期的）

## 约束
- 纯 stdlib
- 泛型实现
- 双向链表自行实现（不用 container/list 亦可用）

## 测试策略
- 验证 LRU 淘汰顺序
- 验证 TTL 过期
- 验证并发安全
- 验证容量限制

## 不做什么
- 不实现分布式缓存
- 不实现序列化
