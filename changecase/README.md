# changecase — standard-library-only Go port of the popular npm package change-case

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/changecase.svg)](https://pkg.go.dev/github.com/malcolmston/express/changecase)

Package changecase is a standard-library-only Go port of the popular npm
package change-case (https://www.npmjs.com/package/change-case), which
converts identifiers and phrases between the many casing conventions used in
code and configuration. It exposes the full change-case family — CamelCase,
PascalCase, SnakeCase, KebabCase (a.k.a. param case), ConstantCase, DotCase,
PathCase, HeaderCase (train case), CapitalCase, SentenceCase and NoCase —
built on a single word-splitting pass so every converter tokenises input the
same way.

Splitting follows change-case's rules: any run of separators (spaces, hyphens,
underscores, dots and slashes) divides words, a lowercase- or
digit-to-uppercase transition starts a new word (so "fooBar" splits into "foo"
and "Bar"), and an uppercase run followed by a lowercase letter is treated as
an acronym boundary (so "HTTPServer" splits into "HTTP" and "Server"). Digits
attach to the preceding letters. The exported Words helper exposes this
tokenisation directly.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/changecase"
```

## Usage

This is the package's own `ExampleCamelCase`, so it compiles and its output is
asserted on every `go test ./changecase/`.

```go
const in = "Test String"
	fmt.Println(changecase.CamelCase(in))
	fmt.Println(changecase.PascalCase(in))
	fmt.Println(changecase.SnakeCase(in))
	fmt.Println(changecase.KebabCase(in))
	fmt.Println(changecase.ConstantCase(in))
	fmt.Println(changecase.DotCase(in))
	fmt.Println(changecase.PathCase(in))
	fmt.Println(changecase.HeaderCase(in))
	fmt.Println(changecase.CapitalCase(in))
	fmt.Println(changecase.SentenceCase(in))
	fmt.Println(changecase.NoCase(in))
```

```
testString
TestString
test_string
test-string
TEST_STRING
test.string
test/string
Test-String
Test String
Test string
test string
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CamelCase(s string) string` | CamelCase converts s to camelCase (first word lowercase, subsequent words title-cased and joined without a delimiter), e.g. |
| `func CapitalCase(s string) string` | CapitalCase converts s to Capital Case (title-cased words joined by spaces). |
| `func ConstantCase(s string) string` | ConstantCase converts s to CONSTANT_CASE (uppercase words joined by underscores). |
| `func DotCase(s string) string` | DotCase converts s to dot.case (lowercase words joined by dots). |
| `func HeaderCase(s string) string` | HeaderCase converts s to Header-Case (title-cased words joined by hyphens), also known as train case. |
| `func KebabCase(s string) string` | KebabCase converts s to kebab-case (lowercase words joined by hyphens). |
| `func NoCase(s string) string` | NoCase joins the lowercased words of s with single spaces, e.g. |
| `func ParamCase(s string) string` | ParamCase is an alias for KebabCase, matching change-case's paramCase name. |
| `func PascalCase(s string) string` | PascalCase converts s to PascalCase (every word title-cased, no delimiter), e.g. |
| `func PathCase(s string) string` | PathCase converts s to path/case (lowercase words joined by forward slashes). |
| `func SentenceCase(s string) string` | SentenceCase converts s to Sentence case (only the first word capitalised, the rest lowercase, joined by spaces). |
| `func SnakeCase(s string) string` | SnakeCase converts s to snake_case (lowercase words joined by underscores). |
| `func SwapCase(s string) string` | SwapCase returns s with the case of every letter inverted (upper becomes lower and vice versa), leaving non-letters untouched. |
| `func Words(s string) []string` | Words splits s into its constituent word tokens using change-case's rules: separator runs, camelCase humps and acronym boundaries all begin new… |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/changecase).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
