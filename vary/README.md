# vary — manipulates the HTTP Vary response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/vary.svg)](https://pkg.go.dev/github.com/malcolmston/express/vary)

Package vary manipulates the HTTP Vary response header. It is a
standard-library-only Go port of the npm "vary" package, the small helper
Express and its middleware use to record which request headers a response
varies on. The Vary header tells caches that a stored response may only be
reused for a later request when the listed request headers match, so getting
it right is essential for correct content negotiation: middleware that
branches on Accept-Encoding, Accept-Language, or Origin must append the
corresponding field to Vary or a shared cache may serve the wrong variant.

The core operation is appending a field name to an existing Vary value without
introducing duplicates. Because HTTP header field names are case-insensitive,
deduplication here is also case-insensitive: appending "accept-encoding" to a
header that already contains "Accept-Encoding" leaves the header unchanged
rather than listing the field twice. Existing fields are preserved in their
original order and casing, and newly appended fields are added in the order
given, so the result is stable and predictable.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/vary"
```

## Usage

This is the package's own `ExampleAppend`, so it compiles and its output is
asserted on every `go test ./vary/`.

```go
out, err := vary.Append("Accept-Encoding", "Accept-Language")
	fmt.Println(out, err)
```

```
Accept-Encoding, Accept-Language <nil>
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Append(header string, field ...string) (string, error)` | Append appends field(s) to the given Vary header value, returning the new value. |
| `func Field(fields []string) (string, error)` | Field builds a Vary header value from a list of field names, deduplicating case-insensitively and collapsing to "*" when appropriate. |
| `func Vary(h http.Header, field ...string)` | Vary appends the given field(s) to the Vary header of h, mutating it in place. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/vary).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
