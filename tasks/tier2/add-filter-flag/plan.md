# Task: 添加过滤参数

## 目标
为 list 命令添加 --status 过滤参数。

## 变更范围
- commands.go: 添加过滤逻辑
- main.go: 解析 --status flag

## 具体要求
- REQ-1: --status done 只显示已完成
- REQ-2: --status pending 只显示未完成
- REQ-3: --status all 或不传显示全部
- REQ-4: 无效 status 值输出错误

## 约束
- 不使用外部依赖

## 测试策略
- 验证各状态过滤
- 验证默认显示全部
- 验证无效值报错

## 不做什么
- 不添加其他过滤条件
