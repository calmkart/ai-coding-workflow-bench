# Task: 添加带 TTL 的内存缓存层

## 目标
为 GET 请求添加内存缓存层，使用 sync.Map 和 time.AfterFunc 实现 TTL 过期。写操作（POST/PUT/DELETE）时清除相关缓存。

## 变更范围
- handlers.go: 当前无缓存的实现
- 新增 cache.go: TTLCache 实现

## 具体要求
- REQ-1: 实现 TTLCache struct，支持 Get/Set/Delete/Clear
- REQ-2: 每个缓存项有 TTL，过期自动删除（time.AfterFunc）
- REQ-3: GET /todos/{id} 结果缓存，命中直接返回
- REQ-4: GET /todos (列表) 结果按参数组合缓存
- REQ-5: POST/PUT/DELETE 操作后清除相关缓存
- REQ-6: TTL 默认 30 秒，可配置
- REQ-7: 缓存并发安全

## 约束
- setupRouter() 函数签名不变
- 纯 stdlib
- 缓存对 API 行为透明

## 测试策略
- 验证 GET 返回正确数据
- 验证写操作后缓存失效
- 验证 TTL 过期后重新获取

## 不做什么
- 不实现 LRU 淘汰
- 不实现分布式缓存
- 不限制缓存大小
