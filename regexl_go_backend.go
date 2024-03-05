package regexl

import (
	"fmt"
	"regexp"
	"strings"
)

type RegexOptions struct {
	CaseSensitive  bool
	FindAllMatches bool
}

// GoBackend produces valid Go regex strings, based on the rules here: https://pkg.go.dev/regexp/syntax
type GoBackend struct {
	Opts RegexOptions
}

func (gb *GoBackend) AstToGoRegex(ast *Ast) (*regexp.Regexp, string, error) {

	if len(ast.Nodes) == 0 {
		return nil, "", fmt.Errorf("ast must have at least one node")
	}

	var err error
	regexString := ""

	for i := 0; i < len(ast.Nodes); i++ {

		switch typedNode := ast.Nodes[i].(type) {

		case *FuncExpr:

			if typedNode.Ident.Name != "set_options" {
				return nil, "", fmt.Errorf("only the function 'set_options' can be used at the top level")
			}

			_, err := gb.execFunc(typedNode)
			if err != nil {
				return nil, "", err
			}

		case *SelectStmt:

			regexString, err = gb.nodeToGoRegex(typedNode)
			if err != nil {
				return nil, "", err
			}

		default:
			return nil, "", fmt.Errorf("only 'select' and the 'set_options' function can be at the top level")
		}
	}

	regexString = gb.ApplyOptionsToRegexString(regexString)
	regexp, err := regexp.Compile(regexString)
	if err != nil {
		return regexp, regexString, fmt.Errorf("compiling regexp failed. Query=%s; Err=%s", regexString, err.Error())
	}

	return regexp, regexString, nil
}

func (gb *GoBackend) nodeToGoRegex(n Node) (out string, err error) {

	switch typedNode := n.(type) {

	case *SelectStmt:

		for i := 0; i < len(typedNode.Es); i++ {

			regexStr, err := gb.nodeToGoRegex(typedNode.Es[i])
			if err != nil {
				return "", err
			}

			out += regexStr
		}

		return out, nil

	case *BinaryExpr:

		lhsStr, err := gb.nodeToGoRegex(typedNode.Lhs)
		if err != nil {
			return "", err
		}

		rhsStr, err := gb.nodeToGoRegex(typedNode.Rhs)
		if err != nil {
			return "", err
		}

		return lhsStr + rhsStr, nil

	case *FuncExpr:
		return gb.execFunc(typedNode)

	case *LiteralExpr:
		return gb.escapeString(typedNode.Value), nil

	default:
		return "", fmt.Errorf("unhandled node type in GoBackend.AstToGoRegex. Node=%+v", n)
	}
}

