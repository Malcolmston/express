# requestid — provides express middleware that assigns each request a unique identifier

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/requestid.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestid)

Package requestid provides express middleware that assigns each request a
unique identifier, echoes it on the response, and stores it on the request for
downstream handlers and logging. It is the express port of the Node
express-request-id middleware and the widely used X-Request-ID convention:
every request is guaranteed to carry an ID that ties together application
logs, client-visible response headers, and, when propagated, traces spanning
multiple services.

Use it as a foundational layer for observability. With a request ID in place,
a log line, an error report, and the response the client received can all be
correlated by the same value, which turns "a user saw an error at 3pm" into a
searchable key. It is also the natural hook for distributed tracing at the
edge: an upstream proxy or gateway can stamp X-Request-Id and this middleware
will honor it, so the same ID flows through your service instead of a new one
being minted at each hop.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/requestid"
```

## Usage

This is the package's own `ExampleNew_reuse`, so it compiles and its output is
asserted on every `go test ./requestid/`.

```go
app := express.New()
	app.Use(requestid.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(requestid.DefaultHeader, "trace-abc")
	app.ServeHTTP(rec, r)

	fmt.Println(rec.Header().Get(requestid.DefaultHeader))
```

```
trace-abc
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that ensures every request carries an id. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the middleware. |

### Constants

`ContextKey`, `DefaultHeader`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
