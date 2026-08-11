# cachecontrol — provides express middleware that sets a Cache-Control header assembled from a set

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/cachecontrol.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/cachecontrol)

Package cachecontrol provides express middleware that sets a Cache-Control
header assembled from a set of caching options. It plays the role of Node's
nocache middleware and the cache-control helpers built into serve-static and
express.static (their maxAge/cacheControl options), but generalized to any
route so that a single mount can express both "cache this aggressively" and
"never store this" policies as a reusable express.Handler.

Use this middleware to centralize caching policy instead of scattering
res.Set("Cache-Control", ...) calls across handlers. Mount it globally with
app.Use to give a whole application a default policy, or attach it to a
specific router or path prefix — long max-age with "public" for immutable
static assets, "no-store" for authenticated or sensitive JSON endpoints, and
"private" for per-user responses that shared caches must not retain. Because
the header value is computed once when the middleware is built, per-request
overhead is a single header write.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/cachecontrol"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./cachecontrol/`.

```go
app := express.New()
	app.Use(cachecontrol.New(cachecontrol.Options{
		Public: true,
		MaxAge: 3600,
	}))
	app.Get("/asset.js", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("// asset")
	})

	r := httptest.NewRequest("GET", "/asset.js", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Header().Get("Cache-Control"))
```

```
public, max-age=3600
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Cache-Control response header built from the provided options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the Cache-Control directive. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/cachecontrol).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
