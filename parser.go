package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	keywords = []string{"for", "select"}
)

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
		fmt.Printf("%d Tokens: %s\n", len(tokens), string(b))
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

	/*
		Try to assign a type to a token that is probably some literal like a string or number, and is no-op if token is already typed.

		Possible callers are:
		* ',' or '}' of an object param
		* ')' closing a function
	*/
	tryAssignTypeToPossibleLiteralToken := func(t *Token) {

		if t.Type != TokenType_Unknown || t.IsEmpty() {
			return
		}

		trimmedVal := strings.TrimSpace(t.Val)
		if trimmedVal == "false" || trimmedVal == "true" {
			t.Type = TokenType_Bool
		} else if _, err := strconv.ParseFloat(trimmedVal, 64); err == nil {
			t.Type = TokenType_Number
		}
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

		// General token 'end' signifiers
		case '\n':
			fallthrough
		case '\t':
			fallthrough
		case ' ':

			tryAssignTypeToPossibleLiteralToken(token)

			// Handle keywords
			if token.Type == TokenType_Unknown {

				if slices.Contains(keywords, token.Val) {
					token.Type = TokenType_Keyword
				}
			}

			addToken(token)

		case ':':
			token.Type = TokenType_Object_Param_Key
			addToken(token)

			token.Val = ":"
			token.Type = TokenType_Colon
			token.Loc = runeStartByteIndex
			addToken(token)

		case '+':
			addToken(token)

			token.Val = "+"
			token.Type = TokenType_Plus
			token.Loc = runeStartByteIndex
			addToken(token)

		case ',':

			// Try to assign a type to previous value (this is for when ',' is after an object param value)
			tryAssignTypeToPossibleLiteralToken(token)
			addToken(token)

			token.Val = ","
			token.Type = TokenType_Comma
			token.Loc = runeStartByteIndex
			addToken(token)

		case '(':
			token.Type = TokenType_Function_Name
			addToken(token)

			token.Val = "("
			token.Type = TokenType_OpenBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case ')':
			tryAssignTypeToPossibleLiteralToken(token)
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
			tryAssignTypeToPossibleLiteralToken(token)
			addToken(token)
			token.Val = "}"
			token.Type = TokenType_CloseCurlyBracket
			token.Loc = runeStartByteIndex

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

	if pErr != nil {
		return tokens, pErr
	}

	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]
		if t.Type == TokenType_Unknown {
			return tokens, &ParserError{
				Err: fmt.Errorf("invalid regexl query: found token with type=unknown after tokenization; token=%+v; query=%s", t, p.Query),
				Loc: t.Loc,
			}
		}
	}

	return tokens, nil
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
