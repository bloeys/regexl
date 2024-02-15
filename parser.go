package regexl

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	keywords = []string{"select"}
)

type Parser struct {
	Query string
}

func NewParser(query string) *Parser {
	return &Parser{
		Query: query,
	}
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

func (p *Parser) Tokenize() (tokens []Token, err error) {

	if p.Query == "" {
		return []Token{}, nil
	}

	tokens = make([]Token, 0, 50)

	getToken := func(index int) *Token {

		if len(tokens) == 0 {
			return nil
		}

		if index < 0 {
			return &tokens[len(tokens)+index]
		}

		return &tokens[index]
	}

	/*
		Trims surrounding space of the token's value, then if the token is not empty it appends a copy of it to the list of tokens, and then makes the passed token empty.

		Returns the last appended token using getToken(-1). The just passed token is returned (its copy appended to the array) if the passed token wasn't empty, otherwise the last appended token is returned.
		The return can be null if no tokens have been appended
	*/
	addToken := func(t *Token) (latestToken *Token) {

		t.Val = strings.TrimSpace(t.Val)
		if t.IsEmpty() {
			return getToken(-1)
		}

		tokens = append(tokens, *t)
		t.MakeEmpty()
		return getToken(-1)
	}

	/*
		Try to assign a type to a token that is probably some literal like a string or number, and is no-op if token is already typed.

		Possible callers are:
		* ',' or '}' of an object param
		* ')' closing a function
	*/
	tryAssignTypeToPossibleLiteralToken := func(t *Token) {

		if t.IsEmpty() || t.Type != TokenType_Unknown {
			return
		}

		trimmedVal := strings.TrimSpace(t.Val)
		if trimmedVal == "false" || trimmedVal == "true" {
			t.Type = TokenType_Bool
		} else if _, err := strconv.ParseInt(trimmedVal, 10, 64); err == nil {
			t.Type = TokenType_Int
		} else if _, err := strconv.ParseFloat(trimmedVal, 64); err == nil {
			t.Type = TokenType_Float
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

			prevToken := addToken(token)
			tryAssignTypeToPossibleLiteralToken(prevToken)

			// Handle keywords and use is empty to protect against nil
			if !prevToken.IsEmpty() && prevToken.Type == TokenType_Unknown {

				if slices.Contains(keywords, prevToken.Val) {
					prevToken.Type = TokenType_Keyword
				}
			}

		case ':':

			prevToken := addToken(token)
			if !prevToken.IsEmpty() {
				prevToken.Type = TokenType_Object_Param
			}

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

			prevToken := addToken(token)
			tryAssignTypeToPossibleLiteralToken(prevToken)

			token.Val = ","
			token.Type = TokenType_Comma
			token.Loc = runeStartByteIndex
			addToken(token)

		case '(':
			prevToken := addToken(token)
			if !prevToken.IsEmpty() {
				prevToken.Type = TokenType_Function_Name
			}

			token.Val = "("
			token.Type = TokenType_OpenBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case ')':
			prevToken := addToken(token)
			tryAssignTypeToPossibleLiteralToken(prevToken)

			token.Val = ")"
			token.Type = TokenType_CloseBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case '{':
			addToken(token)

			token.Val = "{"
			token.Type = TokenType_OpenCurlyBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case '}':
			prevToken := addToken(token)
			tryAssignTypeToPossibleLiteralToken(prevToken)

			token.Val = "}"
			token.Type = TokenType_CloseCurlyBracket
			token.Loc = runeStartByteIndex
			addToken(token)

		case '\'':
			addToken(token)

			// token.Val = "'"
			// token.Type = TokenType_Single_Quote
			// token.Loc = runeStartByteIndex
			// addToken(token)

			inString = true
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

	err = p.ValidateTokens(tokens)
	return tokens, err
}

func (p *Parser) ValidateTokens(tokens []Token) error {

	type OpenBracketsList struct {
		Bracket *Token
		Next    *OpenBracketsList
	}

	bracketsCounter := 0
	openBracketsList := &OpenBracketsList{}

	addOpenBracket := func(b *Token) {

		if b.Type != TokenType_OpenBracket && b.Type != TokenType_OpenCurlyBracket {
			panic("can't add anything other than open bracket or open curly bracket to open brackets list")
		}

		// If first is empty use it
		if openBracketsList.Bracket == nil {
			openBracketsList.Bracket = b
			return
		}

		// Go to end of list and add the new item
		var curr *OpenBracketsList = openBracketsList
		for curr.Next != nil {
			curr = curr.Next
		}

		curr.Next = &OpenBracketsList{
			Bracket: b,
			Next:    nil,
		}
	}

	removeLastBracket := func() {

		var curr *OpenBracketsList = openBracketsList

		// Handle one item remaining
		if curr.Next == nil {
			curr.Bracket = nil
		}

		// Remove from end of list
		for curr.Next != nil && curr.Next.Next != nil {
			curr = curr.Next
		}

		curr.Next = nil
	}

	for i := 0; i < len(tokens); i++ {

		t := &tokens[i]

		switch t.Type {

		case TokenType_Unknown:
			return &ParserError{
				Err: fmt.Errorf("invalid regexl query: found token with type=unknown after tokenization; token=%+v; query=%s", t, p.Query),
				Loc: t.Loc,
			}

		case TokenType_OpenBracket:
			fallthrough
		case TokenType_OpenCurlyBracket:
			bracketsCounter++
			addOpenBracket(t)

		case TokenType_CloseBracket:
			fallthrough
		case TokenType_CloseCurlyBracket:

			bracketsCounter--
			removeLastBracket()
			if bracketsCounter >= 0 {
				continue
			}

			if t.Type == TokenType_CloseCurlyBracket {
				return &ParserError{
					Err: fmt.Errorf("invalid regexl query: found a closed curly bracket without an opening curly bracket; token=%+v; query=%s", t, p.Query),
					Loc: t.Loc,
				}
			}

			return &ParserError{
				Err: fmt.Errorf("invalid regexl query: found a closed bracket without an opening bracket; token=%+v; query=%s", t, p.Query),
				Loc: t.Loc,
			}
		}
	}

	// Negative case is handled inside the switch case, so this is for brackets that opened but didn't close
	// We only show the first unclosed bracket, but we can use the tree to show all of them
	if bracketsCounter != 0 {
		return &ParserError{
			Err: fmt.Errorf("invalid regexl query: found an opening bracket without a closing bracket pair; first unclosed bracket token=%+v; query=%s", openBracketsList.Bracket, p.Query),
			Loc: openBracketsList.Bracket.Loc,
		}
	}

	return nil
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
