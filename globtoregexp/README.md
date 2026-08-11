# globtoregexp — converts glob patterns into regular expressions

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/globtoregexp.svg)](https://pkg.go.dev/github.com/malcolmston/express/globtoregexp)

Package globtoregexp converts glob patterns into regular expressions. It is a
faithful port of the npm package glob-to-regexp, translating shell-style glob
syntax into a compiled *regexp.Regexp that can be used for matching file
paths, keys, or any other strings.

Reach for this package when you have glob-style patterns (from configuration,
user input, or a manifest) and want to match strings against them using Go's
standard regexp engine. Rather than implementing a bespoke matcher, you
convert the glob once and reuse the resulting *regexp.Regexp. GlobToRegExp
uses the library's default options; GlobToRegExpOpts accepts an Options value
for finer control over extended syntax, globstar behavior, and RegExp flags.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/globtoregexp"
```

## Usage

This is the package's own `ExampleGlobToRegExp`, so it compiles and its output is
asserted on every `go test ./globtoregexp/`.

```go
re, err := globtoregexp.GlobToRegExp("*.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(re.String())
	fmt.Println(re.MatchString("foo.js"))
	fmt.Println(re.MatchString("foo/bar.js"))
```

```
^.*\.js$
true
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func GlobToRegExp(glob string) (*regexp.Regexp, error)` | GlobToRegExp converts a glob pattern to a *regexp.Regexp using the default options (Extended off, GlobStar off), matching the default behavior of the… |
| `func GlobToRegExpOpts(glob string, opts Options) (*regexp.Regexp, error)` | GlobToRegExpOpts converts a glob pattern to a *regexp.Regexp using the provided options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures how a glob pattern is converted to a regular expression. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/globtoregexp).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
