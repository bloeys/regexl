package regexl

import (
	"fmt"
)

const (
	AST_INVALID_TOKEN_INDEX = -1
)

var _ fmt.Stringer = &Ast{}

type Ast struct {
	Tokens []Token
	Nodes  []Node
}

var _ error = &AstError{}

type AstError struct {
	Err error
	Pos TokenPos
}

func (te *AstError) Error() string {

	if te == nil || te.Err == nil {
		return ""
	}

	return fmt.Sprintf("ast error: loc=%d; err=%s", te.Pos, te.Err.Error())
}

//
// Ast structure similar to what the go/ast package does as I really like their setup
//

type Node interface {
	// StartPos is the position of the first byte of the first character making up this node
	StartPos() TokenPos
	// EndPos is the position of the first byte of the first character that doesn't belong to this node.
	// This means EndPos is +1 of the last character, so it acts in the same way len() does
	EndPos() TokenPos
}

//
// Statements
//

type Stmt interface {
	Node
	stmt()
}

// QueryStmt is used as the root node and pos is the position of the first statement
type QueryStmt struct {
	Pos        TokenPos
	Statements []Stmt
}

func (s *QueryStmt) stmt()              {}
func (s *QueryStmt) StartPos() TokenPos { return s.Pos }
func (s *QueryStmt) EndPos() TokenPos {

	if len(s.Statements) == 0 {
		return s.Pos + 1
	}

	return s.Statements[len(s.Statements)-1].EndPos()
}

type ExprStmt struct {
	E Expr
}

func (s *ExprStmt) stmt()              {}
func (s *ExprStmt) StartPos() TokenPos { return s.E.StartPos() }
func (s *ExprStmt) EndPos() TokenPos   { return s.E.EndPos() }

type SelectStmt struct {
	Pos  TokenPos
	Type TokenType
	Es   []Expr
}

func (s *SelectStmt) stmt()              {}
func (s *SelectStmt) StartPos() TokenPos { return s.Pos }
func (s *SelectStmt) EndPos() TokenPos {

	if len(s.Es) == 0 {
		return s.Pos + 1
	}

	return s.Es[len(s.Es)-1].EndPos()
}

//
// Expressions
//

type Expr interface {
	Node
	expr()
}

type IdentExpr struct {
	Name string
	Pos  TokenPos
}

func (e *IdentExpr) expr()              {}
func (e *IdentExpr) StartPos() TokenPos { return e.Pos }
func (e *IdentExpr) EndPos() TokenPos   { return e.Pos + TokenPos(len(e.Name)) }

type FuncExpr struct {
	Pos          TokenPos
	Ident        IdentExpr
	Args         []Expr
	OpenBracket  TokenPos
	CloseBracket TokenPos
}

func (e *FuncExpr) expr()              {}
func (e *FuncExpr) StartPos() TokenPos { return e.Pos }
func (e *FuncExpr) EndPos() TokenPos   { return e.CloseBracket + 1 }

type BinaryExpr struct {
	Pos  TokenPos
	Type TokenType
	Lhs  Expr
	Rhs  Expr
}

func (e *BinaryExpr) expr()              {}
func (e *BinaryExpr) StartPos() TokenPos { return e.Pos }
func (e *BinaryExpr) EndPos() TokenPos   { return e.Rhs.EndPos() }

type LiteralExpr struct {
	Pos  TokenPos
	Type TokenType
	// Value depends on the type, so it can contain a numeric, string etc
	Value string
}

func (e *LiteralExpr) expr()              {}
func (e *LiteralExpr) StartPos() TokenPos { return e.Pos }
func (e *LiteralExpr) EndPos() TokenPos   { return e.Pos + TokenPos(len(e.Value)) }

func NewAst(tokens []Token) *Ast {

	ast := &Ast{
		Tokens: tokens,
		Nodes:  make([]Node, 0, 5),
	}

	return ast
}

func (a *Ast) Gen() error {

	i := 0
	for i < len(a.Tokens) {

		n, lastProcessedIndex, err := a.parseFrom(i)
		if err != nil {
			return err
		}

		a.Nodes = append(a.Nodes, n)
		i = lastProcessedIndex + 1
	}

	return nil
}

func (a *Ast) parseFrom(tokenIndex int) (node Node, lastProcessedIndex int, err error) {

	if tokenIndex < 0 || tokenIndex >= len(a.Tokens) {
		panic(fmt.Sprintf("gen ast failed as the passed index '%d' is out of range for the tokens which have len=%d", tokenIndex, len(a.Tokens)))
	}

loopLbl:
	for i := 0; i < len(a.Tokens); i++ {

		t := &a.Tokens[i]

		switch t.Type {

		case TokenType_Function_Name:
			node, lastProcessedIndex, err = a.parseFunc(i)
			break loopLbl

			// Handle literals
		case TokenType_Int:
			fallthrough
		case TokenType_Float:
			fallthrough
		case TokenType_String:
			err = nil
			node = &LiteralExpr{
				Pos:   t.Pos,
				Type:  t.Type,
				Value: t.Val,
			}
			lastProcessedIndex = i
			break loopLbl

		case TokenType_Keyword:
			node, lastProcessedIndex, err = a.parseSelect(i)
			break loopLbl

		default:
			// return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			// 	Err: fmt.Errorf("only certain functions (e.g. set_options()) and keywords (e.g. select) are allowed at the top-level of a query, but found a token of type '%s'", t.Type),
			// 	Pos: t.Pos,
			// }
		}
	}

	if err != nil {
		return nil, AST_INVALID_TOKEN_INDEX, err
	}

	// Handle binary ops
	nextT := a.GetToken(lastProcessedIndex + 1)
	if nextT != nil && nextT.Type == TokenType_Plus {

		rhs, rhsLastProcessedIndex, err := a.parseFrom(lastProcessedIndex + 2)
		if err != nil {
			return nil, AST_INVALID_TOKEN_INDEX, err
		}

		lhsExpr, ok := node.(Expr)
		if !ok {
			return nil, AST_INVALID_TOKEN_INDEX, &AstError{
				Pos: node.StartPos(),
				Err: fmt.Errorf("left side of binary operator '+' at pos=%d is not an expression. Lhs=%+v", node.StartPos(), node),
			}
		}

		rhsExpr, ok := rhs.(Expr)
		if !ok {
			return nil, AST_INVALID_TOKEN_INDEX, &AstError{
				Pos: node.StartPos(),
				Err: fmt.Errorf("right side of binary operator '+' at pos=%d is not an expression. Rhs=%+v", rhs.StartPos(), rhs),
			}
		}

		return &BinaryExpr{
			Pos:  nextT.Pos,
			Type: nextT.Type,
			Lhs:  lhsExpr,
			Rhs:  rhsExpr,
		}, rhsLastProcessedIndex, nil
	}

	// Generic return
	return node, lastProcessedIndex, nil
}

