package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type Regexl struct {
	Query string

	// Debug options
	PrintTokens  bool
	PrintAstJson bool
	PrintAstTree bool
}

func NewRegexl(query string) *Regexl {

	rl := &Regexl{
		Query: query,
	}

	return rl
}

func (rl *Regexl) Compile() (*regexp.Regexp, error) {

	parser := NewParser(rl.Query)

	// Tokenize
	tokens, err := parser.Tokenize()
	if err != nil {
		return nil, err
	}

	if rl.PrintTokens {

		b, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return nil, err
		}

		fmt.Printf("%d Tokens: %s\n", len(tokens), string(b))
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty query is not allowed")
	}

	// Gen AST
	ast := NewAst(tokens)
	err = ast.Gen()
	if err != nil {
		return nil, err
	}

	if rl.PrintAstJson {
		fmt.Printf("AST JSON: %+v\n", ast.Nodes)
	}

	if rl.PrintAstTree {
		ast.PrintTree()
	}

	gb := &GoBackend{}
	goRegexp, err := gb.AstToGoRegex(ast)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Regex: %s\n\n", goRegexp.String())

	return goRegexp, err
}
