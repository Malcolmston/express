# poweredby — provides express middleware that sets the X-Powered-By response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/poweredby.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/poweredby)

Package poweredby provides express middleware that sets the X-Powered-By
response header to a configurable value. It is the deliberate inverse of the
hidepoweredby middleware (and of Helmet's hidePoweredBy feature): where those
strip the header to avoid advertising the technology stack, poweredby writes
an explicit, chosen value so branding is under your control rather than left
to a framework default.

In stock Express, the app implicitly emits "X-Powered-By: Express" on every
response unless the "x-powered-by" application setting is disabled. This
middleware ports that behaviour to the stdlib-only express port: rather than
relying on any implicit framework default, it sets the header to Value on each
response. This makes the branding value explicit and testable, and lets
applications present a custom product string (for example "MyApp/2.0") or
suppress the header entirely.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/poweredby"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./poweredby/`.

```go
app := express.New()
	app.Use(poweredby.New(poweredby.Options{Value: "MyApp/2.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	app.ServeHTTP(rr, req)

	fmt.Println(rr.Header().Get("X-Powered-By"))
```

```
MyApp/2.0
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the X-Powered-By response header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the middleware. |

### Constants

`DefaultValue`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/poweredby).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
