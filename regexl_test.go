package regexl

import (
	"testing"
)

func TestMain(t *testing.T) {

	testCases := []struct {
		desc      string
		isVerbose bool
		rl        Regexl
	}{
		{
			desc: "Simplest query",
			rl: Regexl{
				Query: `
				set_options({
					global_search: false,
				})
				for 'Hello there, friend! This is Omar'
				select 'friend'		
				`,
			},
		},
		{
			desc: "One func query",
			rl: Regexl{
				Query: `
				set_options({
					global_search: true,
				})
				for 'Hello there, friend! This is Omar'
				-- We can accept any number of inputs here!
				select any_strings_of('is', 'Omar')
				`,
			},
		},
		{
			desc:      "Multiple object params",
			isVerbose: true,
			rl: Regexl{
				Query: `
				set_options({
					global_search: true,
					case_sensitive: false,
				})
				for 'Hello there, friend! This is Omar'
				select any_chars_of('is', 'omar') -- Comments work here too
				`,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.desc, func(t *testing.T) {

			IsVerbose = tc.isVerbose
			err := tc.rl.Compile()
			if err != nil {
				t.Fatalf("Compilation failed. Err=%v; Query=%s\n", err, tc.rl.Query)
			}
		})
	}
}
