# accesslog — provides middleware that writes an Apache "combined" style access log line for every

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/accesslog.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/accesslog)

Package accesslog provides middleware that writes an Apache "combined" style
access log line for every request once it completes. It is the Go analogue of
Node request loggers such as morgan (specifically its "combined" format),
exposed as a drop-in express.Handler that emits one line per request to a
configurable writer.

Use this middleware to obtain conventional web-server access logs from an
express application without pulling in an external logging framework. Because
it emits the widely understood Apache combined format, its output can be
consumed directly by log analyzers, shipped to aggregation pipelines, or
tailed during development. Mount it once at the top of the chain to log the
entire application, or attach it to a specific router to log only a subtree.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/accesslog"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./accesslog/`.

```go
var buf bytes.Buffer

	app := express.New()
	app.Use(accesslog.New(accesslog.Options{Writer: &buf}))
	app.Get("/hello", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(201).Send("hi there")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/hello", nil))

	if !strings.Contains(buf.String(), "GET /hello") || !strings.Contains(buf.String(), " 201 ") {
		panic("expected a combined-format log line")
	}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that writes an Apache combined-log-format line after each request, observing the final status code and number of bytes written. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the access log middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/accesslog).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
