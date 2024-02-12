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
	TokenType_Number
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
	TokenType_Object_Param_Key
	TokenType_Function_Name
	TokenType_Keyword
)
