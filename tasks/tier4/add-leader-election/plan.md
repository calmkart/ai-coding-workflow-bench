# Task: 实现 Leader Election (Lease 模式)

## 目标
实现基于内存 Lease 的 Leader Election：多个候选者竞争 leader，通过 lease 续期保持 leadership，lease 过期后其他候选者可竞选。

## 当前状态
- 没有 leader election 机制
- 多个 reconciler 可能同时工作

## 变更范围
- lease.go: Lease 资源和 LeaseStore 接口
- election.go: LeaderElector 实现

## 具体要求
- REQ-1: Lease 包含 HolderID, AcquireTime, RenewTime, LeaseDuration, Version
- REQ-2: LeaseStore 接口：Get/Create/Update（乐观锁 via version）
- REQ-3: InMemoryLeaseStore 实现，并发安全
- REQ-4: LeaderElector 包含 ID(标识自己), LeaseStore, Opts
- REQ-5: LeaderOpts: LeaseDuration, RenewDeadline, RetryPeriod, OnStartedLeading, OnStoppedLeading
- REQ-6: Run(ctx) 启动选举循环
- REQ-7: 竞选逻辑：
  - 尝试获取 lease（Create 或 Update 过期 lease）
  - 如果获取成功，成为 leader 并定期续期
  - 如果续期失败（版本冲突），降级
  - 其他候选者定期检查 lease 是否过期
- REQ-8: IsLeader() 返回当前是否是 leader
- REQ-9: OnStartedLeading 回调在获得 leadership 时调用
- REQ-10: OnStoppedLeading 回调在失去 leadership 时调用
- REQ-11: lease 过期时间 = renewTime + leaseDuration
- REQ-12: Update 使用乐观锁（version 不匹配返回错误）

## Lease 模式
```
1. 候选者 A 创建 lease (holder=A, version=1)
2. A 定期续期 (update renewTime, version++)
3. 如果 A 失败停止续期，lease 过期
4. 候选者 B 检测到 lease 过期，更新 holder=B
5. B 成为新 leader
```

## 约束
- 纯 stdlib
- 内存实现（不需要 etcd/Redis）
- 乐观并发控制

## 测试策略
- 单个候选者获得 leadership
- 多个候选者只有一个 leader
- leader 续期保持 leadership
- leader 停止后新 leader 产生
- 回调正确触发
- 并发安全
- 乐观锁冲突处理

## 不做什么
- 不实现分布式一致性
- 不实现 fencing token
- 不实现持久化
