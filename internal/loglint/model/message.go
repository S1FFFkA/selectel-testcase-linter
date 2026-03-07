package model

import "go/ast"

type LogMessage struct {
	Expr       ast.Expr
	IsFormat   bool
	StaticText string
	IsConst    bool
}
