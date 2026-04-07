# Task: 实现一致性哈希环

## 目标
实现一致性哈希环，支持虚拟节点、节点增删、最小迁移量。用于分布式系统中将 key 映射到 node。

## 当前状态
- 只有空的类型定义

## 变更范围
- hashring.go: 完整实现

## 具体要求
- REQ-1: NewHashRing(replicas) 创建哈希环，replicas 是每个物理节点的虚拟节点数
- REQ-2: AddNode(node) 添加节点（创建 replicas 个虚拟节点）
- REQ-3: RemoveNode(node) 移除节点（移除其所有虚拟节点）
- REQ-4: GetNode(key) 返回 key 应该路由到的节点
- REQ-5: GetNodes(key, count) 返回 key 应该路由到的 count 个不同物理节点（副本）
- REQ-6: 使用 FNV-1a 或 CRC32 哈希
- REQ-7: 虚拟节点均匀分布在环上
- REQ-8: 添加/移除节点时，只有最少的 key 需要重新映射
- REQ-9: 空环返回空字符串
- REQ-10: GetNodes 如果请求数量大于物理节点数，返回所有物理节点
- REQ-11: NodeCount() 返回物理节点数
- REQ-12: 并发安全（读写锁保护）

## 约束
- 纯 stdlib（可用 hash/fnv 或 hash/crc32）
- 哈希环用排序数组 + 二分查找实现

## 测试策略
- 基本路由一致性
- 添加节点后大部分 key 不变
- 移除节点后大部分 key 不变
- 虚拟节点改善分布均匀性
- 多副本路由到不同物理节点
- 空环行为
- 并发安全

## 不做什么
- 不实现加权节点
- 不实现有界一致性哈希
