package regexl

import (
	"fmt"
)

var _ fmt.Stringer = &Ast{}

type Ast struct {
	Tokens []Token
	First  *Node
}

//
// Ast structure similar to what the go/ast package does as I really like their setup
//

type Node interface {
	Loc() int
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

func NewAst(tokens []Token) *Ast {

	ast := &Ast{
		Tokens: tokens,
	}

	return ast
}

func (a *Ast) Gen() error {

	if len(a.Tokens) == 0 {
		return nil
	}

	return nil
}

func (a *Ast) String() string {
	// @TODO
	return ""
}
