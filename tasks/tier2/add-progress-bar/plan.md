# Task: 添加进度条

## 目标
为长操作添加简单的文字进度条。

## 变更范围
- commands.go: 实现 ProgressBar 类型和 process 命令
- main.go: 注册 process 命令

## 具体要求
- REQ-1: ProgressBar 类型: New(total int, writer io.Writer) *ProgressBar
- REQ-2: Update(current int) 更新进度
- REQ-3: 显示格式: [=====>    ] 50% (5/10)
- REQ-4: process 命令模拟处理并显示进度

## 约束
- 不使用外部依赖
- 纯文字输出（不用 ANSI 转义）

## 测试策略
- 验证进度条格式
- 验证百分比计算
- 验证 process 命令输出

## 不做什么
- 不添加 spinner
- 不使用 ANSI 颜色
