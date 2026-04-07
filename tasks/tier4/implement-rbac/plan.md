# Task: 实现基于角色的访问控制 (RBAC)

## 目标
当前 API 没有任何权限控制，所有用户可执行所有操作。需要实现完整的 RBAC 系统：角色定义、权限检查中间件、用户角色映射、角色继承。

## 当前问题
- 所有端点对所有人开放，无权限验证
- 认证 token 存在但未与角色关联
- 有一个硬编码的 admin 判断（直接比较用户名）

## 变更范围
- rbac.go: 实现 RBACManager（角色定义、权限映射、用户角色分配、继承）
- middleware.go: 实现 RBACMiddleware 从请求中提取用户并检查权限
- handlers.go: 保持基本不变，权限控制由中间件处理
- main.go: 在 setupRouter() 中配置角色和应用中间件

## 具体要求
- REQ-1: 定义 Role 和 Permission 类型
- REQ-2: RBACManager 支持 DefineRole(role, permissions) 定义角色权限
- REQ-3: RBACManager 支持 AssignRole(userID, role) 分配角色
- REQ-4: RBACManager 支持 HasPermission(userID, perm) 检查权限
- REQ-5: 支持角色继承 — admin 继承 editor 的所有权限，editor 继承 viewer 的所有权限
- REQ-6: RBACMiddleware 从 X-User-ID header 提取用户身份
- REQ-7: 未认证请求（无 X-User-ID）返回 401
- REQ-8: 无权限请求返回 403
- REQ-9: 预定义三个角色：viewer(读)、editor(读写)、admin(读写删+管理)
- REQ-10: 权限列表：todos:read, todos:create, todos:update, todos:delete, admin:manage
- REQ-11: RBACManager 必须并发安全
- REQ-12: RevokeRole(userID, role) 支持撤销角色

## 角色继承关系
```
admin → editor → viewer
  |        |        |
  admin:manage  todos:create  todos:read
               todos:update
               todos:delete
```

## 约束
- setupRouter() 函数签名不可更改
- 纯 stdlib，不引入外部依赖
- 用户身份从 X-User-ID header 获取（简化认证）
- /health 端点不需要认证

## 测试策略
- viewer 可以 GET /todos，不能 POST/DELETE
- editor 可以 GET/POST/PUT，不能 DELETE
- admin 可以执行所有操作
- 无 X-User-ID 返回 401
- 有 X-User-ID 但无角色返回 403
- 角色继承正确传递
- 撤销角色后权限立即失效
- 并发分配/检查角色安全

## 不做什么
- 不实现真正的认证（JWT/OAuth）
- 不实现 ABAC（基于属性的访问控制）
- 不持久化角色数据
- 不实现 API 管理端点（通过代码配置）
