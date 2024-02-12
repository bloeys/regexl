package regexl

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
	return t == nil || (t.Type == TokenType_Unknown && t.Val == "" && t.Loc == -1)
}

func (t *Token) HasLoc() bool {
	return t.Loc != -1
}
