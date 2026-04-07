# Task: 修复 MergeMaps 浅拷贝 bug

## 目标
修复 MergeMaps 函数：当前实现是浅拷贝，修改返回的 map 会影响原始输入 map，应改为深拷贝。

## 变更范围
- strkit.go: 修复 MergeMaps 函数的拷贝逻辑

## 具体要求
- REQ-1: 返回新的 map，包含所有输入 map 的键值对
- REQ-2: 后面的 map 覆盖前面的同名键
- REQ-3: 修改返回的 map 不影响原 map
- REQ-4: 空输入返回空 map（非 nil）
- REQ-5: nil map 输入被跳过

## 约束
- MergeMaps 函数签名不可更改

## 测试策略
- 验证合并后修改不影响原 map
- 验证后面的值覆盖前面的
- 验证空输入返回空 map

## 不做什么
- 不添加嵌套 map 的深拷贝（只处理 map[string]string）
- 不添加新函数
