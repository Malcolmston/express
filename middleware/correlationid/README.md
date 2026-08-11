# correlationid — provides express middleware that assigns a correlation id to every request so

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/correlationid.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/correlationid)

Package correlationid provides express middleware that assigns a correlation
id to every request so that a single logical operation can be traced as it
fans out across multiple services and log streams. It is the Go analogue of
Node middleware such as express-correlation-id, the correlation-id package,
and express-request-id, packaged as a drop-in express.Handler that reads and
writes the X-Correlation-Id header.

Reach for this middleware whenever you need distributed tracing without a
heavyweight tracing stack: correlating the log lines a request produces in an
API gateway, an application, and its downstream dependencies, or stitching
together the story of one client action across a request/response boundary.
Because an inbound id is preserved rather than replaced, the same value flows
from the caller through every hop that forwards the header, giving you a
stable key to group logs, metrics, and error reports by. Mount it globally
with app.Use so that every route is covered, ideally as one of the first
handlers in the chain.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/correlationid"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./correlationid/`.

```go
app := express.New()
	app.Use(correlationid.New(correlationid.Options{
		Generator: func() string { return "fixed-id-123" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		id, _ := req.Value(correlationid.ContextKey)
		fmt.Println("handler saw:", id)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("response header:", rec.Header().Get(correlationid.DefaultHeader))
```

```
handler saw: fixed-id-123
response header: fixed-id-123
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that ensures every request carries a correlation id. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the middleware. |

### Constants

`ContextKey`, `DefaultHeader`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/correlationid).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
