# encodeurl — encodes a URL to a percent-encoded form while leaving any already-encoded sequences

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/encodeurl.svg)](https://pkg.go.dev/github.com/malcolmston/express/encodeurl)

Package encodeurl encodes a URL to a percent-encoded form while leaving any
already-encoded sequences intact, a port of the npm "encodeurl" package that
Express and the send/serve-static middleware use when reflecting a request
path back into a header such as Location. It exposes the single Encode
function and is built on only the Go standard library.

The problem it solves is double-encoding. When a server takes a URL that may
already contain percent-escapes — most commonly the request URL itself —
and needs to place it somewhere that requires a valid URL (a redirect target,
an error message, a link), naively percent-encoding it would turn an existing
"%20" into "%2520". Encode instead encodes only the characters that are unsafe
or invalid in a URL and passes through sequences that already look like valid
escapes, so the output is safe to emit exactly once without corrupting escapes
the caller had already applied.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/encodeurl"
```

## Usage

This is the package's own `ExampleEncode`, so it compiles and its output is
asserted on every `go test ./encodeurl/`.

```go
fmt.Println(encodeurl.Encode("/foo/bar?msg=hello world&x=1%202"))
```

```
/foo/bar?msg=hello%20world&x=1%202
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Encode(rawurl string) string` | Encode encodes a URL leaving already percent-encoded sequences intact. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/encodeurl).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
