# Task Catalog

workflow-bench includes 100 built-in Go coding tasks across 4 difficulty tiers and 5 code types. This document summarizes each task's type, estimated time, and description.

## Overview

| Tier | Count | Est. Time | Description |
|------|-------|-----------|-------------|
| Tier 1 | 20 | 5 min | Simple: single-point fixes or small feature additions |
| Tier 2 | 32 | 10 min | Medium: new endpoints, middleware, interface extraction |
| Tier 3 | 29 | 15-20 min | Complex: full pattern implementation, multi-file refactoring |
| Tier 4 | 19 | 25-30 min | Advanced: architecture-level refactoring, complete system implementation |

**Type distribution**: http-server (32) · library (24) · concurrency (15) · reconciler (14) · cli (15)

**Total estimated time**: ~1,460 minutes (24+ hours)

---

## Tier 1 — Simple Tasks (20 tasks, 5 min each)

| Task ID | Type | Description |
|---------|------|-------------|
| tier1/fix-handler-bug | http-server | Fix GET /todos pagination: requesting page=2 returns page 3 data (off-by-one) |
| tier1/fix-status-code | http-server | Fix POST /todos status code: should return 201 Created, currently returns 200 |
| tier1/fix-json-content-type | http-server | Add Content-Type: application/json header to all JSON responses |
| tier1/fix-delete-404 | http-server | Fix DELETE /todos/{id}: should return 404 for non-existent todo, currently returns 500 |
| tier1/fix-empty-body | http-server | Fix POST /todos: empty body causes panic (nil pointer), should return 400 |
| tier1/fix-query-param | http-server | Fix GET /todos: invalid page_size causes panic, should return 400 |
| tier1/add-health-check | http-server | Add GET /health endpoint |
| tier1/add-cors-headers | http-server | Add CORS support for cross-origin frontend calls |
| tier1/fix-string-reverse | library | Fix Reverse: byte-level reversal corrupts multi-byte UTF-8 characters, should reverse by rune |
| tier1/fix-contains-bug | library | Fix ContainsAny: empty candidates slice should return false, currently returns true |
| tier1/fix-slice-dedup | library | Fix Dedup: panics on empty/nil slice, should return empty slice |
| tier1/fix-map-merge | library | Fix MergeMaps: shallow copy causes mutations to affect original map, should deep copy |
| tier1/add-string-truncate | library | Add Truncate function: truncate at word boundary with "..." suffix |
| tier1/add-min-max | library | Add generic Min/Max functions for all cmp.Ordered types |
| tier1/fix-flag-default | cli | Fix --port flag default: currently 0, should be 8080 |
| tier1/fix-output-format | cli | Fix formatTasks: tab concatenation misaligns, should use tabwriter |
| tier1/fix-exit-code | cli | Fix exit code: should return 1 on failure, currently always returns 0 |
| tier1/add-version-cmd | cli | Add version subcommand outputting "taskcli v0.1.0" |
| tier1/fix-race-condition | concurrency | Fix Counter concurrency: Inc/Get lack lock protection, causing data race |
| tier1/fix-goroutine-leak | concurrency | Fix worker pool Stop: doesn't signal workers to exit, causing goroutine leak |

---

## Tier 2 — Medium Tasks (32 tasks, 10 min each)

### http-server (10)

| Task ID | Description |
|---------|-------------|
| tier2/extract-middleware-logging | Extract repeated log.Printf from handlers into unified logging middleware |
| tier2/add-request-validation | Add request validation for POST /todos, reject invalid input |
| tier2/add-pagination-headers | Add pagination response headers (X-Total-Count, etc.) to GET /todos |
| tier2/extract-error-handler | Extract errorResponse function, unify all error responses as JSON |
| tier2/add-request-id | Add middleware to generate UUID in X-Request-ID header per request |
| tier2/add-timeout-middleware | Add context timeout middleware to prevent long-running requests |
| tier2/fix-concurrent-map | Fix concurrent-unsafe map[string]int access in handlers |
| tier2/add-graceful-shutdown | Add signal handling and graceful shutdown |
| tier2/add-list-filtering | Add ?done=true/false query parameter filtering to GET /todos |
| tier2/add-bulk-create | Add POST /todos/bulk batch creation endpoint |

