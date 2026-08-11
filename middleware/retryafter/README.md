# retryafter — provides middleware that automatically attaches a Retry-After header to responses whose

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/retryafter.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/retryafter)

Package retryafter provides middleware that automatically attaches a
Retry-After header to responses whose status indicates the client should back
off (429 Too Many Requests and 503 Service Unavailable by default). It has no
single Node counterpart; instead it packages, as a reusable express.Handler,
the Retry-After hinting that libraries such as express-rate-limit and
express's own maintenance/503 handlers emit ad hoc. The header tells
well-behaved clients and proxies how many seconds to wait before retrying, per
RFC 9110 section 10.2.3.

Reach for this middleware when your app returns throttling or maintenance
responses and you want a consistent, machine-readable back-off hint without
repeating res.Set("Retry-After", ...) at every call site. Mount it once, high
in the chain, and any downstream handler or rate limiter that produces a 429
or 503 automatically gains the header. For the occasional handler that emits
such a status directly and wants a bespoke value, the package also exposes the
Set helper, which writes the header immediately without the middleware.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/retryafter"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./retryafter/`.

```go
app := express.New()
	app.Use(retryafter.New(retryafter.Options{Seconds: 30}))

	app.Get("/down", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(503).Send("maintenance")
	})
	app.Get("/ok", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("healthy")
	})

	call := func(path string) {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		fmt.Printf("%s -> %d Retry-After=%q\n", path, w.Code, w.Header().Get("Retry-After"))
	}

	call("/down")
	call("/ok")
```

```
/down -> 503 Retry-After="30"
/ok -> 200 Retry-After=""
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns retry-after middleware configured by opts. |
| `func Set(res *express.Response, seconds int)` | Set writes a Retry-After header with the given number of seconds directly on res. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the retry-after middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/retryafter).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
