# 开发指南

## 前置条件

- Go 1.22+
- git
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-cli)（运行基准测试时需要，开发本身不需要）

可选工具（提升 L3 静态分析效果）：
- `staticcheck` -- `go install honnef.co/go/tools/cmd/staticcheck@latest`
- `gosec` -- `go install github.com/securego/gosec/v2/cmd/gosec@latest`

## 项目结构

```
cmd/workflow-bench/main.go       CLI 入口，cobra 命令定义
internal/
  config/config.go               配置加载、任务发现、任务过滤
  engine/
    runner.go                    主执行流水线（发现 -> 运行 -> 验证 -> 存储）
    isolation.go                 Git worktree 创建和清理
    verify.go                    从嵌入模板生成验证脚本
    collector.go                 解析 BENCH_RESULT 输出为 L1-L4 分数
    templates/
      http_server.sh.tmpl        http-server 任务的验证脚本模板
  adapter/
    adapter.go                   Adapter 接口、RunOutput 类型、注册表
    vanilla.go                   VanillaAdapter：使用计划运行 `claude -p`
    v4claude.go                  V4ClaudeAdapter：运行 `claude --agent manager`
    custom.go                    CustomAdapter：用户自定义命令执行
  metrics/
    correctness.go               正确性评分计算（加权 L1-L4）
  store/
    db.go                        SQLite 数据库操作（纯 Go，无 CGO）
    schema.sql                   数据库 schema（通过 go:embed 嵌入）
  report/
    summary.go                   Markdown 报告生成
    templates/
      summary.md.tmpl            报告模板（通过 go:embed 嵌入）
tasks/
  tier1/
    fix-handler-bug/             T1 任务：修复分页 off-by-one
    add-health-check/            T1 任务：添加 /health 端点
```

## 构建

```bash
# 构建二进制
make build
# 或
go build -o workflow-bench ./cmd/workflow-bench

# 运行测试
make test
# 或
go test ./... -count=1 -race

# 清理
make clean
```

## 运行测试

```bash
# 全部测试
go test ./... -count=1

# 启用竞态检测器
go test ./... -count=1 -race

# 指定包
go test ./internal/metrics/... -v

# 生成覆盖率报告
go test ./... -count=1 -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试组织

- `*_test.go` -- 与源文件同目录的单元测试
- `scenario_test.go` -- 各包内的场景/集成测试
- `integration_test.go` -- 跨包集成测试（在 engine/ 中）

store 包使用 `:memory:` SQLite 进行测试，无需文件清理。

## 关键依赖

| 依赖 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 框架 |
| `gopkg.in/yaml.v3` | 解析 bench.yaml 和 task.yaml |
| `modernc.org/sqlite` | 纯 Go SQLite 驱动（无需 CGO） |

## 代码模式

### 嵌入模板

模板在编译时通过 `//go:embed` 嵌入：

```go
//go:embed templates/http_server.sh.tmpl
var httpServerTemplate string
```

修改模板后需要重新构建二进制。

### 错误处理

函数返回带上下文包装的错误：

```go
return nil, fmt.Errorf("load plan: %w", err)
```

基础设施错误向上传播。预期中的失败（如验证中的测试失败）作为包含结构化数据的非错误结果处理。

### 数据库操作

所有数据库操作通过 `internal/store/db.go` 执行。migration 通过 `schema.sql` 嵌入并在 `Open()` 时自动应用。数据库使用 WAL 模式以支持并发安全。

### 任务发现

任务通过 `filepath.Glob` 扫描 `tasks/tier*/*/task.yaml` 发现。不需要索引文件——添加包含有效 `task.yaml` 的新任务目录即可。

## 添加新 Adapter

**简单场景**：使用 `custom` adapter——仅需在 `bench.yaml` 中配置 YAML，无需编写 Go 代码。定义 `entry_command` 和可选的 `setup_commands` 即可运行任何外部工具。详见 [configuration.md](configuration.md)。

**复杂场景**（自定义 token 解析、特殊错误处理等）：按下述方式编写 Go adapter。

1. 创建 `internal/adapter/myadapter.go`：

```go
package adapter

import "context"

type MyAdapter struct{}

func NewMyAdapter(cfg map[string]any) (Adapter, error) {
    return &MyAdapter{}, nil
}

func (a *MyAdapter) Name() string { return "my-adapter" }

func (a *MyAdapter) Setup(ctx context.Context, worktreeDir string) error {
    // 准备 worktree（复制 agent 文件、配置等）
    return nil
}

func (a *MyAdapter) Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error) {
    // 执行 workflow 并返回结果
    start := time.Now()
    // ...
    return &RunOutput{
        ExitCode: 0,
        WallTime: time.Since(start),
    }, nil
}
```

2. 在 `adapter.go` 中注册：

```go
var Registry = map[string]func(cfg map[string]any) (Adapter, error){
    "vanilla":    NewVanilla,
    "my-adapter": NewMyAdapter,
}
```

3. 添加到 `bench.yaml`：

```yaml
workflows:
  my-workflow:
    adapter: my-adapter
```

## 添加新任务类型

目前仅支持 `http-server`。添加新类型的步骤：

1. 在 `internal/engine/templates/<type>.sh.tmpl` 创建验证模板
2. 在 `verify.go` 中嵌入：

```go
//go:embed templates/my_type.sh.tmpl
var myTypeTemplate string
```

3. 在 `GenerateVerifyDir` 中添加 case：

```go
switch cfg.TaskType {
case "http-server":
    tmplStr = httpServerTemplate
case "my-type":
    tmplStr = myTypeTemplate
}
```

4. 模板接收 worktree 路径作为 `$1`，必须输出如下格式：
```
BENCH_RESULT: L1=PASS L2=N/M L3=K L4=N/M
```

## 添加新任务

完整任务编写指南见 [tasks.md](tasks.md)。快速检查清单：

1. 创建目录：`tasks/tier{N}/{task-name}/`
2. 编写 `task.yaml`，填写所有必填字段
3. 编写 `plan.md`，包含清晰的指令
4. 准备 `repo/` 为可编译的 git 仓库
5. 编写 `verify/e2e_test.go.src`（真值测试）
6. 验证：`workflow-bench validate --tasks tier{N}/{task-name} -v`

## 调试

### 详细日志

```bash
workflow-bench run --tasks tier1 --tag debug -v
```

`-v` 选项通过 `slog` 启用 debug 级别的结构化日志。

### 查看结果

结果存储在 `~/.claude/workflow-bench/results.db` 中，可直接查询：

```bash
sqlite3 ~/.claude/workflow-bench/results.db "SELECT task_id, status, correctness_score FROM runs WHERE tag='my-tag'"
```

### 验证脚本

验证脚本生成到临时目录，每次运行后删除。如需查看，可添加调试输出或临时注释掉 `runner.go` 中的 `defer os.RemoveAll(verifyDir)`。
