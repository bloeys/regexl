package regexl

import "fmt"

type TokenType int

const (
	TokenType_Unknown TokenType = iota
	TokenType_Space
	TokenType_String
	TokenType_Number
	TokenType_Operator
	TokenType_OpenBracket
	TokenType_CloseBracket
)

type Token struct {
	Val  string
	Type TokenType
	Loc  int32
}

var _ error = TokenizerError{}

type TokenizerError struct {
	Err error
	Loc int32
}

func (te TokenizerError) Error() string {

	if te.Err == nil {
		return ""
	}

	return fmt.Sprintf("error: tokenizer: loc=%d; err=%s", te.Loc, te.Err.Error())
}

func TokenizeRegexlQuery(query string) (tokens []Token, err error) {

	tokens = make([]Token, 0, 20)

	addToken := func(t Token) {

		if t.Type == TokenType_Space {
			return
		}

		tokens = append(tokens, t)
	}

	for loc, c := range query {

		switch c {
		case ' ':
			addToken(Token{Val: " ", Type: TokenType_Space, Loc: int32(loc)})
		default:
		}
	}

	return tokens, err
}
