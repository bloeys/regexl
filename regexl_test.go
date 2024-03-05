package regexl

import (
	"testing"
)

// @TODO: Add 1+ matching strings for each positive case to be tested with Regexp.MatchString

func TestMain(t *testing.T) {

	testCases := []struct {
		desc          string
		rl            Regexl
		expectedRegex string
		shouldError   bool
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
			expectedRegex: "(?i)friend",
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
			expectedRegex: "(?i)is|Omar",
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
			expectedRegex: "(?i)[isomar]",
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
			expectedRegex: "(?i)^friend",
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
			expectedRegex: "(?i)omar$",
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
			expectedRegex: "(?i)Hell(?:o)*",
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
			expectedRegex: "(?i)Hell(?:o)+",
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
			expectedRegex: "(?)^Golang$",
		},
		{
			desc: "Combined funcs 1",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: true,
				})
				select starts_with('Hello') + any_chars() + 'Omar'
				`,
			},
			expectedRegex: "(?i)^Hello.*Omar",
		},
		{
			desc: "Combined funcs 2",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: true,
					case_sensitive: false,
				})
				select starts_with('Hello there, ') + one_plus_of(any_chars_of(from_to('A', 'Z'), '.!-'))
				`,
			},
			expectedRegex: "(?i)^Hello there, (?:[A-Z\\.!-])+",
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
					count_between(
						any_chars_of(from_to('A', 'Z')),
						2,
						10
					)
				`,
			},
			expectedRegex: "(?i)(?:[A-Z0-9\\._%+-])+@(?:[A-Z0-9\\.-])+\\.[A-Z]{2,10}",
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
			expectedRegex: "(?i)^Hello.*Omar",
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
			expectedRegex: "(?i)^Hello.*Omar",
		},
		{
			desc: "Crazy formatting 3 - one line",
			rl: Regexl{
				Query: `
				set_options({find_all_matches: true}) select starts_with('Hello') + any_chars() + 'Omar'				
			`,
			},
			expectedRegex: "(?i)^Hello.*Omar",
		},

		//
		// Negative test cases
		//
		{
			desc: "Invalid 1",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: false,
				})
				select 'friend
				`,
			},
			shouldError: true,
		},
		{
			desc: "Invalid 2",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches:,
				})
				select 'friend'
				`,
			},
			shouldError: true,
		},
		{
			desc: "Invalid 3",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: galse,
				})
				select 'friend'
				`,
			},
			shouldError: true,
		},
		{
			desc: "Invalid 4",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: false,
				})
				'friend'
				`,
			},
			shouldError: true,
		},
		{
			desc: "Invalid 5",
			rl: Regexl{
				Query: `
				set_options({
					find_all_matches: false,
				})
				select
				`,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {

		success := t.Run(tc.desc, func(t *testing.T) {

			err := tc.rl.Compile()
			if err != nil {

				if tc.shouldError {
					return
				}

				t.Errorf("Compilation failed. Err=%v; Query=%s\n", err, tc.rl.Query)
				return
			}

			if tc.shouldError {
				t.Errorf("Compilation should have thrown an error but didn't. Query=%s\n", tc.rl.Query)
				return
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