func (a *Ast) parseSelect(tokenIndex int) (sStmt *SelectStmt, lastProcessedToken int, err error) {

	selectToken := a.GetToken(tokenIndex)
	if selectToken == nil {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Err: fmt.Errorf("failed to find select token using index=%d", tokenIndex),
		}
	}

	if selectToken.Type != TokenType_Keyword || selectToken.Val != "select" {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Err: fmt.Errorf("parseSelect failed because it was invoked on a token at index=%d which is not a select keyword (probably a bug in the code). Token=%+v", tokenIndex, selectToken),
		}
	}

	sStmt = &SelectStmt{
		Pos:  selectToken.Pos,
		Type: TokenType_Keyword,
		Es:   make([]Expr, 0, 10),
	}

	for i := tokenIndex + 1; i < len(a.Tokens); i++ {

		t := &a.Tokens[i]

		node, lastProcessedToken, err := a.parseFrom(i)
		if err != nil {
			return nil, AST_INVALID_TOKEN_INDEX, err
		}

		expr, ok := node.(Expr)
		if !ok {
			return nil, AST_INVALID_TOKEN_INDEX, &AstError{
				Pos: t.Pos,
				Err: fmt.Errorf("select has a non-expression (i.e. not a function, not a literal like string etc) in front of it. Node after select=%+v", node),
			}
		}

		sStmt.Es = append(sStmt.Es, expr)
		i = lastProcessedToken
	}

	return sStmt, lastProcessedToken, nil
}

func (a *Ast) parseFunc(tokenIndex int) (fExpr *FuncExpr, lastProcessedToken int, err error) {

	funcToken := a.GetToken(tokenIndex)
	if funcToken == nil {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Err: fmt.Errorf("failed to find function token using index=%d", tokenIndex),
		}
	}

	if funcToken.Type != TokenType_Function_Name {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Err: fmt.Errorf("parseFunc failed because it was invoked on a token at index=%d which is not a function (probably a bug in the code). Token=%+v", tokenIndex, funcToken),
		}
	}

	fExpr = &FuncExpr{
		Pos: funcToken.Pos,
		Ident: IdentExpr{
			Name: funcToken.Val,
			Pos:  funcToken.Pos,
		},
		Args:         make([]Expr, 0, 5),
		OpenBracket:  AST_INVALID_TOKEN_INDEX,
		CloseBracket: AST_INVALID_TOKEN_INDEX,
	}

	openBracketToken := a.GetToken(tokenIndex + 1)
	if openBracketToken == nil || openBracketToken.Type != TokenType_OpenBracket {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Pos: fExpr.Ident.StartPos(),
			Err: fmt.Errorf("expected '(' after function name but found token='%+v'", openBracketToken),
		}
	}
	fExpr.OpenBracket = openBracketToken.Pos

forLoopLbl:
	for i := tokenIndex + 2; i < len(a.Tokens); i++ {

		t := &a.Tokens[i]

		switch t.Type {

		case TokenType_CloseBracket:
			fExpr.CloseBracket = t.Pos
			lastProcessedToken = i
			break forLoopLbl

		default:
			node, lastProcessedToken, err := a.parseFrom(i)
			if err != nil {
				return nil, AST_INVALID_TOKEN_INDEX, err
			}

			expr, ok := node.(Expr)
			if !ok {
				return nil, AST_INVALID_TOKEN_INDEX, &AstError{
					Pos: t.Pos,
					Err: fmt.Errorf("expected expression to be returned within arguments of a function call, but found node=%+v", node),
				}
			}

			// @TODO Handle '+' in function arguments
			fExpr.Args = append(fExpr.Args, expr)
			i = lastProcessedToken
		}
	}

	if fExpr.CloseBracket == AST_INVALID_TOKEN_INDEX {
		return nil, AST_INVALID_TOKEN_INDEX, &AstError{
			Pos: funcToken.Pos,
			Err: fmt.Errorf("function of name=%s at pos=%d does not have a closing bracket", funcToken.Val, funcToken.Pos),
		}
	}

	return fExpr, lastProcessedToken, nil
}

func (a *Ast) GetToken(index int) *Token {

	if index < 0 {
		panic(fmt.Sprintf("tried getting a token using a negative index '%d'", index))
	}

	if index >= len(a.Tokens) {
		return nil
	}

	return &a.Tokens[index]
}

func (a *Ast) String() string {
	// @TODO
	return ""
}
