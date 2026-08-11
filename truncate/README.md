# truncate — shortens a string to a maximum length, appending an ellipsis when content is cut

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/truncate.svg)](https://pkg.go.dev/github.com/malcolmston/express/truncate)

Package truncate shortens a string to a maximum length, appending an ellipsis
when content is cut. It is a Go substitute inspired by the npm truncate
package and the closely related lodash _.truncate helper, and is meant for
producing fixed-width previews such as list summaries, table cells, or meta
descriptions where an overly long value must be clipped.

Length is measured in runes rather than bytes, so multibyte characters count
as one unit and a truncated string is never split in the middle of a rune. The
length budget is inclusive of the ellipsis: the ellipsis length is subtracted
from the requested length before content is taken. This matches the lodash
convention where the omission string occupies part of the total length rather
than being added on top of it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/truncate"
```

## Usage

This is the package's own `ExampleTruncateOpts_wordBoundary`, so it compiles and its output is
asserted on every `go test ./truncate/`.

```go
out := truncate.TruncateOpts("The quick brown fox", 14, truncate.Options{
		Ellipsis:     "...",
		WordBoundary: true,
	})
	fmt.Println(out)
```

```
The quick...
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Truncate(s string, length int) string` | Truncate shortens s to at most length runes, appending the default ellipsis "..." when truncation occurs. |
| `func TruncateOpts(s string, length int, opts Options) string` | TruncateOpts shortens s to at most length runes using the provided options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures TruncateOpts. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/truncate).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
