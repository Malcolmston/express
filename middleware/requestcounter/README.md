# requestcounter — provides middleware that maintains a per-process sequential counter of the total number

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/requestcounter.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestcounter)

Package requestcounter provides middleware that maintains a per-process
sequential counter of the total number of requests handled. It is the express
analogue of the lightweight in-memory request tallies people wire up in Node
with a closure over a counter variable (the pattern behind minimal hit
counters and the request-count portion of ad-hoc metrics middleware), packaged
here as a constructor that returns both the middleware and a thread-safe
accessor for the current total.

Use it when you want a cheap, dependency-free measure of load — a health or
debug endpoint that reports how many requests the process has served, a smoke
test that asserts traffic reached the app, or a coarse building block for
homegrown metrics. It is intentionally minimal: for labelled counters,
histograms, or Prometheus-style export you should reach for the metrics
middleware instead. The counter lives only in memory and is per-process, so it
resets to zero on restart and is not shared across instances.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/requestcounter"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./requestcounter/`.

```go
handler, count := requestcounter.New()

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	fmt.Printf("handled %d requests\n", count())
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() (express.Handler, func() int64)` | New returns request-counting middleware together with an accessor that reports the number of requests observed so far. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestcounter).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
