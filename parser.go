package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type TokenType int

const (
	TokenType_Unknown TokenType = iota
	TokenType_Space
	TokenType_String
	TokenType_Number
	TokenType_Operator
	TokenType_OpenBracket
	TokenType_CloseBracket
	TokenType_OpenCurlyBracket
	TokenType_CloseCurlyBracket
	TokenType_FunctionName
)

type Token struct {
	Val  string
	Type TokenType
	Loc  int
}

func (t *Token) MakeEmpty() {

	if t == nil {
		return
	}

	t.Val = ""
	t.Type = TokenType_Unknown
	t.Loc = 0
}

func (t *Token) IsEmpty() bool {
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "")
}

type Parser struct {
	Query    string
	QueryLoc int
	PError   ParserError
}

var _ error = ParserError{}

type ParserError struct {
	Err error
	Loc int
}

func (te ParserError) Error() string {

	if te.Err == nil {
		return ""
	}

	return fmt.Sprintf("parser error: loc=%d; err=%s", te.Loc, te.Err.Error())
}

func ParseQuery(query string) (*regexp.Regexp, error) {

	parser := Parser{}

	tokens, err := parser.Tokenize(query)
	if err != nil {
		return nil, err
	}

	b, _ := json.MarshalIndent(tokens, "", "  ")
	fmt.Printf("Tokens: %s\n", string(b))
	// fmt.Printf("Tokens: %+v\n", tokens)

	return nil, nil
}

func (p *Parser) Tokenize(query string) (tokens []Token, pErr *ParserError) {

	tokens = make([]Token, 0, 50)

	addToken := func(t *Token) {

		t.Val = strings.TrimSpace(t.Val)
		if t.IsEmpty() {
			return
		}

		tokens = append(tokens, *t)
		t.MakeEmpty()
	}

	inString := false
	inComment := false
	token := &Token{}
	for runeStartByteIndex, c := range query {

		if inComment {

			if c != '\n' {
				continue
			}

			inComment = false
		}

		if inString {

			if c != '\'' && c != '\\' {
				token.Val += string(c)
				continue
			}

			if c == '\'' {
				token.Val += string(c)
				inString = false
				addToken(token)
				continue
			}

			// Handle backslash in string as it might escape the end string character
			nextRune, err := p.GetNextRuneByByteIndex(runeStartByteIndex)
			if err != nil {
				return tokens, &ParserError{
					Err: err,
					Loc: runeStartByteIndex,
				}
			}

			if nextRune == '\'' {
				token.Val += string(c)
				inString = false
				addToken(token)
				continue
			}

			// Its just a normal backslash not escaping anything so we let it be
			token.Val += string(c)
			continue
		}

		switch c {
		case ' ':
			addToken(token)
		case '(':
			token.Type = TokenType_FunctionName
			addToken(token)

			token.Val = "("
			token.Type = TokenType_OpenBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case ')':
			addToken(token)
			token.Val = ")"
			token.Type = TokenType_CloseBracket
			token.Loc = runeStartByteIndex

		case '{':
			addToken(token)
			token.Val = "{"
			token.Type = TokenType_CloseCurlyBracket
			token.Loc = runeStartByteIndex

		case '}':
			addToken(token)
			token.Val = "}"
			token.Type = TokenType_CloseCurlyBracket
			token.Loc = runeStartByteIndex

		case '\n':
			addToken(token)
		case '\'':
			addToken(token)
			inString = true

			token.Val = "'"
			token.Type = TokenType_String
			token.Loc = runeStartByteIndex

		case '-':

			nextRune, err := p.GetNextRuneByByteIndex(runeStartByteIndex)
			if err != nil {
				return tokens, &ParserError{
					Err: err,
					Loc: runeStartByteIndex,
				}
			}

			if nextRune != '-' {
				return tokens, &ParserError{
					Err: fmt.Errorf("found '-' in an unexpected location. '-' can only be used for comments or in strings"),
					Loc: runeStartByteIndex,
				}
			}

			inComment = true

		default:
			token.Val += string(c)
		}
	}

	return tokens, pErr
}

func (p *Parser) GetRuneByByteIndex(index int) (rune, error) {

	r, _ := utf8.DecodeRuneInString(p.Query[index:])
	if r == utf8.RuneError {
		return 0, fmt.Errorf("decoding utf8 query failed. index=%d", index)
	}

	return r, nil
}

func (p *Parser) GetNextRuneByByteIndex(index int) (rune, error) {

	if index >= len(p.Query) {
		return 0, fmt.Errorf("getting next rune failed because index is out of range. index=%d; queryLen=%d", index, len(p.Query))
	}

	r, rLen := utf8.DecodeRuneInString(p.Query[index:])
	if r == utf8.RuneError {
		return 0, fmt.Errorf("decoding utf8 query failed. index=%d", index)
	}

	r, _ = utf8.DecodeRuneInString(p.Query[index+rLen:])
	if r == utf8.RuneError {
		return 0, fmt.Errorf("decoding utf8 query failed. index=%d", index+rLen)
	}

	return r, nil
}
