# Task: 实现多 goroutine 优雅关闭

## 目标
实现 ServiceGroup 管理多个 goroutine 服务，支持 signal 通知和 context 级联取消。

## 变更范围
- service.go: 当前无优雅关闭的服务框架

## 具体要求
- REQ-1: Service 接口 {Start(ctx) error, Stop(ctx) error, Name() string}
- REQ-2: ServiceGroup struct 管理多个 Service
- REQ-3: Add(svc) 注册服务
- REQ-4: Run(ctx) error 启动所有服务，阻塞直到 ctx 取消或任一服务出错
- REQ-5: 关闭顺序：发 cancel -> 并行调用所有 Stop -> 等待完成或超时
- REQ-6: 支持配置关闭超时（默认 30 秒）
- REQ-7: 任一服务 Start 失败触发全部关闭
- REQ-8: 所有服务 Stop 完成后返回聚合的错误

## 约束
- 纯 stdlib
- 不依赖 errgroup

## 测试策略
- 验证正常启停
- 验证 context 取消级联
- 验证单个失败触发全部停止
- 验证关闭超时

## 不做什么
- 不实现服务发现
- 不实现重启策略
