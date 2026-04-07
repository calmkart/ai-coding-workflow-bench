# Task: 实现 Worker Pool

## 目标
实现固定大小的 worker pool。

## 变更范围
- pool.go: 实现 Pool 类型

## 具体要求
- REQ-1: NewPool(workers int) *Pool
- REQ-2: Submit(task func()) 提交任务
- REQ-3: Wait() 等待所有已提交任务完成
- REQ-4: Shutdown() 关闭 pool 不再接受新任务

## 约束
- 不使用外部依赖
- 固定 worker 数量

## 测试策略
- 验证任务执行
- 验证 Wait 等待完成
- 验证并发限制

## 不做什么
- 不添加任务优先级
- 不添加动态 worker 数
