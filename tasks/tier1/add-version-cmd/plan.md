# Task: 添加 version 子命令

## 目标
为 taskcli 添加 version 子命令，输出 "taskcli v0.1.0"。

## 变更范围
- main.go: 在 runCmd 函数中添加 "version" 子命令处理

## 具体要求
- REQ-1: `taskcli version` 输出包含 "taskcli v0.1.0"
- REQ-2: 输出写到 stdout（使用传入的 io.Writer）
- REQ-3: version 命令返回 nil error

## 约束
- runCmd 函数签名不可更改
- 现有子命令逻辑不受影响

## 测试策略
- 验证 version 子命令输出包含版本号
- 验证其他子命令不受影响

## 不做什么
- 不添加 --version flag
- 不修改其他子命令