### library (8)

| Task ID | Description |
|---------|-------------|
| tier2/extract-cache-interface | Extract Cache[K,V] generic interface from MapCache |
| tier2/add-lru-eviction | Add LRU eviction policy to cache |
| tier2/add-retry-func | Implement generic retry function with exponential backoff |
| tier2/add-result-type | Implement Result[T] generic type with Map/FlatMap functional operations |
| tier2/extract-logger-interface | Extract Logger interface from hardcoded fmt.Printf and inject |
| tier2/add-ring-buffer | Implement fixed-size RingBuffer[T] |
| tier2/add-semaphore | Implement channel-based semaphore |
| tier2/fix-deep-copy | Fix nested struct Clone() to perform deep copy |

### cli (5)

| Task ID | Description |
|---------|-------------|
| tier2/add-json-output | Add --format json output to list command |
| tier2/add-config-file | Add JSON config file loading support |
| tier2/add-table-output | Change list output to tabwriter table format |
| tier2/add-filter-flag | Add --status filter parameter to list command |
| tier2/add-progress-bar | Add text progress bar for long operations |

### concurrency (5)

| Task ID | Description |
|---------|-------------|
| tier2/add-worker-pool | Implement fixed-size worker pool |
| tier2/fix-channel-deadlock | Fix deadlock caused by unbuffered channel |
| tier2/add-fan-out | Implement FanOut concurrent processing function |
| tier2/add-rate-limiter | Implement token bucket RateLimiter |
| tier2/fix-waitgroup-leak | Fix goroutine panic causing WaitGroup to never call Done |

### reconciler (4)

| Task ID | Description |
|---------|-------------|
| tier2/fix-infinite-loop | Fix Reconcile Requeue:true without backoff causing infinite fast loop |
| tier2/fix-status-conflict | Add version number and conflict retry mechanism |
| tier2/add-finalizer | Execute cleanup logic on resource deletion (finalizer pattern) |
| tier2/add-requeue-backoff | Add exponential backoff for reconciler failure retries |

---

## Tier 3 — Complex Tasks (29 tasks, 15-20 min)

### http-server (8)

| Task ID | Time | Description |
|---------|------|-------------|
| tier3/refactor-to-service | 18m | Split monolithic handler into 3-layer: Handler → Service → Repository |
| tier3/add-cache-layer | 18m | Add in-memory cache for GET requests (sync.Map + TTL), invalidate on writes |
| tier3/add-auth-middleware | 15m | Add Bearer token auth middleware protecting CRUD endpoints, Health excluded |
| tier3/add-rate-limit | 18m | Add per-IP token bucket rate limiting middleware, return 429 when exceeded |
| tier3/add-sse-notifications | 20m | Add SSE endpoint GET /events, push events on todo CRUD changes |
| tier3/add-batch-operations | 18m | POST /todos/batch for bulk create/update/delete, all-or-nothing rollback |
| tier3/refactor-config | 15m | Extract hardcoded config to Config struct with env var overrides and defaults |
| tier3/refactor-error-types | 15m | Unify scattered http.Error into AppError type (code + message + HTTP status) |

### library (6)

| Task ID | Time | Description |
|---------|------|-------------|
| tier3/implement-lru-cache | 18m | Full LRU Cache: Get/Set/Delete/Len, capacity limit + TTL + concurrency-safe |
| tier3/implement-circuit-breaker | 18m | Circuit Breaker pattern: Closed → Open → HalfOpen state machine |
| tier3/implement-observer | 15m | Event bus: Subscribe/Publish/Unsubscribe, topic-isolated |
| tier3/refactor-to-generics | 15m | Refactor interface{} collection utils to generic versions |
| tier3/implement-pipeline | 18m | Type-safe Pipeline[I,O]: chained Stages + auto error propagation |
| tier3/implement-pool | 15m | Generic object Pool[T]: Get/Put, concurrency-safe, configurable max size |

### concurrency (5)

