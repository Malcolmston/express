# healthz — provides a minimal liveness endpoint

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/healthz.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/healthz)

Package healthz provides a minimal liveness endpoint. It is a stdlib-only
analogue of the tiny Node "express-healthcheck" style middleware: a single
handler owns one path and answers it with a fixed plain-text 200 response,
signaling only that the process is up and able to serve requests. It performs
no dependency checks — for that, use the sibling healthcheck package.

Use healthz for the common "is this instance alive?" probe wired to a load
balancer, a container orchestrator's liveness check, or an uptime monitor.
Because it returns immediately with a static body and never touches external
systems, it is cheap and cannot itself be made to fail by a slow dependency,
which is exactly what you want from a liveness (as opposed to readiness)
probe. Reach for healthcheck instead when the endpoint must verify that
downstream dependencies are reachable before reporting healthy.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/healthz"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./healthz/`.

```go
app := express.New()
	app.Use(healthz.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", rec.Body.String())
```

```
status: 200
body: ok
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns healthz middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the healthz middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/healthz).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
