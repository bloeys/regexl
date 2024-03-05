package regexl

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// @TODO: remove or make something nicer
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

// Compile tries to compile the query within this Regexl object and then sets Regexl.CompiledRegexp.
// Regexl.CompiledRegexp is only set if no error is found, otherwise the error is returned and Regexl.CompiledRegexp is unchanged.
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

// MustCompile compiles the query within this regexl object by calling Regexl.Compile and panics if an error is thrown
func (rl *Regexl) MustCompile() *Regexl {

	err := rl.Compile()
	if err != nil {
		panic(err)
	}

	return rl
}
