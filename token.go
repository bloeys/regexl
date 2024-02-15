package regexl

import "fmt"

//go:generate stringer -type=TokenType
type TokenType int

func (tt TokenType) MarshalText() (text []byte, err error) {
	return []byte(tt.String()), nil
}

var _ fmt.Stringer = TokenType_Unknown

const (
	TokenType_Unknown TokenType = iota
	TokenType_Space
	TokenType_String
	// TokenType_Single_Quote
	TokenType_Int
	TokenType_Float
	TokenType_Operator
	TokenType_OpenBracket
	TokenType_CloseBracket
	TokenType_OpenCurlyBracket
	TokenType_CloseCurlyBracket
	TokenType_Colon
	TokenType_Comma
	TokenType_Bool
	TokenType_Plus
	TokenType_Comment
	TokenType_Object_Param
	TokenType_Function_Name
	TokenType_Keyword
)

type TokenPos int

type Token struct {
	Val  string
	Type TokenType
	Pos  TokenPos
}

func (t *Token) MakeEmpty() {

	if t == nil {
		return
	}

	t.Val = ""
	t.Type = TokenType_Unknown
	t.Pos = -1
}

func (t *Token) IsEmpty() bool {
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "" && t.Pos == -1)
}

func (t *Token) HasLoc() bool {
	return t.Pos != -1
}
