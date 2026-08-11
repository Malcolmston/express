# timeout — provides middleware that bounds the time a downstream handler chain may take before

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/timeout.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/timeout)

Package timeout provides middleware that bounds the time a downstream handler
chain may take before the client is answered. It is the express framework's Go
analogue of Connect's connect-timeout middleware and the classic
express-timeout wrappers: a drop-in express.Handler that runs the rest of the
chain under a deadline and, if that deadline elapses first, answers the client
with 503 Service Unavailable instead of leaving the request hanging.

Reach for this middleware when a slow or stuck handler must not be allowed to
tie up a client indefinitely: routes that call flaky upstream services,
database queries without their own timeout, report generation, or any place
where a bounded worst-case latency is more valuable than always waiting for
the "real" answer. Placing it near the front of the chain -- before the
handlers you want to bound -- makes every handler downstream of it subject to
the same Duration budget.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/timeout"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./timeout/`.

```go
app := express.New()
	app.Use(timeout.New(timeout.Options{
		Duration: 20 * time.Millisecond,
		Message:  "upstream too slow",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(80 * time.Millisecond)
		res.Send("late answer")
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	time.Sleep(90 * time.Millisecond)
	fmt.Printf("%d %s\n", w.Code, w.Body.String())
```

```
503 upstream too slow
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns timeout middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the timeout middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/timeout).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
