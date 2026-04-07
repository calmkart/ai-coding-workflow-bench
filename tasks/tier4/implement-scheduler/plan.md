# Task: 实现任务调度器（cron 风格）

## 目标
实现任务调度器：解析简化版 cron 表达式，按计划定时执行任务，支持取消。

## 当前状态
- 只有空的接口定义

## 变更范围
- cron.go: cron 表达式解析
- scheduler.go: 调度器实现

## 具体要求
- REQ-1: 简化版 cron 格式：`second minute hour` (三个字段)
- REQ-2: 字段支持：*(每), 具体数字, 逗号分隔(多值), 斜杠(间隔)
  - `*` = 每秒/分/时
  - `5` = 第 5 秒/分/时
  - `1,3,5` = 第 1, 3, 5
  - `*/10` = 每 10 秒/分/时
- REQ-3: ParseCron(expr) 解析表达式返回 CronExpr
- REQ-4: CronExpr.Next(from time.Time) 返回下一次触发时间
- REQ-5: CronExpr.Matches(t time.Time) 检查时间是否匹配
- REQ-6: NewScheduler() 创建调度器
- REQ-7: Schedule(name, cron, fn) 添加定时任务，返回 ID
- REQ-8: Cancel(id) 取消任务
- REQ-9: Start(ctx) 启动调度器（阻塞），ctx 取消时停止
- REQ-10: 同一秒内只触发一次
- REQ-11: ListJobs() 返回所有已注册的任务信息
- REQ-12: 调度器并发安全

## Cron 表达式示例
```
"* * *"      = 每秒
"*/5 * *"    = 每 5 秒
"0 * *"      = 每分钟第 0 秒
"0 */30 *"   = 每 30 分钟
"0 0 */2"    = 每 2 小时
"0,30 * *"   = 每分钟第 0 秒和第 30 秒
```

## 约束
- 纯 stdlib
- 简化版 cron（只有 second/minute/hour 三个字段）
- 不需要日期/星期/月份

## 测试策略
- 解析各种 cron 表达式
- Next 计算正确
- Matches 匹配正确
- 定时执行
- 取消任务
- 并发安全
- 无效表达式错误

## 不做什么
- 不实现完整 5 字段 cron
- 不实现持久化
- 不实现重试
