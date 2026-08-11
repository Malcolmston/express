# latencyheader — provides middleware that measures how long a request takes to process and records

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/latencyheader.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/latencyheader)

Package latencyheader provides middleware that measures how long a request
takes to process and records the elapsed milliseconds in a response header
(X-Response-Latency by default). It is the Go analogue of the small
"response-time" style middleware common in the Node/Express ecosystem (for
example the response-time npm package used with connect/express), exposing the
server-side processing time to clients, proxies, and browser developer tools
without any external dependencies.

Use this middleware when you want lightweight, always-on latency visibility
for every response: profiling slow endpoints in development, feeding synthetic
monitoring that scrapes the header, or letting a downstream load balancer or
CDN observe per-request processing cost. Because the value is computed from a
monotonic wall-clock delta and written as a plain integer, it is cheap enough
to leave enabled in production, and the injectable clock makes the timing
reproducible in tests.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/latencyheader"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./latencyheader/`.

```go
cur := time.Unix(0, 0)
	app := express.New()
	app.Use(latencyheader.New(latencyheader.Options{
		Now: func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cur = cur.Add(25 * time.Millisecond)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get("X-Response-Latency"))
```

```
25
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns latency-header middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the latency-header middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/latencyheader).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
