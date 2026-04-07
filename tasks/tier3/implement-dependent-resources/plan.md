# Task: 实现依赖资源自动创建

## 目标
添加依赖资源自动创建：main resource -> config resource -> secret resource 链式依赖。

## 变更范围
- reconciler.go: 添加依赖资源创建逻辑

## 具体要求
- REQ-1: DependentResource struct {Name, Kind string, DependsOn string}
- REQ-2: DependencyGraph 管理资源间依赖关系
- REQ-3: Reconciler 按依赖顺序创建资源（先创建被依赖的）
- REQ-4: 被依赖资源 Ready 后才创建依赖它的资源
- REQ-5: 任一依赖失败，上层资源标记为等待
- REQ-6: EnsureDependents(resource) (allReady bool, error) 确保所有依赖资源就绪
- REQ-7: 依赖链：App -> Config -> Secret

## 约束
- Reconciler 接口不变
- 纯 stdlib

## 测试策略
- 验证链式创建顺序
- 验证依赖未就绪时等待
- 验证全部就绪后主资源 Ready

## 不做什么
- 不实现循环依赖检测
- 不实现并行创建
