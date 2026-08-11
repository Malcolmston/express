# cookie — parses and serializes HTTP cookie headers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/cookie.svg)](https://pkg.go.dev/github.com/malcolmston/express/cookie)

Package cookie parses and serializes HTTP cookie headers, a port of the npm
"cookie" package that Express and cookie-parser use to read the Cookie request
header and to build Set-Cookie response headers. It exposes Parse to turn a
raw "Cookie" header into a map of names to values and Serialize to render a
single name/value pair plus its attributes into a Set-Cookie value, using only
the Go standard library.

You reach for this package on both sides of the request/response cycle. On the
way in, Parse gives you the individual cookies a client sent so a handler can
look up a session id or a preference flag without string-splitting the header
itself. On the way out, Serialize produces a well-formed Set-Cookie value with
Path, Domain, Expires, Max-Age, Secure, HttpOnly, and SameSite attributes, so
you can issue login cookies, clear a cookie by expiring it, or scope a cookie
to a path without hand-assembling the syntax and risking a malformed header.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/cookie"
```

## Usage

This is the package's own `ExampleSerialize`, so it compiles and its output is
asserted on every `go test ./cookie/`.

```go
s, err := cookie.Serialize("sid", "a b+c", &cookie.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
```

```
sid=a%20b%2Bc; Path=/; Max-Age=3600; HttpOnly
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Parse(header string) map[string]string` | Parse parses a Cookie header (for example "a=1; b=2") into a map of names to values. |
| `func Serialize(name, value string, opts *Options) (string, error)` | Serialize builds a Set-Cookie header value for the given name and value, applying the attributes in opts (which may be nil). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options holds the attributes used when serializing a cookie. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/cookie).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
