# Regexl

Regexl is a high level language for regex that can be used in any project as a simple library.

You can read about the reasoning for creating Regexl [here](https://bloeys.com/thoughts/thought-2-regex-is-like-assembly/).

**Table of contents:**

- [Regexl](#regexl)
  - [Regexl Query Examples](#regexl-query-examples)
  - [Usage in Go](#usage-in-go)

## Regexl Query Examples

- `/friend/i` is equivalent to the regexl:

``` sql
select 'friend'
```

- `/^friend/i` is equivalent to the regexl:

``` sql
// This is a regexl comment.
// This set_options configuration is equivalent to: '/i'
set_options({
    case_sensitive: false,
})

select starts_with('friend')
```

- `/Hello*/g` is equivalent to the regexl:

``` sql
set_options({
    find_all_matches: true,
})

--// This '--' is to help the syntax highlighter :)
--// The '+' performs a simple concatenation, as all functions return strings
select 'Hell' + zero_plus_of('o')
```

- `/^Golang$/` is equivalent to the regexl:

``` sql
set_options({
    case_sensitive: false,
})
--// Functions can be nested, as outputs are strings.
--// Alternative regexl: select starts_and_ends_with('Golang')
select ends_with(starts_with('Golang'))
```

- `/[abcd]/ig` (match any of these 4 letters) is equivalent to the regexl:

``` sql
set_options({
    find_all_matches: true,
    case_sensitive: false,
})
--// Can also be: select any_chars_of('abcd')
select any_chars_of('abc', 'd')
```

- `/[A-Z0-9]/ig` (match letters and numbers only) is equivalent to the regexl:

``` sql
set_options({
    find_all_matches: true,
    case_sensitive: false,
})
--// Can also be: select any_chars_of('abcd')
select any_chars_of(from_to('A', 'Z'), from_to(0, 9))
```

- `/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,10}/i` (a 'simple' email regex) is equivalent to the regexl:

``` sql
set_options({
    case_sensitive: false,
})
select
    --// Converts to: [A-Z0-9._%+-]+
    one_plus_of(
        any_chars_of(from_to('A', 'Z'), from_to(0, 9), '._%+-')
    ) +
    --// Converts to: @
    '@' +
    --// Converts to: [A-Z0-9.-]+
    one_plus_of(
        any_chars_of(from_to('A', 'Z'), from_to(0, 9), '.-')
    ) +
    --// Converts to: \.
    '.' +
    --// Converts to: [A-Z]{2,10}
    count_between(
        any_chars_of(from_to('A', 'Z')),
        2,
        10
    )
```

## Usage in Go

```go
package main

import (
	"fmt"

	"github.com/bloeys/regexl"
)

func main() {

	regexlQuery := `
		set_options({
			find_all_matches: true,
			case_sensitive: false,
		})

		select starts_with('Hello there, ') + one_plus_of(any_chars_of(from_to('A', 'Z'), '.!-'))
	`

	rl := regexl.NewRegexl(regexlQuery)
	hasMatch := rl.MustCompile().CompiledRegexp.MatchString("Hello there, friend!")

	fmt.Printf("Produced regex: %s\nHas match: %v\n", rl.CompiledRegexp.String(), hasMatch)
}
```
