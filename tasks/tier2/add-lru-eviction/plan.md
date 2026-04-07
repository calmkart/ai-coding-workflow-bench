# Task: 添加 LRU 淘汰

## 目标
为 cache 添加 LRU 淘汰策略。

## 变更范围
- cachebox.go: 添加 LRUCache 实现

## 具体要求
- REQ-1: LRUCache 有固定容量
- REQ-2: 超过容量时淘汰最近最少使用的项
- REQ-3: Get 操作更新访问时间
- REQ-4: 使用 container/list 实现双向链表

## 约束
- 线程安全
- 不使用外部依赖

## 测试策略
- 验证容量限制
- 验证 LRU 淘汰顺序
- 验证 Get 更新访问时间

## 不做什么
- 不添加 TTL 功能
- 不添加回调通知
