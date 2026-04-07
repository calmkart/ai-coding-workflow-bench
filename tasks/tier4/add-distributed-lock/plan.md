# Task: 添加分布式锁（内存模拟）

## 目标
当前 API 并发操作同一 Todo 无业务级锁保护。实现内存模拟的分布式锁：Locker 接口、TTL 自动释放、owner 验证。

## 当前问题
- 只有 sync.RWMutex 保护整个 store，无细粒度锁
- 没有 TTL，如果客户端崩溃锁永远不释放
- 没有 owner 概念，任何人可以释放任何锁

## 变更范围
- lock.go: Locker 接口和 InMemoryLocker 实现
- middleware.go: 锁中间件保护写操作
- handlers.go: 现有 handler（可能需要小修改）
- main.go: 集成 Locker

## 具体要求
- REQ-1: Locker 接口：Acquire(key, owner, ttl) (Lock, error)
- REQ-2: Locker 接口：Release(key, owner) error
- REQ-3: Locker 接口：Extend(key, owner, ttl) error
- REQ-4: Lock 结构：Key, Owner, ExpiresAt, AcquiredAt
- REQ-5: TTL 到期自动释放（后台 goroutine 清理或惰性清理）
- REQ-6: 只有 owner 可以释放/续期自己的锁
- REQ-7: 尝试获取已被锁定的资源返回 409 Conflict
- REQ-8: PUT/DELETE /todos/{id} 需要先获取锁
- REQ-9: 提供 POST /locks 获取锁、DELETE /locks/{key} 释放锁、PUT /locks/{key} 续期
- REQ-10: InMemoryLocker 并发安全
- REQ-11: 获取锁时通过 X-Lock-Owner header 指定 owner
- REQ-12: 默认 TTL 30 秒，最大 300 秒

## Locker 接口
```go
type Locker interface {
    Acquire(key, owner string, ttl time.Duration) (*Lock, error)
    Release(key, owner string) error
    Extend(key, owner string, ttl time.Duration) error
}
```

## 约束
- setupRouter() 签名不变
- 纯 stdlib
- 内存实现（不需要 Redis/etcd）

## 测试策略
- 获取锁成功
- 重复获取同一锁返回 409
- TTL 到期后可重新获取
- 非 owner 无法释放锁
- 续期成功延长 TTL
- 并发获取锁只有一个成功
- 写操作需要锁
- 读操作不需要锁

## 不做什么
- 不实现分布式一致性（Raft/Paxos）
- 不实现锁等待队列
- 不实现可重入锁
