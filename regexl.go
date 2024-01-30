package regexl

var (
	IsVerbose = false
)

func (t *Token) IsEmpty() bool {
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "")
}

type AstNode struct {
	Type TokenType
	Val  string

	Left  *AstNode
	Right *AstNode
}

type Regexl struct {
	Query string
}

func (rl *Regexl) Compile() error {

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
