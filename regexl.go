package regexl

var (
	IsVerbose = false
)

type Regexl struct {
	Query string
}

func (rl *Regexl) Compile() error {

	_, err := ParseQuery(rl.Query)
	if err != nil {
		return err
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
