# rangeparser — parses the HTTP Range header, a port of the npm "range-parser" package used by Express

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/rangeparser.svg)](https://pkg.go.dev/github.com/malcolmston/express/rangeparser)

Package rangeparser parses the HTTP Range header, a port of the npm
"range-parser" package used by Express for req.range and by static file
servers to implement partial-content (HTTP 206) responses. It resolves a Range
header such as "bytes=0-499" against a total resource size and reports the
concrete byte ranges requested.

The Range header lets a client ask for one or more sub-ranges of a resource
instead of the whole thing, which is what makes resumable downloads and media
seeking possible. A header has the form "unit=spec,spec,..." where unit is
typically "bytes" and each spec is either "start-end", "start-" (from start to
the end of the resource), or "-suffix" (the final suffix bytes). Parse
resolves each spec against the given size, clamping ends that run past the
last byte and translating suffix ranges into absolute [Start, End] offsets.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/rangeparser"
```

## Usage

This is the package's own `ExampleParse_combine`, so it compiles and its output is
asserted on every `go test ./rangeparser/`.

```go
ranges, code := rangeparser.Parse(1000, "bytes=0-100,50-200,300-400", true)
	fmt.Println(ranges, code)
```

```
[{0 200} {300 400}] 0
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Range` | Range represents a single inclusive byte range [Start, End]. |
| `Ranges` | Ranges holds the parsed ranges together with the unit ("type") that was requested, e.g. |

<details>
<summary><code>Range</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(size int64, header string, combine bool) (ranges []Range, result int)` | Parse parses the given Range header against a resource of the given size. |

</details>

<details>
<summary><code>Ranges</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func ParseRanges(size int64, header string, combine bool) (Ranges, int)` | ParseRanges is like Parse but also returns the requested unit via Ranges.Type. |

</details>

### Constants

`ResultOK`, `ResultUnsatisfiable`, `ResultMalformed`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/rangeparser).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
