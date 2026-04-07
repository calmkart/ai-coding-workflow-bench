# Task: 添加 JSON 输出

## 目标
为 list 命令添加 --format json 输出选项。

## 变更范围
- commands.go: 添加 JSON 格式化输出
- main.go: 添加 --format flag

## 具体要求
- REQ-1: --format text (默认) 保持原有文本输出
- REQ-2: --format json 输出 JSON 数组
- REQ-3: 无效 format 值输出错误信息并退出

## 约束
- 不使用外部依赖
- 保持现有文本输出不变

## 测试策略
- 验证 --format json 输出有效 JSON
- 验证默认输出是文本格式
- 验证 JSON 包含正确字段

## 不做什么
- 不添加 YAML 输出
- 不改变其他命令
