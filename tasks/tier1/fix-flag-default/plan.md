# Task: 修复 --port flag 默认值

## 目标
修复 taskcli 的 --port flag 默认值：当前默认值为 0，应该是 8080。

## 变更范围
- main.go: 修复 flag 定义的默认值

## 具体要求
- REQ-1: 不传 --port 时使用默认值 8080
- REQ-2: 传 --port 时使用指定值
- REQ-3: run 函数返回的 Config 包含正确的端口值

## 约束
- run() 函数签名不可更改
- Config struct 不可更改

## 测试策略
- 验证不传参数时 port = 8080
- 验证传 --port=3000 时 port = 3000
- 验证传 --port 9090 时 port = 9090

## 不做什么
- 不添加新的 flag
- 不修改 Config struct
