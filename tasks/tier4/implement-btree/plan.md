# Task: 实现 B-tree 数据结构

## 目标
实现完整的泛型 B-tree，支持 Insert/Search/Delete/Range 操作。B-tree 是一种自平衡树，广泛用于数据库索引。

## 当前状态
- 只有空的类型定义和接口骨架
- 节点结构不完整

## 变更范围
- btree.go: 完整实现 B-tree

## 具体要求
- REQ-1: 泛型实现 BTree[K, V]，K 可比较通过 less 函数
- REQ-2: NewBTree(order, less) 创建指定阶数的 B-tree
- REQ-3: Insert(key, value) 插入键值对，重复 key 更新 value
- REQ-4: Search(key) 返回 (value, found)
- REQ-5: Delete(key) 删除键值对，返回是否存在
- REQ-6: Range(from, to) 返回 [from, to] 范围内的所有键值对，有序
- REQ-7: Len() 返回元素总数
- REQ-8: 插入时节点满则分裂（split）
- REQ-9: 删除时节点不足则合并（merge）或从兄弟节点借位（borrow）
- REQ-10: 维护 B-tree 性质：所有叶子在同一层，非根节点至少 ceil(order/2)-1 个键
- REQ-11: Min() 和 Max() 返回最小/最大键值对
- REQ-12: InOrder() 返回所有键值对的有序遍历

## B-tree 性质（阶 m）
- 每个节点最多 m 个子节点
- 每个非根节点至少 ceil(m/2) 个子节点
- 根节点至少 2 个子节点（如果非叶子）
- 每个节点最多 m-1 个键
- 所有叶子在同一层

## 约束
- 纯 stdlib，使用 Go 泛型
- 不使用 sort 包（自己实现节点内二分查找）

## 测试策略
- 基本插入和查找
- 大量插入触发分裂
- 删除后合并/借位
- Range 查询正确
- 重复 key 更新 value
- 空树操作
- 大量数据（1000+）性能合理
- InOrder 遍历有序

## 不做什么
- 不实现 B+ tree（只实现 B-tree）
- 不实现持久化到磁盘
- 不实现并发安全（单线程使用）
