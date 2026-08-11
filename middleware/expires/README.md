# expires — provides express middleware that sets an HTTP Expires header a fixed duration into

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/expires.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/expires)

Package expires provides express middleware that sets an HTTP Expires header a
fixed duration into the future. It fills the same role as the Node
"express-expires" style helpers and the Expires option of static-file
middleware such as serve-static: a relative freshness lifetime expressed as a
duration is turned into an absolute HTTP-date on every response.

Use this middleware to give responses a coarse cache lifetime for HTTP/1.0
caches and proxies that predate Cache-Control, or simply as a belt-and-braces
companion to Cache-Control: max-age. It is well suited to content that should
be considered fresh for a fixed window after it is served — semi-static
pages, infrequently changing JSON, downloadable assets. Mount it with app.Use
over the routes whose responses should carry an expiry.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/expires"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./expires/`.

```go
app := express.New()
	app.Use(expires.New(expires.Options{Duration: time.Hour}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	exp, err := http.ParseTime(rec.Header().Get("Expires"))
	fmt.Println(err == nil && exp.After(time.Now()))
```

```
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Expires response header to the current time plus Duration, formatted as an HTTP date (RFC 1123 GMT). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the expires middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/expires).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
