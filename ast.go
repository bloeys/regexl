package regexl

import "fmt"

var _ fmt.Stringer = &Ast{}

type Ast struct {
}

func NewAstFromTokens(tokens []Token) *Ast {

	return &Ast{}
}

func (a *Ast) GenAst() error {

	return nil

}

func (a *Ast) String() string {
	// @TODO
	return ""
}
