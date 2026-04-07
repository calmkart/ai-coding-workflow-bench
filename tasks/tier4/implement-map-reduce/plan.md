# Task: 实现并行 MapReduce 框架

## 目标
实现泛型的并行 MapReduce 框架：Map 阶段并行处理输入，Shuffle 阶段按 key 分组，Reduce 阶段并行聚合。

## 当前状态
- 只有空的接口定义

## 变更范围
- mapreduce.go: 完整实现

## 具体要求
- REQ-1: Mapper 接口：Map(input In) []KeyValue[K, V]
- REQ-2: Reducer 接口：Reduce(key K, values []V) Out
- REQ-3: MapReduceJob 组合 Mapper、Reducer、worker 数
- REQ-4: Run 执行完整的 MapReduce 流程，返回 map[K]Out
- REQ-5: Map 阶段并行：将输入分配给 workers 个 goroutine
- REQ-6: Shuffle 阶段：按 key 分组所有 map 输出
- REQ-7: Reduce 阶段并行：将不同 key 分配给 workers 个 goroutine
- REQ-8: 支持 context 取消
- REQ-9: Map/Reduce 函数 panic 时不会导致整个 Job 崩溃
- REQ-10: 空输入返回空 map
- REQ-11: KeyValue 辅助类型包含 Key 和 Value
- REQ-12: 提供 MapFunc/ReduceFunc 便捷包装（从函数创建 Mapper/Reducer）

## MapReduce 流程
```
Input → [Map Phase] → Shuffle(group by key) → [Reduce Phase] → Output

Map Phase:   input_i → [(k1,v1), (k2,v2), ...]  (parallel)
Shuffle:     group {k1: [v1,v3], k2: [v2,v4]}
Reduce Phase: (k, [values]) → result            (parallel)
```

## 约束
- 纯 stdlib，使用 Go 泛型
- Map 和 Reduce 都并行执行

## 测试策略
- 单词计数（经典 MapReduce）
- 并行执行验证
- 空输入
- 大量数据
- context 取消
- panic 安全

## 不做什么
- 不实现分布式 MapReduce
- 不实现中间结果持久化
- 不实现 Combiner
