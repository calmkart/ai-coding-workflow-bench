# Task: 重构为 Command 接口模式

## 目标
将 main 中的 switch-case 命令分发重构为 Command 接口 + Registry 模式。

## 变更范围
- main.go: 使用 Registry 替代 switch-case
- commands.go: 每个命令实现 Command 接口
- 新增 registry.go: CommandRegistry

## 具体要求
- REQ-1: Command 接口 {Name() string, Description() string, Run(args []string) error}
- REQ-2: Registry struct {Register(Command), Get(name) (Command, bool), List() []Command}
- REQ-3: AddCommand, ListCommand, DoneCommand 各实现 Command 接口
- REQ-4: main 通过 Registry 分发命令
- REQ-5: 自动生成 help 命令列出所有已注册命令
- REQ-6: 未知命令返回错误（exit code 1）
- REQ-7: 命令执行错误返回 error 而非 os.Exit

## 约束
- 纯 stdlib
- 保持现有命令行为不变

## 测试策略
- 验证所有命令正常工作
- 验证 help 列出所有命令
- 验证未知命令返回错误

## 不做什么
- 不添加新命令
- 不改变输出格式
