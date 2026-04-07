# Task: 添加配置文件

## 目标
添加 JSON 配置文件加载支持。

## 变更范围
- commands.go: 添加配置加载逻辑
- main.go: 加载配置

## 具体要求
- REQ-1: 从 config.json 读取配置
- REQ-2: 配置项: data_file(数据文件路径), default_format(默认输出格式)
- REQ-3: 配置文件不存在时使用默认值
- REQ-4: 支持 --config 指定配置文件路径

## 约束
- 使用 encoding/json（不用 yaml）
- 不使用外部依赖

## 测试策略
- 验证读取配置文件
- 验证默认值
- 验证 --config 指定路径

## 不做什么
- 不添加配置写入
- 不添加环境变量支持
