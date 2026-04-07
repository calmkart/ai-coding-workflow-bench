# Task: 提取配置为 Config struct

## 目标
将硬编码在代码中的配置提取为 Config struct，支持环境变量覆盖和默认值。

## 变更范围
- main.go: 使用 Config 初始化
- handlers.go: 使用 Config 中的参数
- 新增 config.go: Config struct + 加载逻辑

## 具体要求
- REQ-1: Config struct {Port int, ReadTimeout, WriteTimeout time.Duration, MaxPageSize int, DefaultPageSize int, MaxTitleLength int}
- REQ-2: LoadConfig() Config 从环境变量读取，未设置用默认值
- REQ-3: 环境变量命名：TODO_PORT, TODO_READ_TIMEOUT, TODO_WRITE_TIMEOUT, TODO_MAX_PAGE_SIZE, TODO_DEFAULT_PAGE_SIZE, TODO_MAX_TITLE_LENGTH
- REQ-4: 默认值：Port=8080, ReadTimeout=5s, WriteTimeout=10s, MaxPageSize=100, DefaultPageSize=10, MaxTitleLength=200
- REQ-5: Config.Validate() error 验证参数合法性
- REQ-6: handler 使用 Config 中的值而非硬编码

## 约束
- setupRouter() 函数签名不变（但内部可以使用默认 config）
- 纯 stdlib
- 环境变量是可选的

## 测试策略
- 验证默认配置值正确
- 验证环境变量覆盖
- 验证无效配置报错
- 验证 API 行为不变

## 不做什么
- 不实现配置文件解析
- 不实现配置热更新
