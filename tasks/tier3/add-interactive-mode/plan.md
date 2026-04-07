# Task: 添加交互式 REPL 模式

## 目标
添加 REPL（Read-Eval-Print-Loop）交互模式，用户可以连续执行命令。

## 变更范围
- main.go: 添加 repl 子命令
- 新增 repl.go: REPL 实现

## 具体要求
- REQ-1: taskcli repl 进入交互模式
- REQ-2: 显示提示符 "> "
- REQ-3: 支持所有已有命令（add, list, done）
- REQ-4: 支持 help 显示可用命令
- REQ-5: 支持 exit/quit 退出
- REQ-6: 空输入忽略
- REQ-7: 未知命令显示错误提示
- REQ-8: RunREPL(reader, writer) 函数可注入 io.Reader/Writer 用于测试

## 约束
- 纯 stdlib
- bufio.Scanner 读取输入

## 测试策略
- 验证命令执行（通过注入 reader/writer）
- 验证 exit 退出
- 验证 help 输出

## 不做什么
- 不实现命令历史
- 不实现 Tab 补全