func (gb *GoBackend) execFunc(fExpr *FuncExpr) (out string, err error) {

	switch fExpr.Ident.Name {

	case "set_options":

		// Loop over args and change state depending on each
		for i := 0; i < len(fExpr.Args); i++ {

			switch typedArg := fExpr.Args[i].(type) {

			case *ObjectLiteralExpr:

				for i := 0; i < len(typedArg.KeyVals); i++ {

					kva := &typedArg.KeyVals[i]
					valStr, err := gb.nodeToGoRegex(kva.Val)
					if err != nil {
						return "", err
					}

					switch kva.Key.Name {

					case "case_sensitive":
						flagVal, err := gb.stringToBool(valStr)
						if err != nil {
							return "", fmt.Errorf("invalid value for case_sensitive. err=%s", err)
						}

						gb.Opts.CaseSensitive = flagVal

					case "find_all_matches":
						flagVal, err := gb.stringToBool(valStr)
						if err != nil {
							return "", fmt.Errorf("invalid value for find_all_matches. err=%s", err)
						}

						gb.Opts.FindAllMatches = flagVal

					default:
						return "", fmt.Errorf("unknown parameter '%s' in the function %s", kva.Key.Name, fExpr.Ident.Name)
					}
				}

			default:
				return "", fmt.Errorf("only one passed object (e.g. {case_sensitive:true}) is allowed as input to the function %s", fExpr.Ident.Name)
			}
		}

	case "any_strings_of":

		for i := 0; i < len(fExpr.Args); i++ {

			regexString, err := gb.nodeToGoRegex(fExpr.Args[i])
			if err != nil {
				return "", err
			}

			if i == len(fExpr.Args)-1 {
				out += regexString
			} else {
				out += regexString + "|"
			}
		}

	case "any_chars_of":

		if len(fExpr.Args) == 0 {
			break
		}

		out += "["
		for i := 0; i < len(fExpr.Args); i++ {

			regexString, err := gb.nodeToGoRegex(fExpr.Args[i])
			if err != nil {
				return "", err
			}

			out += regexString
		}
		out += "]"

	case "starts_with":

		if len(fExpr.Args) != 1 {
			return "", fmt.Errorf("function '%s' must have one argument but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		regexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		out += "^" + regexString

	case "ends_with":

		if len(fExpr.Args) != 1 {
			return "", fmt.Errorf("function '%s' must have one argument but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		regexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		out += regexString + "$"

	case "any_chars":

		if len(fExpr.Args) != 0 {
			return "", fmt.Errorf("function '%s' must have no arguments but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		out += ".*"

	case "zero_plus_of":

		if len(fExpr.Args) != 1 {
			return "", fmt.Errorf("function '%s' must have one argument but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		regexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		// Use a non capturing group for performance
		out += "(?:" + regexString + ")*"

	case "one_plus_of":

		if len(fExpr.Args) != 1 {
			return "", fmt.Errorf("function '%s' must have one argument but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		regexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		out += "(?:" + regexString + ")+"

	case "from_to":

		if len(fExpr.Args) != 2 {
			return "", fmt.Errorf("function '%s' must have two arguments but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		firstParamRegexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		secondParamRegexString, err := gb.nodeToGoRegex(fExpr.Args[1])
		if err != nil {
			return "", err
		}

		out += firstParamRegexString + "-" + secondParamRegexString

	case "count_between":

		if len(fExpr.Args) != 3 {
			return "", fmt.Errorf("function '%s' must have three arguments but was passed %d arguments", fExpr.Ident.Name, len(fExpr.Args))
		}

		firstParamRegexString, err := gb.nodeToGoRegex(fExpr.Args[0])
		if err != nil {
			return "", err
		}

		secondParamRegexString, err := gb.nodeToGoRegex(fExpr.Args[1])
		if err != nil {
			return "", err
		}

		thirdParamRegexString, err := gb.nodeToGoRegex(fExpr.Args[2])
		if err != nil {
			return "", err
		}

		out += firstParamRegexString + "{" + secondParamRegexString + "," + thirdParamRegexString + "}"

	default:
		return "", fmt.Errorf("trying to call unknown function '%s'", fExpr.Ident.Name)
	}

	return out, err
}

func (gb *GoBackend) stringToBool(str string) (bool, error) {

	if str == "true" {
		return true, nil
	}

	if str == "false" {
		return false, nil
	}

	return false, fmt.Errorf("value '%v' is not a valid boolean value, only 'true' and 'false' are allowed", str)
}

func (gb *GoBackend) ApplyOptionsToRegexString(regexString string) string {

	flagsString := "(?"
	if !gb.Opts.CaseSensitive {
		flagsString += "i"
	}

	// In Go regex, 'g' flag doesn't exist, rather finding one or many is controlled by the regex.Regexp function used.
	// For example, for Go regex '(?i)case', Regexp.FindString("casecase") returns 'case',
	// while Regexp.FindAllString("casecase") returns ["case", "case"].
	//
	// if gb.Opts.FindAllMatches {
	// 	flagsString += "g"
	// }

	return flagsString + ")" + regexString
}

func (gb *GoBackend) escapeString(original string) string {

	sb := strings.Builder{}

	for _, r := range original {

		if r == '.' || r == '(' || r == ')' || r == '[' || r == ']' || r == '{' || r == '}' || r == '\\' {
			sb.WriteRune('\\')
		}

		sb.WriteRune(r)
	}

	return sb.String()
}
