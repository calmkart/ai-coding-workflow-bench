# Task: 添加 JSON 配置文件插件系统

## 目标
从 plugins 目录加载 JSON 配置文件，每个文件定义一个自定义命令。

## 变更范围
- main.go: 加载插件目录
- 新增 plugin.go: 插件加载和执行逻辑

## 具体要求
- REQ-1: JSON 插件格式 {"name":"cmd","description":"desc","template":"echo {{.Args}}"}
- REQ-2: LoadPlugins(dir) ([]Plugin, error) 从目录加载所有 .json 文件
- REQ-3: 插件命令通过 text/template 渲染模板
- REQ-4: 内置命令优先级高于插件命令
- REQ-5: taskcli --plugins-dir <dir> 指定插件目录
- REQ-6: 插件无效（格式错误、缺少字段）跳过并记录警告
- REQ-7: taskcli help 列出所有命令（包括插件）

## 约束
- 纯 stdlib
- 不使用 Go 插件（.so）

## 测试策略
- 验证从目录加载插件
- 验证插件命令执行
- 验证无效插件跳过
- 验证内置命令不被覆盖

## 不做什么
- 不实现 Go 插件（.so）
- 不实现远程插件
