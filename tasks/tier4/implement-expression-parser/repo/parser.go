package exprparser

import (
	"errors"
	"fmt"
)

// TokenType identifies the type of a token.
type TokenType int

const (
	TokenNumber TokenType = iota
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenLParen
	TokenRParen
	TokenEOF
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Value   string
	Pos     int
}

// Expr represents a node in the AST.
type Expr interface {
	exprNode()
}

// NumberExpr is a numeric literal.
type NumberExpr struct {
	Value float64
}
func (NumberExpr) exprNode() {}

// BinaryExpr is a binary operation (e.g., a + b).
type BinaryExpr struct {
	Op    TokenType
	Left  Expr
	Right Expr
}
func (BinaryExpr) exprNode() {}

// UnaryExpr is a unary operation (e.g., -a).
type UnaryExpr struct {
	Op      TokenType
	Operand Expr
}
func (UnaryExpr) exprNode() {}

var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrInvalidInput   = errors.New("invalid input")
)

// Tokenize splits the input string into tokens.
// TODO: Implement lexical analysis.
func Tokenize(input string) ([]Token, error) {
	return nil, fmt.Errorf("tokenize: %w", ErrInvalidInput)
}

// Parse builds an AST from tokens using recursive descent.
// TODO: Implement parser following the grammar:
//   expression = term (('+' | '-') term)*
//   term       = factor (('*' | '/') factor)*
//   factor     = '-' factor | atom
//   atom       = NUMBER | '(' expression ')'
func Parse(tokens []Token) (Expr, error) {
	return nil, fmt.Errorf("parse: %w", ErrInvalidInput)
}

// Evaluate computes the value of an AST.
// TODO: Implement tree-walking evaluator.
func Evaluate(expr Expr) (float64, error) {
	return 0, fmt.Errorf("evaluate: %w", ErrInvalidInput)
}

// Calc is a convenience function: tokenize → parse → evaluate.
func Calc(input string) (float64, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return 0, err
	}
	expr, err := Parse(tokens)
	if err != nil {
		return 0, err
	}
	return Evaluate(expr)
}
