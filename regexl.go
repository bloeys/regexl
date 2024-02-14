package regexl

import (
	"encoding/json"
	"fmt"
)

var (
	IsVerbose = false
)

type Regexl struct {
	Query string
}

func (rl *Regexl) Compile() error {

	parser := NewParser(rl.Query)

	// Tokenize
	tokens, err := parser.Tokenize()
	if err != nil {
		return err
	}

	if IsVerbose {

		b, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%d Tokens: %s\n", len(tokens), string(b))
	}

	// Gen AST
	ast := NewAst(tokens)
	if IsVerbose {
		fmt.Printf("AST: %s\n", ast.String())
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
