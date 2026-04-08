# 任务目录 / Task Catalog

workflow-bench 内置 100 个 Go 编码任务，分为 4 个难度等级、5 种代码类型。本文档汇总每个任务的类型、预计时间和描述。

## 概览

| 等级 | 数量 | 预计时间 | 说明 |
|------|------|----------|------|
| Tier 1 | 20 | 5 分钟 | 简单：单点修复或小功能添加 |
| Tier 2 | 32 | 10 分钟 | 中等：新端点、中间件、接口提取 |
| Tier 3 | 29 | 15-20 分钟 | 复杂：完整模式实现、多文件重构 |
| Tier 4 | 19 | 25-30 分钟 | 高级：架构级重构、完整系统实现 |

**类型分布**：http-server (32) · library (24) · concurrency (15) · reconciler (14) · cli (15)

**总预计时间**：约 1,460 分钟（24+ 小时）

---

## Tier 1 — 简单任务（20 个，各 5 分钟）

| 任务 ID | 类型 | 描述 |
|---------|------|------|
| tier1/fix-handler-bug | http-server | 修复 GET /todos 分页逻辑：请求 page=2 时返回了第三页数据（off-by-one） |
| tier1/fix-status-code | http-server | 修复 POST /todos 状态码：创建成功应返回 201 Created，当前返回 200 |
| tier1/fix-json-content-type | http-server | 为所有 JSON 响应添加 Content-Type: application/json 头 |
| tier1/fix-delete-404 | http-server | 修复 DELETE /todos/{id}：删除不存在的 todo 应返回 404，当前返回 500 |
| tier1/fix-empty-body | http-server | 修复 POST /todos：请求体为空时 panic（nil pointer），应返回 400 |
| tier1/fix-query-param | http-server | 修复 GET /todos：page_size 为非法值时 panic，应返回 400 |
| tier1/add-health-check | http-server | 添加 GET /health 健康检查端点 |
| tier1/add-cors-headers | http-server | 添加 CORS 支持，允许前端跨域调用 |
| tier1/fix-string-reverse | library | 修复 Reverse 函数：按 byte 反转导致多字节 UTF-8 字符损坏，应按 rune 反转 |
| tier1/fix-contains-bug | library | 修复 ContainsAny：candidates 为空 slice 时应返回 false，当前返回 true |
| tier1/fix-slice-dedup | library | 修复 Dedup：输入为空 slice 或 nil 时 panic，应返回空 slice |
| tier1/fix-map-merge | library | 修复 MergeMaps：浅拷贝导致修改返回值影响原 map，应改为深拷贝 |
| tier1/add-string-truncate | library | 添加 Truncate 函数，按单词边界截断并加 "..." |
| tier1/add-min-max | library | 添加泛型 Min/Max 函数，支持所有 cmp.Ordered 类型 |
| tier1/fix-flag-default | cli | 修复 --port flag 默认值：当前为 0，应为 8080 |
| tier1/fix-output-format | cli | 修复 formatTasks：制表符拼接不对齐，应使用 tabwriter 对齐 |
| tier1/fix-exit-code | cli | 修复退出码：命令失败时应返回 1，当前总返回 0 |
| tier1/add-version-cmd | cli | 添加 version 子命令，输出 "taskcli v0.1.0" |
| tier1/fix-race-condition | concurrency | 修复 Counter 并发安全：Inc/Get 无锁保护，导致数据竞态 |
| tier1/fix-goroutine-leak | concurrency | 修复 worker pool Stop 方法：不通知 worker 退出，导致 goroutine 泄漏 |

---

## Tier 2 — 中等任务（32 个，各 10 分钟）

### http-server（10 个）

| 任务 ID | 描述 |
|---------|------|
| tier2/extract-middleware-logging | 从每个 handler 提取重复的 log.Printf，创建统一 logging 中间件 |
| tier2/add-request-validation | 为 POST /todos 添加请求验证，拒绝无效输入 |
| tier2/add-pagination-headers | 为 GET /todos 添加分页响应头（X-Total-Count 等） |
| tier2/extract-error-handler | 提取 errorResponse 函数，统一所有错误响应为 JSON 格式 |
| tier2/add-request-id | 添加中间件为每个请求生成 UUID 放入 X-Request-ID 头 |
| tier2/add-timeout-middleware | 添加 context timeout 中间件，防止请求超时 |
| tier2/fix-concurrent-map | 修复 handler 中 map[string]int 的并发不安全访问 |
| tier2/add-graceful-shutdown | 添加信号处理和优雅关闭 |
| tier2/add-list-filtering | 为 GET /todos 添加 ?done=true/false 过滤 |
| tier2/add-bulk-create | 添加 POST /todos/bulk 批量创建端点 |

