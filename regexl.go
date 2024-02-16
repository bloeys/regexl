package regexl

import (
	"encoding/json"
	"fmt"
)

type Regexl struct {
	Query string

	// Debug options
	PrintTokens  bool
	PrintAstJson bool
	PrintAstTree bool
}

func (rl *Regexl) Compile() error {

	parser := NewParser(rl.Query)

	// Tokenize
	tokens, err := parser.Tokenize()
	if err != nil {
		return err
	}

	if rl.PrintTokens {

		b, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%d Tokens: %s\n", len(tokens), string(b))
	}

	if len(tokens) == 0 {
		return nil
	}

	// Gen AST
	ast := NewAst(tokens)
	err = ast.Gen()
	if err != nil {
		return err
	}

	if rl.PrintAstJson {
		fmt.Printf("AST JSON: %+v\n", ast.Nodes)
	}

	if rl.PrintAstTree {
		ast.PrintTree()
	}

	return nil
}

func NewRegexl(query string) (*Regexl, error) {

	rl := &Regexl{
		Query: query,
	}

	err := rl.Compile()
	if err != nil {
		return nil, err
	}

	return rl, nil
}
