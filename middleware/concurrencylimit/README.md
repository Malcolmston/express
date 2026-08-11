# concurrencylimit — provides express middleware that caps the number of requests processed concurrently

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/concurrencylimit.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/concurrencylimit)

Package concurrencylimit provides express middleware that caps the number of
requests processed concurrently and sheds load once that cap is reached. It is
the Go analogue of the Node load-shedding pattern used by middleware such as
express-queue (run with a zero-length queue) or the connection-limiting
behaviour of toobusy-js, packaged as a drop-in express.Handler. Rather than
smoothing traffic over time like a rate limiter, it bounds the instantaneous
in-flight request count so a burst cannot overwhelm downstream resources.

Use this middleware to protect a service whose expensive work (database pools,
external APIs, CPU-bound handlers) has a known safe concurrency ceiling. When
more requests arrive than the ceiling allows, failing them fast with a 503 is
often healthier than letting them queue unboundedly and drive latency and
memory upward. Mount it globally with app.Use to guard the whole application,
or attach it to a specific router or path prefix to fence off only the costly
endpoints while leaving cheap ones unthrottled.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/concurrencylimit"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./concurrencylimit/`.

```go
app := express.New()
	app.Use(concurrencylimit.New(concurrencylimit.Options{
		Max:     2,
		Message: "Busy, try again later",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rr.Code)
```

```
200
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns concurrency-limiting middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the concurrency limiter. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/concurrencylimit).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