### library（8 个）

| 任务 ID | 描述 |
|---------|------|
| tier2/extract-cache-interface | 从 MapCache 提取 Cache[K,V] 泛型接口 |
| tier2/add-lru-eviction | 为 cache 添加 LRU 淘汰策略 |
| tier2/add-retry-func | 实现通用重试函数，带指数退避 |
| tier2/add-result-type | 实现 Result[T] 泛型类型，提供 Map/FlatMap 等函数式操作 |
| tier2/extract-logger-interface | 从硬编码 fmt.Printf 提取 Logger 接口并注入 |
| tier2/add-ring-buffer | 实现固定大小的 RingBuffer[T] |
| tier2/add-semaphore | 实现基于 channel 的信号量 |
| tier2/fix-deep-copy | 修复嵌套 struct 的 Clone() 为深拷贝 |

### cli（5 个）

| 任务 ID | 描述 |
|---------|------|
| tier2/add-json-output | 为 list 命令添加 --format json 输出 |
| tier2/add-config-file | 添加 JSON 配置文件加载支持 |
| tier2/add-table-output | 将 list 输出改为 tabwriter 表格格式 |
| tier2/add-filter-flag | 为 list 命令添加 --status 过滤参数 |
| tier2/add-progress-bar | 为长操作添加文字进度条 |

### concurrency（5 个）

| 任务 ID | 描述 |
|---------|------|
| tier2/add-worker-pool | 实现固定大小的 worker pool |
| tier2/fix-channel-deadlock | 修复 unbuffered channel 导致的死锁 |
| tier2/add-fan-out | 实现 FanOut 并发处理函数 |
| tier2/add-rate-limiter | 实现令牌桶 RateLimiter |
| tier2/fix-waitgroup-leak | 修复 goroutine panic 导致 WaitGroup 永不 Done |

### reconciler（4 个）

| 任务 ID | 描述 |
|---------|------|
| tier2/fix-infinite-loop | 修复 Reconcile Requeue:true 时无退避导致的无限快速循环 |
| tier2/fix-status-conflict | 添加版本号和冲突重试机制 |
| tier2/add-finalizer | 资源删除时执行清理逻辑（finalizer 模式） |
| tier2/add-requeue-backoff | 为 reconciler 失败重试添加指数退避 |

---

## Tier 3 — 复杂任务（29 个，15-20 分钟）

### http-server（8 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier3/refactor-to-service | 18m | 拆分巨型 handler 为三层架构：Handler → Service → Repository |
| tier3/add-cache-layer | 18m | 为 GET 请求添加内存缓存层（sync.Map + TTL），写操作时清除缓存 |
| tier3/add-auth-middleware | 15m | 添加 Bearer token 认证中间件，保护 CRUD 端点，Health 不认证 |
| tier3/add-rate-limit | 18m | 添加 per-IP 令牌桶限速中间件，超限返回 429 |
| tier3/add-sse-notifications | 20m | 添加 SSE 端点 GET /events，todo CRUD 变更时推送事件 |
| tier3/add-batch-operations | 18m | POST /todos/batch 批量 create/update/delete，全部成功或全部回滚 |
| tier3/refactor-config | 15m | 提取硬编码配置为 Config struct，支持环境变量覆盖和默认值 |
| tier3/refactor-error-types | 15m | 统一分散的 http.Error 为 AppError 类型（错误码 + 消息 + HTTP 状态码） |

### library（6 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier3/implement-lru-cache | 18m | 完整 LRU Cache：Get/Set/Delete/Len，容量限制 + TTL + 并发安全 |
| tier3/implement-circuit-breaker | 18m | 熔断器模式：Closed → Open → HalfOpen 状态机 |
| tier3/implement-observer | 15m | 事件总线：Subscribe/Publish/Unsubscribe，按 topic 隔离 |
| tier3/refactor-to-generics | 15m | 将 interface{} 集合工具重构为泛型版本 |
| tier3/implement-pipeline | 18m | 类型安全 Pipeline[I,O]：Stage 链式调用 + 自动错误传播 |
| tier3/implement-pool | 15m | 泛型对象池 Pool[T]：Get/Put，并发安全，可配置最大大小 |

