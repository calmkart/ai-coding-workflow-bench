# Task: 修复 Status 更新冲突

## 目标
添加版本号和冲突重试机制。

## 变更范围
- types.go: 添加 ResourceVersion 字段
- reconciler.go: Store.Put 检查版本号

## 具体要求
- REQ-1: Resource 添加 ResourceVersion int 字段
- REQ-2: Store.Put 检查版本号，冲突时返回 ErrConflict
- REQ-3: Reconciler 在冲突时重新获取资源再重试
- REQ-4: 最多重试 3 次

## 约束
- 不使用外部依赖

## 测试策略
- 验证版本号递增
- 验证冲突检测
- 验证重试成功

## 不做什么
- 不添加分布式锁
- 不添加 watch 机制
