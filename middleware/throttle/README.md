# throttle — provides a per-client token-bucket rate limiter for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/throttle.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/throttle)

Package throttle provides a per-client token-bucket rate limiter for the
express framework. It is the Go analogue of Node rate-limiting middleware such
as express-rate-limit, express-throttle, or the token-bucket limiters built on
top of the "limiter" npm module, packaged as a drop-in express.Handler. Each
client (keyed by IP by default) is given its own bucket that refills at a
fixed Rate up to a Burst capacity; each request consumes one token, and when
no token is available the request is rejected with 429 Too Many Requests plus
a Retry-After header.

Reach for this middleware when you want to protect an endpoint from abuse or
accidental flooding: login and password-reset routes, expensive search or
report endpoints, public APIs that must enforce a fair-use quota, or any
handler whose cost you want to spread out over time. The token-bucket model
allows short bursts (up to Burst requests back-to-back) while still capping
the sustained request rate at Rate requests per second, which is friendlier to
legitimate clients than a hard fixed-window counter.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/throttle"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./throttle/`.

```go
clock := time.Unix(0, 0)
	app := express.New()
	app.Use(throttle.New(throttle.Options{
		Rate:  1,
		Burst: 2,
		Now:   func() time.Time { return clock },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.7:5000"
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("request %d -> %d retry-after=%q\n", i+1, w.Code, w.Header().Get("Retry-After"))
	}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns throttle middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the throttle middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/throttle).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
