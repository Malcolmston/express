# readiness — provides a readiness-probe endpoint for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/readiness.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/readiness)

Package readiness provides a readiness-probe endpoint for the express
framework. It exposes a single well-known path (defaulting to "/readyz") that
answers 200 "ready" when the application reports itself ready to serve traffic
and 503 "not ready" otherwise; requests to any other path are left untouched
and continue down the chain. It ports the readiness-gating concept that Node
process managers such as terminus and lightship attach to an HTTP server,
where an orchestrator (Kubernetes, a load balancer, a service mesh) polls the
endpoint and only routes live traffic to instances that report themselves
ready.

Use it to separate "the process is up" from "the process should receive
requests." A server may be running yet still be warming a cache, replaying a
migration, draining connections before shutdown, or waiting on a downstream
dependency; during those windows the readiness probe should fail so the
orchestrator holds traffic back without killing the pod (that is the job of a
separate liveness probe). Because the Ready callback is evaluated on every
probe, the same endpoint reflects the current state and flips back to 200 the
moment the application recovers, which the tests exercise directly.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/readiness"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./readiness/`.

```go
ready := false

	app := express.New()
	app.Use(readiness.New(readiness.Options{
		Path:  "/readyz",
		Ready: func() bool { return ready },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("home")
	})

	probe := func() (int, string) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/readyz", nil)
		app.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}

	code, body := probe()
	fmt.Printf("before: %d %s\n", code, body)

	ready = true
	code, body = probe()
	fmt.Printf("after: %d %s\n", code, body)

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("route: %d %s\n", rec.Code, rec.Body.String())
```

```
before: 503 not ready
after: 200 ready
route: 200 home
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns readiness middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the readiness middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/readiness).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
