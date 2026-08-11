# nocache — provides express middleware that instructs clients and proxies never to cache

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/nocache.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/nocache)

Package nocache provides express middleware that instructs clients and proxies
never to cache the response. It is the express port of the Node nocache
package (the same helper popularized by helmet's legacy noCache module and by
lusca), which sets the classic trio of anti-caching headers that HTTP/1.0
proxies, HTTP/1.1 caches, and older browsers each pay attention to.

Reach for it on any endpoint whose response must always be revalidated:
authenticated pages, dynamic dashboards, CSRF-token or nonce carrying HTML,
account and billing views, and API responses that reflect frequently changing
or user-specific state. Serving such content with default caching can leak one
user's data to another through a shared proxy or let a stale, sensitive page
linger in the browser's back/forward cache, so making "do not store this"
explicit is both a correctness and a security measure.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/nocache"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./nocache/`.

```go
app := express.New()
	app.Use(nocache.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("dynamic")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Cache-Control"))
	fmt.Println(rec.Header().Get("Pragma"))
	fmt.Println(rec.Header().Get("Expires"))
```

```
no-store, no-cache, must-revalidate
no-cache
0
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that sets the standard set of no-cache headers: Cache-Control, Pragma, and Expires. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/nocache).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
