# maintenance — provides middleware that puts an application into maintenance mode

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/maintenance.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/maintenance)

Package maintenance provides middleware that puts an application into
maintenance mode. While enabled, every request is short-circuited with a 503
Service Unavailable response and a Retry-After header. It plays the role that
the Node/Express ecosystem fills with packages such as express-maintenance or
a hand-rolled "under construction" gate, or that operators implement at the
reverse proxy with an HTTP 503 maintenance page — here delivered as a
dependency-free handler that also exposes a programmatic runtime toggle.

Reach for this middleware when you need to take an application offline
gracefully: during a database migration, a deploy that must not receive
traffic, an incident where you want to shed load while keeping the process
alive, or a scheduled maintenance window. Returning a well-formed 503 with a
Retry-After header is friendlier to clients, crawlers, and monitoring than
dropping connections, and because the gate flips at runtime you can enable or
disable it without restarting or redeploying the server.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/maintenance"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./maintenance/`.

```go
handler, toggle := maintenance.New(maintenance.Options{RetryAfter: 120})

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("off:  status=%d\n", rec.Code)

	toggle.Set(true)
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	fmt.Printf("on:   status=%d retry-after=%s\n", rec.Code, rec.Header().Get("Retry-After"))
```

```
off:  status=200
on:   status=503 retry-after=120
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the maintenance middleware. |
| `Toggle` | Toggle enables or disables maintenance mode for middleware constructed without a custom EnabledFunc. |

<details>
<summary><code>Toggle</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(opts ...Options) (express.Handler, *Toggle)` | New returns maintenance middleware and a Toggle for controlling it. |
| `func (t *Toggle) Enabled() bool` | Enabled reports whether maintenance mode is currently on. |
| `func (t *Toggle) Set(on bool)` | Set turns maintenance mode on (true) or off (false). |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/maintenance).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
