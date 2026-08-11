# requestcontext — provides middleware that attaches a small request-scoped context to each request

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/requestcontext.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestcontext)

Package requestcontext provides middleware that attaches a small
request-scoped context to each request: a unique request ID and a start
timestamp. It is the express analogue of the request-scoped value injection
commonly done in Node with a context.Context-style object or libraries such as
express-http-context and cls-hooked, where per-request data is stashed so any
downstream handler can retrieve it without threading parameters through every
call. The ID is also echoed in the X-Request-Id response header, easing log
correlation across services.

Use it when handlers, loggers, or instrumentation deeper in the chain need a
stable identifier and a request start time without each layer re-deriving
them. Because the *Ctx is stored on the request itself, a logging helper or
error formatter can call From(req) and immediately obtain the same ID that was
returned to the client, which is the foundation of tracing a single request
across application logs and, when propagated, across service boundaries. The
Ctx.Elapsed helper turns the captured Start time into a running request
duration suitable for latency logging.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/requestcontext"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./requestcontext/`.

```go
app := express.New()
	app.Use(requestcontext.New(requestcontext.Options{
		Generator: func() string { return "req-123" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		ctx := requestcontext.From(req)
		res.Send(ctx.ID)
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("ctx id:", rec.Body.String())
	fmt.Println("header:", rec.Header().Get(requestcontext.HeaderName))
```

```
ctx id: req-123
header: req-123
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that attaches a *Ctx via req.Set(Key, ctx) and sets the X-Request-Id response header. |

### Types

| Type | What it is |
| --- | --- |
| `Ctx` | Ctx is the per-request context attached by this middleware. |
| `Options` | Options configures the requestcontext middleware. |

<details>
<summary><code>Ctx</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func From(req *express.Request) *Ctx` | From returns the *Ctx attached to the request, or nil if none is present. |
| `func (c *Ctx) Elapsed() time.Duration` | Elapsed returns the duration since the request started. |

</details>

### Constants

`HeaderName`, `Key`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestcontext).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
