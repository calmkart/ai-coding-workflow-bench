# Task: 添加配置验证和默认值合并

## 目标
为 CLI 配置添加 Validate() 验证、默认值合并和多错误聚合。

## 变更范围
- config.go: 当前无验证的配置加载，需要添加验证

## 具体要求
- REQ-1: Config struct {DataDir, MaxTasks, DefaultPriority, OutputFormat string}
- REQ-2: LoadConfig(path) 从 JSON 文件加载配置
- REQ-3: DefaultConfig() 返回默认配置
- REQ-4: MergeConfig(base, override) 合并配置（override 非零值覆盖 base）
- REQ-5: Validate() error 验证所有字段，返回 ValidationErrors（聚合多个错误）
- REQ-6: ValidationErrors 实现 error 接口，包含所有错误详情
- REQ-7: 验证规则：DataDir 非空，MaxTasks > 0，OutputFormat 必须是 text/json/table
- REQ-8: LoadAndValidate(path) 组合加载+默认值+验证

## 约束
- 纯 stdlib
- JSON 格式配置

## 测试策略
- 验证默认配置值
- 验证验证规则
- 验证错误聚合
- 验证合并逻辑

## 不做什么
- 不实现 YAML
- 不实现配置热更新
