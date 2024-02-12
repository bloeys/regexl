package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

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
	t.Loc = -1
}

func (t *Token) IsEmpty() bool {
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "")
}

func (t *Token) HasLoc() bool {
	return t.Loc != -1
}

type Parser struct {
	Query  string
	PError ParserError
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

	parser := Parser{
		Query: query,
	}

	tokens, err := parser.Tokenize()
	if err != nil {
		return nil, err
	}

	b, _ := json.MarshalIndent(tokens, "", "  ")

	if IsVerbose {
		fmt.Printf("Tokens: %s\n", string(b))
	}

	return nil, nil
}

func (p *Parser) Tokenize() (tokens []Token, pErr *ParserError) {

	if p.Query == "" {
		return []Token{}, nil
	}

	tokens = make([]Token, 0, 50)

	// getToken := func(index int) *Token {

	// 	if len(tokens) == 0 {
	// 		return nil
	// 	}

	// 	if index < 0 {
	// 		return &tokens[len(tokens)+index]
	// 	}

	// 	return &tokens[index]
	// }

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
	token.MakeEmpty()
	for runeStartByteIndex, c := range p.Query {

		if inComment {

			if c != '\n' {
				token.Val += string(c)
				continue
			}

			// Remove the second '-' of the comment start
			token.Val = token.Val[1:]
			addToken(token)
			inComment = false
		}

		if inString {

			if c != '\'' && c != '\\' {
				token.Val += string(c)
				continue
			}

			if c == '\'' {

				addToken(token)

				// token.Val = "'"
				// token.Type = TokenType_Single_Quote
				// token.Loc = runeStartByteIndex
				// addToken(token)

				inString = false
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

		case '\t':
			fallthrough
		case ' ':
			addToken(token)

		case ':':
			token.Type = TokenType_Object_Param_Key
			addToken(token)

			token.Val = ":"
			token.Type = TokenType_Colon
			token.Loc = runeStartByteIndex
			addToken(token)

		case ',':

			// Try to assign a type to previous value (this is for when ',' is after an object param value)
			if token.Type == TokenType_Unknown {

				trimmedVal := strings.TrimSpace(token.Val)
				if trimmedVal == "false" || trimmedVal == "true" {
					token.Type = TokenType_Bool
				} else if _, err := strconv.ParseFloat(trimmedVal, 64); err == nil {
					token.Type = TokenType_Number
				}
			}
			addToken(token)

			token.Val = ","
			token.Type = TokenType_Comma
			token.Loc = runeStartByteIndex
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
			token.Type = TokenType_OpenCurlyBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case '}':
			addToken(token)
			token.Val = "}"
			token.Type = TokenType_CloseCurlyBracket
			token.Loc = runeStartByteIndex

		case '\n':
			addToken(token)

		case '\'':
			addToken(token)

			// token.Val = "'"
			// token.Type = TokenType_Single_Quote
			// token.Loc = runeStartByteIndex
			// addToken(token)

			inString = true
			token.Val = ""
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

			addToken(token)
			token.Type = TokenType_Comment
			token.Loc = runeStartByteIndex
			inComment = true

		default:
			token.Val += string(c)
			if !token.HasLoc() {
				token.Loc = runeStartByteIndex
			}
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
