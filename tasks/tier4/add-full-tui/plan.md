# Task: 实现完整 TUI 界面

## 目标
实现完整的 TUI（终端用户界面）：列表选择器、文本输入框、状态栏。使用纯 ANSI escape 序列，不依赖外部库。

## 当前状态
- 只有空的接口定义

## 变更范围
- tui.go: App 框架和 ANSI 渲染
- components.go: 组件实现

## 具体要求
- REQ-1: Component 接口：Render() string, HandleKey(KeyEvent) bool, Focused() bool, SetFocused(bool)
- REQ-2: ListSelector 组件：上下键选择，回车确认，支持多项显示
- REQ-3: TextInput 组件：文字输入，退格删除，光标显示
- REQ-4: StatusBar 组件：底部状态栏，显示消息文本
- REQ-5: App 管理组件列表，Tab 切换焦点
- REQ-6: Render() 返回完整的 ANSI 字符串（可用于测试验证）
- REQ-7: HandleKey 返回 true 表示已处理该按键
- REQ-8: KeyEvent 类型包含：Rune(普通字符), Up, Down, Enter, Tab, Backspace, Escape
- REQ-9: ANSI 序列：清屏、移动光标、颜色（前景/背景）、粗体、反色
- REQ-10: ListSelector 高亮当前选中项（反色）
- REQ-11: TextInput 显示提示符和光标位置
- REQ-12: 组件可组合为完整界面

## ANSI Escape 序列
```
清屏: \033[2J
移动光标: \033[{row};{col}H
前景色: \033[3{color}m
背景色: \033[4{color}m
粗体: \033[1m
反色: \033[7m
重置: \033[0m
```

## 约束
- 纯 stdlib，不用 termbox/tcell/bubbletea
- 不需要实际运行终端（通过 Render() 返回字符串测试）
- 不需要实际读取键盘输入（通过 HandleKey 注入事件）

## 测试策略
- 列表选择上下移动
- 列表选择确认
- 文本输入和删除
- 状态栏显示
- Tab 切换焦点
- 渲染输出包含 ANSI 序列
- 组件组合

## 不做什么
- 不实现实际终端 I/O（只是渲染引擎）
- 不实现窗口管理
- 不实现鼠标支持
