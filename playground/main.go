//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/bloeys/regexl"
)

func main() {

	done := make(chan struct{}, 0)
	global := js.Global()
	global.Set("regexlCompileAndMatch", js.FuncOf(regexlCompileAndMatch))
	<-done
}

type RegexlCompileAndMatchOutput struct {
	RegexString string
	ErrString   string
	HasMatch    bool
}

func regexlCompileAndMatch(this js.Value, args []js.Value) (outputString any) {

	// Always output string with helper function
	output := RegexlCompileAndMatchOutput{}

	updateOutputString := func() {

		b, err := json.Marshal(output)
		if err != nil {
			outputString = err.Error()
			return
		}

		outputString = string(b)
	}

	// Panic recovery
	defer func() {
		if err := recover(); err != nil {
			output.ErrString = fmt.Sprint(err)
			updateOutputString()
		}
	}()

	// Validate args
	if len(args) != 2 {
		output.ErrString = "Must pass 2 arguments to regexlCompileAndMatch, first is the query and second is the string to match the regex against"
		updateOutputString()
		return
	}

	regexlQuery := args[0].String()
	textToMatchAgainst := args[1].String()

	rl := regexl.NewRegexl(regexlQuery)

	err := rl.Compile()
	if err != nil {
		output.ErrString = err.Error()
		updateOutputString()
		return
	}

	hasMatch := rl.CompiledRegexp.MatchString(textToMatchAgainst)

	output.HasMatch = hasMatch
	output.RegexString = rl.CompiledRegexp.String()

	updateOutputString()
	return
}
