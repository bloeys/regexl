package regexl

import (
	"testing"
)

func TestMain(t *testing.T) {

	testCases := []struct {
		desc          string
		rl            Regexl
		expectedRegex string
	}{
		{
			desc: "Simplest",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: false,
				})
				select 'friend'
				`,
			},
			expectedRegex: "/friend/i",
		},
		{
			desc: "One func",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: true,
				})
				// We can accept any number of inputs here!
				select any_strings_of('is', 'Omar')
				`,
			},
			expectedRegex: "/is|Omar/ig",
		},
		{
			desc: "Multiple object params",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: true,
					case_sensitive: false,
				})
				select any_chars_of('is', 'omar') // Comments work here too
				`,
			},
			expectedRegex: "/[isomar]/ig",
		},
		{
			desc: "Func: starts_with",
			rl: Regexl{
				Query: `
				// /^friend/i
				// Strings that can match:
				//   'Friend, how are you?'
				set_options({
					case_sensitive: false,
				})
				select starts_with('friend')
				`,
			},
			expectedRegex: "/^friend/i",
		},
		{
			desc: "Func: ends_with",
			rl: Regexl{
				Query: `
				// /omar$/i
				// Strings that can match:
				//   'Hello there, friend! This is Omar'
				set_options({
					case_sensitive: false,
				})
				select ends_with('omar')
				`,
			},
			expectedRegex: "/omar$/i",
		},
		{
			desc: "Func: zero_plus_of",
			rl: Regexl{
				Query: `
				// /Hell(o)*/g
				// Equivalent to: /Hello*/g
				// Strings that can match:
				//   'Hello there, friend!'
				//   'Hell there, friend!'
				//   'Hellooooo there, friend!'
				set_options({
					find_all_matches: true,
				})
				select 'Hell' + zero_plus_of('o')
				`,
			},
			expectedRegex: "/Hell(o)*/ig",
		},
		{
			desc: "Func: one_plus_of",
			rl: Regexl{
				Query: `
				// /Hell(o)+/g
				// Equivalent to: /Hello+/g
				// Strings that can match:
				//   'Hello there, friend!'
				// 'Helloooo' will match but not 'Hell'
				set_options({
					find_all_matches: true,
				})
				select 'Hell' + one_plus_of('o')
				`,
			},
			expectedRegex: "/Hell(o)+/ig",
		},
		{
			desc: "Nested funcs",
			rl: Regexl{
				Query: `
				set_options({
					case_sensitive: true,
				})
				// Alternative: select starts_and_ends_with('Golang')
				select ends_with(starts_with('Golang'))
				`,
			},
			expectedRegex: "/^Golang$/",
		},
		{
			desc: "Combined funcs",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: true,
				})
				select starts_with('Hello') + any_chars() + 'Omar'
				`,
			},
			expectedRegex: "/^Hello.*Omar/ig",
		},
		{
			desc: "Email query",
			rl: Regexl{
				Query: `
				set_options({
					case_sensitive: false,
				})
				select
					// Converts to: [A-Z0-9._%+-]+
					one_plus_of(
						any_chars_of(from_to('A', 'Z'), from_to(0, 9), '._%+-')
					) +
					// Converts to: @
					'@' +
					// Converts to: [A-Z0-9.-]+
					one_plus_of(
						any_chars_of(from_to('A','Z'), from_to(0, 9), '.-')
					) +
					// Converts to: \.
					'.' +
					// Converts to: [A-Z]{2,10}
					char_count_between(
						any_chars_of(from_to('A', 'Z')),
						2,
						10
					)
				`,
			},
			// /([A-Z0-9\._%+-])+@([A-Z0-9\.-])+\.[A-Z]{2,10}/i
			expectedRegex: "/([A-Z0-9\\._%+-])+@([A-Z0-9\\.-])+\\.[A-Z]{2,10}/i",
		},
		{
			desc: "Crazy formatting 1",
			rl: Regexl{
				Query: `
			set_options  (  {
				find_all_matches  : true  ,
			}	)
			select starts_with( 'Hello'  )        +any_chars (  )+ 'Omar'
			`,
			},
			expectedRegex: "/^Hello.*Omar/ig",
		},
		{
			desc: "Crazy formatting 2",
			rl: Regexl{
				Query: `
			set_options  (  
				
				{
				
					find_all_matches  : true}	
				)
			select starts_with( 'Hello'  )        +any_chars (  )+ 'Omar'
			`,
			},
			expectedRegex: "/^Hello.*Omar/ig",
		},
		{
			desc: "Crazy formatting 3 - one line",
			rl: Regexl{
				Query: `
				set_options({find_all_matches: true}) select starts_with('Hello') + any_chars() + 'Omar'				
			`,
			},
			expectedRegex: "/^Hello.*Omar/ig",
		},
	}

	for _, tc := range testCases {

		success := t.Run(tc.desc, func(t *testing.T) {

			err := tc.rl.Compile()
			if err != nil {
				t.Errorf("Compilation failed. Err=%v; Query=%s\n", err, tc.rl.Query)
			}

			if tc.expectedRegex != tc.rl.CompiledRegexp.String() {
				t.Errorf("Compiled regex does not equal expected regex. Expected=%s; Compiled=%s\n", tc.expectedRegex, tc.rl.CompiledRegexp.String())
			}
		})

		if !success {
			break
		}
	}
}
