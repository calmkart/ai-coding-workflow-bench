# Task: 实现数学表达式解析器

## 目标
实现完整的数学表达式解析器：tokenize → parse → evaluate 三阶段管道。支持 +, -, *, /, 括号, 一元负号。

## 当前状态
- 只有空的接口骨架

## 变更范围
- parser.go: 完整实现

## 具体要求
- REQ-1: Tokenize(input) 将字符串分解为 Token 序列
- REQ-2: Token 类型：Number, Plus, Minus, Star, Slash, LParen, RParen, EOF
- REQ-3: Parse(tokens) 使用递归下降解析器构建 AST
- REQ-4: 运算符优先级：括号 > 一元负号 > 乘除 > 加减
- REQ-5: Evaluate(expr) 对 AST 求值返回 float64
- REQ-6: Calc(input) 一步完成 tokenize → parse → evaluate
- REQ-7: AST 节点类型：NumberExpr, BinaryExpr, UnaryExpr
- REQ-8: 支持整数和浮点数（如 3.14）
- REQ-9: 支持嵌套括号
- REQ-10: 除以零返回错误
- REQ-11: 非法输入返回有意义的错误信息
- REQ-12: 支持多个空格和前后空格

## 语法（EBNF）
```
expression = term (('+' | '-') term)*
term       = factor (('*' | '/') factor)*
factor     = '-' factor | atom
atom       = NUMBER | '(' expression ')'
NUMBER     = [0-9]+ ('.' [0-9]+)?
```

## 约束
- 纯 stdlib
- 不使用 regexp
- 手写 tokenizer 和 parser

## 测试策略
- 基本四则运算
- 运算符优先级
- 括号嵌套
- 一元负号
- 浮点数
- 除以零
- 非法输入
- 空格处理
- 复杂表达式

## 不做什么
- 不实现变量
- 不实现函数调用（sin/cos）
- 不实现赋值