### concurrency（5 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier3/implement-concurrent-pipeline | 18m | producer → transform → consumer 三阶段 channel 管道 |
| tier3/implement-pubsub | 18m | 内存 PubSub：Topic 隔离、异步投递、关闭清理 |
| tier3/implement-batch-processor | 18m | 批处理器：攒够 N 条或超时 T 后批量处理 |
| tier3/fix-context-cancel | 15m | 修复外层 context 取消不传播到内部 goroutine |
| tier3/add-graceful-shutdown | 18m | ServiceGroup 管理多 goroutine 服务，signal 通知 + context 级联取消 |

### reconciler（6 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier3/refactor-reconciler | 18m | 拆分巨型 Reconcile() 为 validate/ensure/updateStatus 三个子 reconciler |
| tier3/implement-conditions | 18m | 从 Phase 字符串迁移到 Conditions 数组（type/status/reason/message） |
| tier3/implement-dependent-resources | 20m | 依赖资源自动创建：main → config → secret 链式依赖 |
| tier3/add-event-recording | 15m | 添加 EventRecorder 接口记录 reconcile 事件 |
| tier3/add-metrics | 18m | MetricsCollector 接口 + InstrumentedReconciler 收集次数/延迟/错误率 |
| tier3/add-owner-reference | 18m | OwnerReference 管理，父资源删除时级联删除子资源 |

### cli（4 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier3/refactor-commands | 15m | switch-case 命令分发重构为 Command 接口 + Registry 模式 |
| tier3/add-interactive-mode | 18m | 添加 REPL 交互模式，用户可连续执行命令 |
| tier3/add-plugin-system | 18m | 从 plugins 目录加载 JSON 配置定义自定义命令 |
| tier3/add-config-validation | 15m | 配置 Validate() 验证、默认值合并、多错误聚合 |

---

## Tier 4 — 高级任务（19 个，25-30 分钟）

### http-server（6 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier4/implement-rbac | 30m | 完整 RBAC 系统：角色定义、权限检查中间件、用户角色映射、角色继承 |
| tier4/add-api-versioning | 25m | /v1/ 和 /v2/ 版本路由，v2 支持新字段，v1 向后兼容 |
| tier4/implement-event-sourcing | 30m | 重构为事件溯源：所有变更记录为事件，状态通过回放重建 |
| tier4/add-distributed-lock | 25m | 内存模拟分布式锁：Locker 接口、TTL 自动释放、owner 验证 |
| tier4/implement-saga | 30m | Saga 模式：步骤定义、顺序执行、失败补偿回滚 |
| tier4/full-refactor | 30m | 500+ 行 God Handler 完整重构为分层架构 |

### library（4 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier4/implement-btree | 30m | 完整泛型 B-tree：Insert/Search/Delete/Range，自平衡 |
| tier4/implement-consistent-hash | 25m | 一致性哈希环：虚拟节点、节点增删、最小迁移量 |
| tier4/implement-raft-log | 30m | 简化版 Raft 日志复制：追加、提交、应用 |
| tier4/implement-expression-parser | 25m | 数学表达式解析器：tokenize → parse → evaluate，支持四则运算和括号 |

### concurrency（3 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier4/implement-map-reduce | 25m | 泛型并行 MapReduce：Map 并行处理 → Shuffle 按 key 分组 → Reduce 并行聚合 |
| tier4/implement-actor-model | 30m | Actor 模型：独立并发单元 + Mailbox 接收 + 顺序处理 + ActorRef 发送 |
| tier4/implement-scheduler | 30m | 任务调度器：解析简化 cron 表达式，定时执行，支持取消 |

### reconciler（4 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier4/implement-full-operator | 30m | 完整 K8s-style Operator：资源定义、Reconciler 状态机、Status 管理、事件记录 |
| tier4/implement-garbage-collector | 30m | GC 控制器：OwnerReference 建立 parent-child 关系，自动级联删除孤儿资源 |
| tier4/implement-multi-resource | 30m | 多资源协调器：A → B → C 依赖链自动创建 |
| tier4/add-leader-election | 30m | 基于内存 Lease 的 Leader Election：竞争、续期、过期竞选 |

### cli（2 个）

| 任务 ID | 时间 | 描述 |
|---------|------|------|
| tier4/implement-workflow-engine | 30m | 工作流引擎：解析 YAML 工作流定义，支持顺序和并行执行 |
| tier4/add-full-tui | 30m | 完整 TUI：列表选择器、文本输入框、状态栏，纯 ANSI escape 序列实现 |
