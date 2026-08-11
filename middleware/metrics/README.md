# metrics — provides middleware that records basic request metrics for the express framework

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/metrics.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/metrics)

Package metrics provides middleware that records basic request metrics for the
express framework: the total number of requests and a breakdown by status
class (2xx, 3xx, 4xx, 5xx). It is a minimal, dependency-free analog of Node
instrumentation middleware such as express-status-monitor, morgan's counting
variants, or a prom-client HTTP collector, but it stays in the standard
library and simply accumulates counters in memory rather than exporting a
Prometheus endpoint.

Use it when you want a cheap, always-on sense of request volume and error
ratios — for health dashboards, a debug/admin route, or lightweight alerting
on the 5xx count — without pulling in a full metrics stack. It answers "how
many requests have we served and how many failed" and nothing more, which is
often exactly what a small service needs.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/metrics"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./metrics/`.

```go
handler, m := metrics.New()

	app := express.New()
	app.Use(handler)
	app.Get("/ok", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	app.Get("/boom", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(500).Send("boom")
	})

	for _, path := range []string{"/ok", "/ok", "/boom"} {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	}

	snap := m.Snapshot()
	_ = snap["total"]
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Metrics` | Metrics holds request counters observed by the middleware. |

<details>
<summary><code>Metrics</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New() (express.Handler, *Metrics)` | New returns metrics middleware together with the *Metrics accumulator it writes to. |
| `func (m *Metrics) Snapshot() map[string]int64` | Snapshot returns a copy of the current counters. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/metrics).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
