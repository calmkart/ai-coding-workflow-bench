# Task: 修复并发 map 访问

## 目标
修复 handler 中 map[string]int 的并发不安全访问。

## 变更范围
- handlers.go: 用 sync.Mutex 保护统计 map

## 具体要求
- REQ-1: 使用 sync.Mutex 保护 stats map 的读写
- REQ-2: GET /stats 端点返回统计数据（并发安全）
- REQ-3: 每个请求处理后递增对应路径的计数
- REQ-4: -race 检测器不报错

## 约束
- setupRouter() 函数签名不可更改
- 不改变统计逻辑

## 测试策略
- 并发请求不 panic
- 统计数据正确
- race detector 通过

## 不做什么
- 不使用 sync.Map（用 sync.Mutex 更清晰）
- 不改变端点功能