| Task ID | Time | Description |
|---------|------|-------------|
| tier3/implement-concurrent-pipeline | 18m | producer → transform → consumer 3-stage channel pipeline |
| tier3/implement-pubsub | 18m | In-memory PubSub: topic isolation, async delivery, shutdown cleanup |
| tier3/implement-batch-processor | 18m | Batch processor: flush after N items or timeout T |
| tier3/fix-context-cancel | 15m | Fix outer context cancellation not propagating to inner goroutines |
| tier3/add-graceful-shutdown | 18m | ServiceGroup managing multiple goroutine services with signal + context cascade |

### reconciler (6)

| Task ID | Time | Description |
|---------|------|-------------|
| tier3/refactor-reconciler | 18m | Split monolithic Reconcile() into validate/ensure/updateStatus sub-reconcilers |
| tier3/implement-conditions | 18m | Migrate from Phase string to Conditions array (type/status/reason/message) |
| tier3/implement-dependent-resources | 20m | Auto-create dependent resources: main → config → secret chain |
| tier3/add-event-recording | 15m | Add EventRecorder interface for reconcile event recording |
| tier3/add-metrics | 18m | MetricsCollector + InstrumentedReconciler for count/latency/error rate |
| tier3/add-owner-reference | 18m | OwnerReference management, cascade delete children on parent deletion |

### cli (4)

| Task ID | Time | Description |
|---------|------|-------------|
| tier3/refactor-commands | 15m | Refactor switch-case dispatch to Command interface + Registry pattern |
| tier3/add-interactive-mode | 18m | Add REPL interactive mode for continuous command execution |
| tier3/add-plugin-system | 18m | Load JSON config from plugins directory to define custom commands |
| tier3/add-config-validation | 15m | Config Validate(), default merging, multi-error aggregation |

---

## Tier 4 — Advanced Tasks (19 tasks, 25-30 min)

### http-server (6)

| Task ID | Time | Description |
|---------|------|-------------|
| tier4/implement-rbac | 30m | Full RBAC: role definitions, permission-check middleware, user-role mapping, role inheritance |
| tier4/add-api-versioning | 25m | /v1/ and /v2/ versioned routes, v2 adds new fields, v1 backward compatible |
| tier4/implement-event-sourcing | 30m | Refactor to event sourcing: all changes as events, state rebuilt by replay |
| tier4/add-distributed-lock | 25m | In-memory distributed lock: Locker interface, TTL auto-release, owner verification |
| tier4/implement-saga | 30m | Saga pattern: step definitions, sequential execution, failure compensation rollback |
| tier4/full-refactor | 30m | Refactor 500+ line God Handler into layered architecture |

### library (4)

| Task ID | Time | Description |
|---------|------|-------------|
| tier4/implement-btree | 30m | Full generic B-tree: Insert/Search/Delete/Range, self-balancing |
| tier4/implement-consistent-hash | 25m | Consistent hash ring: virtual nodes, add/remove nodes, minimal migration |
| tier4/implement-raft-log | 30m | Simplified Raft log replication: append, commit, apply |
| tier4/implement-expression-parser | 25m | Math expression parser: tokenize → parse → evaluate, supports +,-,*,/ and parentheses |

### concurrency (3)

| Task ID | Time | Description |
|---------|------|-------------|
| tier4/implement-map-reduce | 25m | Generic parallel MapReduce: Map parallel → Shuffle by key → Reduce parallel |
| tier4/implement-actor-model | 30m | Actor model: independent concurrent unit + Mailbox + sequential processing + ActorRef |
| tier4/implement-scheduler | 30m | Task scheduler: parse simplified cron expressions, timed execution, cancellable |

### reconciler (4)

| Task ID | Time | Description |
|---------|------|-------------|
| tier4/implement-full-operator | 30m | Full K8s-style Operator: resource definition, Reconciler state machine, Status, events |
| tier4/implement-garbage-collector | 30m | GC controller: OwnerReference parent-child, auto cascade delete orphans |
| tier4/implement-multi-resource | 30m | Multi-resource reconciler: A → B → C dependency chain auto-creation |
| tier4/add-leader-election | 30m | Memory-based Lease Leader Election: compete, renew, expire and re-elect |

### cli (2)

| Task ID | Time | Description |
|---------|------|-------------|
| tier4/implement-workflow-engine | 30m | Workflow engine: parse YAML workflow definitions, sequential and parallel execution |
| tier4/add-full-tui | 30m | Full TUI: list selector, text input, status bar, pure ANSI escape sequences |
