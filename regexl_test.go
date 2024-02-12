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
			desc: "Simplest",
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
			desc: "One func",
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
			desc: "Multiple object params",
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
		{
			desc: "Nested funcs",
			rl: Regexl{
				Query: `
				for 'Golang'
				select ends_with(starts_with('Golang'))
				-- select starts_and_ends_with('Golang') -- Alternative way of writing it
				`,
			},
		},
		{
			desc: "Combined funcs",
			rl: Regexl{
				Query: `
				set_options({
					global_search: true,
				})
				for 'Hello there, friend! This is Omar'
				select starts_with('Hello') + any_chars() + 'Omar'
				`,
			},
		},
		{
			desc:      "Email query",
			isVerbose: true,
			rl: Regexl{
				Query: `
				set_options({
					case_sensitive: false,
				})
				for 'some-email@wow.com'
				select
					-- Converts to: [A-Z0-9._%+-]+
					one_plus_of(
						any_chars_of(from_to('A', 'Z'), from_to(0, 9), '._%+-')
					) +
					-- Converts to: @
					'@' +
					-- Converts to: [A-Z0-9.-]+
					one_plus_of(
						any_chars_of(from_to('A','Z'), from_to(0, 9), '.-')
					) +
					-- Converts to: \.
					'.' +
					-- Converts to: [A-Z]{2,10}
					char_count_between(
						any_chars_of(from_to('A', 'Z')),
						2,
						10
					)
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
