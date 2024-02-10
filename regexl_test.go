package regexl

import (
	"testing"
)

func TestMain(t *testing.T) {

	rl := Regexl{
		Query: `
		set_options({
			global_search: false,
		})
		for 'Hello there, friend! This is Omar'
		select 'friend'		
		`,
	}

	err := rl.Compile()
	if err != nil {
		t.Fatalf("Compilation failed. Err=%v\n", err)
	}
}
