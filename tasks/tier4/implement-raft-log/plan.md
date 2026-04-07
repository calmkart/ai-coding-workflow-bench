# Task: 实现简化版 Raft 日志复制

## 目标
实现简化版 Raft 共识协议的日志复制部分。不需要实现完整的选举和心跳，只实现日志追加、提交、应用的核心机制。

## 当前状态
- 只有空的类型定义

## 变更范围
- raft.go: 完整实现

## 具体要求
- REQ-1: LogEntry 包含 Index, Term, Command(string)
- REQ-2: RaftNode 有角色：Leader, Follower
- REQ-3: NewRaftNode(id) 创建节点，初始为 Follower
- REQ-4: BecomeLeader(term) 切换为 Leader
- REQ-5: AppendEntries(entries, leaderTerm, prevIndex, prevTerm) 实现 Raft AppendEntries 逻辑
  - 如果 leaderTerm < currentTerm，拒绝（返回 false, currentTerm）
  - 如果 prevIndex/prevTerm 不匹配日志，拒绝
  - 删除冲突的尾部日志
  - 追加新条目
  - 更新 currentTerm
- REQ-6: Leader.Propose(command) 提议新的日志条目
- REQ-7: Commit(index) 标记日志条目已提交（commitIndex = index）
- REQ-8: Apply() 返回并标记已提交但未应用的条目（lastApplied → commitIndex）
- REQ-9: GetLog() 返回完整日志
- REQ-10: GetEntry(index) 返回指定索引的条目
- REQ-11: LastIndex() 和 LastTerm() 返回最后一条日志的信息
- REQ-12: 日志索引从 1 开始

## Raft 日志性质
- 如果两个日志在相同 index 有相同 term，则该 index 及之前的所有条目相同
- Leader 的日志永远不会被覆盖
- 只有已提交的条目才能被应用

## 约束
- 纯 stdlib
- 不需要网络通信（直接调用方法）
- 不需要实现选举/心跳
- 不需要并发安全

## 测试策略
- 基本日志追加
- prevIndex/prevTerm 不匹配拒绝
- 冲突日志截断
- term 过期拒绝
- commit 和 apply
- Leader propose
- 日志索引从 1 开始

## 不做什么
- 不实现 Leader 选举
- 不实现心跳
- 不实现快照
- 不实现网络通信
