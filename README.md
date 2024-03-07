# Regexl

Regexl is a high level language for regex that can be used in any project as a simple library.

You can read about the reasoning for creating Regexl [here](https://bloeys.com/thoughts/thought-2-regex-is-like-assembly/).

**Table of contents:**

- [Regexl](#regexl)
  - [Playground](#playground)
  - [Regexl Query Examples](#regexl-query-examples)
  - [Usage in Go](#usage-in-go)
  - [Technical Details](#technical-details)
  - [Todo](#todo)

## Playground

There is a (WASM based) playground where you can play with Regexl [here](https://regexl-playground.bloeys.com/).

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

//-- This '--' is to help the syntax highlighter :)
//-- The '+' performs a simple concatenation, as all functions return strings
select 'Hell' + zero_plus_of('o')
```

- `/^Golang$/` is equivalent to the regexl:

``` sql
set_options({
    case_sensitive: false,
})
//-- Functions can be nested, as outputs are strings.
//-- Alternative regexl: select starts_and_ends_with('Golang')
select ends_with(starts_with('Golang'))
```

- `/[abcd]/ig` (match any of these 4 letters) is equivalent to the regexl:

``` sql
set_options({
    find_all_matches: true,
    case_sensitive: false,
})
//-- Can also be: select any_chars_of('abcd')
select any_chars_of('abc', 'd')
```

- `/[A-Z0-9]/ig` (match letters and numbers only) is equivalent to the regexl:

``` sql
set_options({
    find_all_matches: true,
    case_sensitive: false,
})
//-- Can also be: select any_chars_of('abcd')
select any_chars_of(from_to('A', 'Z'), from_to(0, 9))
```

- `/[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,10}/i` (a 'simple' email regex) is equivalent to the regexl:

``` sql
set_options({
    case_sensitive: false,
})
select
    //-- Converts to: [A-Z0-9._%+-]+
    one_plus_of(
        any_chars_of(from_to('A', 'Z'), from_to(0, 9), '._%+-')
    ) +
    //-- Converts to: @
    '@' +
    //-- Converts to: [A-Z0-9.-]+
    one_plus_of(
        any_chars_of(from_to('A', 'Z'), from_to(0, 9), '.-')
    ) +
    //-- Converts to: \.
    '.' +
    //-- Converts to: [A-Z]{2,10}
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

## Technical Details

The Regexl code is that of a very simple compiler, where the general steps involved are:

1. Input query text is tokenized (implemented by `parser.go`)
2. Tokens are used to create an Abstract Syntax Tree (AST) (implemented by `ast.go`)
3. The AST is fed into a 'backend' that outputs a specific regex string (e.g. Go regex) (implemented by `regex_go_backend.go`)

To explain the above, lets look at how the following query is compiled:

```sql
select starts_with('hello')
```

By tokenization we mean turning the input string into higher level segments, where each segment is split by some separator like a space, a bracket, and so on.
In the above query you will get the following tokens:

- Token value: `select`; Type: `keyword`
- Token value: `starts_with`; Type: `function name`
- Token value: `(`; Type: `open bracket`
- Token value: `hello`; Type: `string`
- Token value: `)`; Type: `close bracket`

With this list of tokens, an AST is created. An Abstract Syntax Tree represents the structure of a program as a tree, where the parent nodes have a dependency on the children nodes.
For example, if function A calls B, then this function call node becomes a child of A, and the arguments of this call are children of the function call node.

In our query, the linear tokens list produces this AST tree:

```text
|-- select
|   |-- starts_with
|   |   |-- hello
```

With the AST in place, we can traverse the tree and generate some output.
In normal programming languages (e.g. C, Go, Python, etc...) the final output would be machine code, assembly, or perhaps byte code to be interpreted.

In Regexl, the output is some specific regex like Go-compatible regex, python-compatible regex, and so on (regex syntax and features differ between implementations).

The Go regex produced for our example Regexl query is:

```text
(?i)^hello
```

Equivalent to the more common regex expression:

```text
/^hello/i
```

The nice thing about this setup is that to support a new regex implementation all one has to do is implement a new backend (step 3), while tokenization and AST generation are reused as-is.

## Todo

- Become feature complete with Go regex
- Better error messages
- More test cases
