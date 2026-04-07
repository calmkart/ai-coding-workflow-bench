# Task: 添加 reconcile metrics 收集器

## 目标
添加 MetricsCollector 接口和 InstrumentedReconciler 包装器，收集 reconcile 次数/延迟/错误率。

## 变更范围
- 新增 metrics.go: MetricsCollector 接口和实现
- reconciler.go: InstrumentedReconciler 包装器

## 具体要求
- REQ-1: MetricsCollector 接口 {RecordReconcile(name, duration, error), RecordRequeue(name), Snapshot() MetricsSnapshot}
- REQ-2: InMemoryMetrics 实现 MetricsCollector
- REQ-3: InstrumentedReconciler 包装原有 Reconciler，在 Reconcile 前后收集 metrics
- REQ-4: MetricsSnapshot struct {TotalReconciles, TotalErrors, TotalRequeues int64, AvgDuration, P99Duration time.Duration}
- REQ-5: 按资源名统计
- REQ-6: 记录延迟直方图（用切片收集）
- REQ-7: 并发安全
- REQ-8: NewInstrumentedReconciler(inner Reconciler, collector MetricsCollector) Reconciler

## 约束
- Reconciler 接口不变
- 纯 stdlib，不依赖 Prometheus
- 接口模式，可替换实现

## 测试策略
- 验证 metrics 记录正确
- 验证延迟测量准确
- 验证错误计数
- 验证 snapshot 准确

## 不做什么
- 不集成 Prometheus
- 不实现 HTTP metrics 端点
