# panicjson — provides middleware that recovers from panics raised by downstream handlers, logs them

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/panicjson.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/panicjson)

Package panicjson provides middleware that recovers from panics raised by
downstream handlers, logs them, and responds with a generic 500 JSON error so
that a single failing request cannot crash the process. It is the express
analogue of Node error-recovery middleware — Express's default error handler
and packages such as Connect's errorhandler or Koa's try/catch bodies — but
specialized to always answer with a compact JSON envelope.

Use it as an outer safety net around routes that may panic, whether from
programmer error, a failed type assertion, a nil dereference, or a
deliberately raised value. In Go a panic that escapes an HTTP handler would
otherwise unwind the goroutine serving the request; wrapping the chain in this
middleware converts that fatal condition into a well-formed 500 response and a
log line, keeping the server available for other clients.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/panicjson"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./panicjson/`.

```go
app := express.New()
	app.Use(panicjson.New(panicjson.Options{
		Logger: log.New(io.Discard, "", 0),
	}))
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
```

```
500
{"error":"Internal Server Error"}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that recovers from panics raised by later handlers. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the panic recovery middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/panicjson).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
