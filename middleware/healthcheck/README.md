# healthcheck — provides a liveness/readiness endpoint for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/healthcheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/healthcheck)

Package healthcheck provides a liveness/readiness endpoint for the express
framework. It is a stdlib-only analogue of Node health-check middlewares such
as "express-healthcheck" and the "@cloudnative/health-connect" / Terminus
style probes: a single mounted handler owns one path, runs a set of
user-supplied component checks, and reports their aggregate result as JSON
suitable for a load balancer, Kubernetes readiness probe, or uptime monitor.

Use this rather than the sibling healthz package when a bare "the process is
up" signal is not enough and you need to verify that dependencies — a
database, a cache, a downstream API — are actually reachable before
declaring the instance ready to receive traffic. Each dependency is expressed
as a named function returning error; returning nil means healthy and returning
a non-nil error means unhealthy, with the error text surfaced in the response
so operators can see *why* a check failed.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/healthcheck"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./healthcheck/`.

```go
app := express.New()
	app.Use(healthcheck.New(healthcheck.Options{
		Checkers: map[string]func() error{
			"db":    func() error { return nil },
			"cache": func() error { return nil },
		},
	}))

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", rec.Body.String())
```

```
status: 200
body: {"status":"ok","checks":{"cache":"ok","db":"ok"}}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns healthcheck middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the healthcheck middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/healthcheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
