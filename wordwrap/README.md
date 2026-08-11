# wordwrap — wraps text to a fixed width, modeled on the npm "word-wrap" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/wordwrap.svg)](https://pkg.go.dev/github.com/malcolmston/express/wordwrap)

Package wordwrap wraps text to a fixed width, modeled on the npm "word-wrap"
package. It is a standard-library-only Go port that reflows a paragraph of
text so that no output line exceeds a chosen width, which is useful for
rendering help text, log messages, email bodies, terminal output, and any
other place where a long run of prose must fit inside a bounded column.

Wrapping is greedy and whitespace-based. The input is split into words on runs
of whitespace, and words are packed onto the current line one at a time until
adding the next word (plus the single space that would join it) would push the
line past Width; at that point the current line is emitted and the word starts
a new line. Width is measured in runes and counts only the text, excluding the
per-line Indent, so multibyte and accented characters count as one column each
and the indent does not eat into the usable width.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/wordwrap"
```

## Usage

This is the package's own `ExampleWrap`, so it compiles and its output is
asserted on every `go test ./wordwrap/`.

```go
text := "The quick brown fox jumps over the lazy dog"
	fmt.Println(wordwrap.Wrap(text, wordwrap.Options{Width: 20}))
```

```
The quick brown fox
jumps over the lazy
dog
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Wrap(text string, opts Options) string` | Wrap wraps text according to opts and returns the wrapped result. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures Wrap. |

<details>
<summary><code>Options</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func NewOptions() Options` | NewOptions returns an Options populated with the package defaults: Width 50, Indent two spaces, and Newline "\n". |

</details>

### Constants

`DefaultWidth`, `DefaultIndent`, `DefaultNewline`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/wordwrap).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
