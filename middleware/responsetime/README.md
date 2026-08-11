# responsetime — provides express middleware that measures how long a request takes to process

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/responsetime.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/responsetime)

Package responsetime provides express middleware that measures how long a
request takes to process and reports it in an X-Response-Time header. It is
the express framework's Go port of the popular Connect/Express response-time
middleware (the npm "response-time" package), reproducing its default behavior
of stamping each response with the elapsed handling time in milliseconds.

Reach for this middleware to get cheap, always-on latency visibility during
development and in production: the header lets you spot slow endpoints from a
browser's network panel, a curl -I, or a log of response headers without
wiring up a metrics backend. It is a diagnostic aid rather than a full
observability solution — for percentiles, histograms, or per-route timing
you would still export to a real metrics system — but for a quick read on
how long a request spent inside the app it is effectively free.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/responsetime"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./responsetime/`.

```go
app := express.New()
	app.Use(responsetime.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))

	got := w.Header().Get(responsetime.HeaderName)
	fmt.Printf("header set: %t\n", got != "")
	fmt.Printf("has ms suffix: %t\n", strings.HasSuffix(got, "ms"))
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that records the time spent handling a request and sets X-Response-Time (in milliseconds, e.g. |

### Constants

`HeaderName`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/responsetime).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
