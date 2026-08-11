# uptime — provides middleware that reports how long the middleware -- and, by extension

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/uptime.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/uptime)

Package uptime provides middleware that reports how long the middleware --
and, by extension, the process serving requests -- has been running. It has no
single Node original but plays the role of the small "uptime" helpers and
health-endpoint fields common in Express apps (for example the
process.uptime() value surfaced by express-status-monitor or a custom /healthz
handler), packaged here as a drop-in express.Handler paired with a
programmatic accessor.

Reach for this middleware when you want a cheap, always-available liveness
signal: load balancers and orchestrators can read the header to gauge how long
an instance has been up, operators can spot a process that recently restarted
(a crash-loop shows as a header that keeps resetting to a small number), and
dashboards can chart per-instance age. Because it adds a single header and
never blocks, it is safe to mount globally in front of the whole application.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/uptime"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./uptime/`.

```go
clock := time.Unix(1000, 0)
	handler, since := uptime.New(uptime.Options{
		Now: func() time.Time { return clock },
	})

	app := express.New()
	app.Use(handler)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	clock = clock.Add(42 * time.Second)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Printf("X-Uptime=%s since=%v\n", w.Header().Get("X-Uptime"), since())
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) (express.Handler, func() time.Duration)` | New returns uptime middleware together with a Since accessor reporting the elapsed time since the middleware was created. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the uptime middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/uptime).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
