# urljoin — joins URL path segments together, normalizing slashes

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/urljoin.svg)](https://pkg.go.dev/github.com/malcolmston/express/urljoin)

Package urljoin joins URL path segments together, normalizing slashes. It is a
standard-library-only Go port of the npm "url-join" package, the small utility
countless Node projects use to assemble a URL from a base and one or more path
fragments without ending up with doubled or missing slashes. The problem it
solves is that naive string concatenation of URL parts produces artifacts like
"http://host//api" or "api/v1users" depending on which parts carried trailing
or leading slashes; url-join normalizes all of that away.

URLJoin takes a variadic list of parts and returns a single joined URL. Empty
parts are filtered out entirely, segments are joined with exactly one "/",
runs of duplicate slashes between segments are collapsed to a single slash,
leading slashes are stripped from every part after the first, and trailing
slashes are stripped from every part before the last. The last part keeps at
most a single trailing slash, so an intentional trailing "/" survives while
accidental doubles do not.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/urljoin"
```

## Usage

This is the package's own `ExampleURLJoin_query`, so it compiles and its output is
asserted on every `go test ./urljoin/`.

```go
fmt.Println(urljoin.URLJoin("http://example.com", "foo?a=1", "?b=2"))
```

```
http://example.com/foo?a=1&b=2
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func URLJoin(parts ...string) string` | URLJoin joins the given URL parts into a single URL, collapsing duplicate slashes while preserving the protocol separator and correctly combining… |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/urljoin).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
