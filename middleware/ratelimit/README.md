# ratelimit — provides a fixed-window, per-client rate limiter for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/ratelimit.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/ratelimit)

Package ratelimit provides a fixed-window, per-client rate limiter for the
express framework. Each client (keyed by IP by default) is allowed a maximum
number of requests within a fixed window; requests beyond the limit receive a
429 Too Many Requests response carrying standard rate-limit headers.

It ports the fixed-window strategy popularized by the Node express-rate-limit
middleware: a counter per key that resets at a fixed boundary, together with
the conventional X-RateLimit-Limit, X-RateLimit-Remaining, and
X-RateLimit-Reset response headers (and a Retry-After header on rejection).
Like the classic express-rate-limit default store, counters are held in an
in-process map, so the limiter is self-contained and requires no external
datastore.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/ratelimit"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./ratelimit/`.

```go
clock := time.Unix(1000, 0)
	app := express.New()
	app.Use(ratelimit.New(ratelimit.Options{
		Max:    1,
		Window: time.Minute,
		Now:    func() time.Time { return clock },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	do := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "1.2.3.4:5555"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println(do())
	fmt.Println(do())
```

```
200
429
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns rate-limiting middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the rate limiter. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/ratelimit).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
