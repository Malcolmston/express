# parseurl — parses the URL of an incoming HTTP request into its components

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/parseurl.svg)](https://pkg.go.dev/github.com/malcolmston/express/parseurl)

Package parseurl parses the URL of an incoming HTTP request into its
components, a stdlib-only Go port of the npm package "parseurl"
(https://www.npmjs.com/package/parseurl). In the Node/Express world parseurl
is the tiny utility that turns a request's raw target string (req.url) into a
structured URL object with fields such as path, query and pathname, and it
underpins routing, static file serving and any middleware that needs to reason
about the request path independently of the query string.

The reason parseurl exists at all in Node is performance and correctness of
caching. Parsing a URL is not free, and a single request typically has its URL
inspected by many layers of middleware, so the original library memoizes the
parsed result on the request object and reparses only when the underlying
req.url string has changed. It also distinguishes the current URL (which
Express may rewrite as it strips mounted route prefixes) from the original,
unmodified request target, exposing the latter through parseurl.original.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/parseurl"
```

## Usage

This is the package's own `ExampleParse`, so it compiles and its output is
asserted on every `go test ./parseurl/`.

```go
req := httptest.NewRequest("GET", "/foo/bar?baz=1&qux=2", nil)

	u := parseurl.Parse(req)
	fmt.Println("path:", u.Path)
	fmt.Println("query:", u.RawQuery)
	fmt.Println("baz:", u.Query().Get("baz"))
```

```
path: /foo/bar
query: baz=1&qux=2
baz: 1
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func OriginalURL(req *http.Request) *url.URL` | OriginalURL parses the request's original URL. |
| `func Parse(req *http.Request) *url.URL` | Parse parses the request's effective URL, using req.RequestURI and falling back to req.URL when RequestURI is empty or cannot be parsed. |
| `func ParseString(rawurl string) (*url.URL, error)` | ParseString parses a raw URL string into a *url.URL. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/parseurl).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
