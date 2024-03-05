package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Debug options
var (
	PrintTokens  bool
	PrintAstJson bool
	PrintAstTree bool
)

type Regexl struct {
	Query          string
	CompiledRegexp *regexp.Regexp
}

func NewRegexl(query string) *Regexl {

	rl := &Regexl{
		Query: query,
	}

	return rl
}

func (rl *Regexl) Compile() error {

	parser := NewParser(rl.Query)

	// Tokenize
	tokens, err := parser.Tokenize()
	if err != nil {
		return err
	}

	if PrintTokens {

		b, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("%d Tokens: %s\n", len(tokens), string(b))
	}

	if len(tokens) == 0 {
		return fmt.Errorf("empty query is not allowed")
	}

	// Gen AST
	ast := NewAst(tokens)
	err = ast.Gen()
	if err != nil {
		return err
	}

	if PrintAstJson {
		fmt.Printf("AST JSON: %+v\n", ast.Nodes)
	}

	if PrintAstTree {
		ast.PrintTree()
	}

	gb := &GoBackend{}
	goRegexp, _, err := gb.AstToGoRegex(ast)
	if err != nil {
		return err
	}

	rl.CompiledRegexp = goRegexp
	return nil
}

func (rl *Regexl) MustCompile() *Regexl {

	err := rl.Compile()
	if err != nil {
		panic(err)
	}

	return rl
}
