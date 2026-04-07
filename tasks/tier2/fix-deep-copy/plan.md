# Task: 修复深拷贝

## 目标
修复嵌套 struct 的 Clone() 方法为深拷贝。

## 变更范围
- config.go: 修复 Config.Clone() 方法

## 具体要求
- REQ-1: Clone() 必须深拷贝所有嵌套 struct
- REQ-2: Clone() 必须深拷贝 slice（不共享底层数组）
- REQ-3: Clone() 必须深拷贝 map（不共享底层数据）
- REQ-4: 修改克隆后的对象不影响原始对象

## 约束
- 不使用 encoding/json 或反射
- 手动实现深拷贝

## 测试策略
- 验证修改克隆不影响原始
- 验证嵌套 slice 独立
- 验证嵌套 map 独立

## 不做什么
- 不使用第三方深拷贝库
- 不添加新方法
