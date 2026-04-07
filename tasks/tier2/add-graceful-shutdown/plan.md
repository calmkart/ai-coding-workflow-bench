# Task: 添加优雅关闭

## 目标
为 HTTP 服务器添加信号处理和优雅关闭功能。

## 变更范围
- main.go: 重写 main() 使用 http.Server 和 signal 处理

## 具体要求
- REQ-1: 使用 http.Server 而非直接 http.ListenAndServe
- REQ-2: 监听 SIGTERM 和 SIGINT 信号
- REQ-3: 收到信号后调用 srv.Shutdown(ctx) 优雅关闭
- REQ-4: 关闭超时 30 秒
- REQ-5: 提供 runServer(handler http.Handler, addr string) 函数方便测试

## 约束
- setupRouter() 函数签名不可更改

## 测试策略
- 验证服务器启动正常
- 验证关闭信号被处理
- 验证 setupRouter 仍然可用

## 不做什么
- 不添加健康检查端点
- 不添加 readiness probe
