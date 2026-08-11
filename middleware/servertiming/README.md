# servertiming — provides express middleware that collects timing metrics during request handling

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/servertiming.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/servertiming)

Package servertiming provides express middleware that collects timing metrics
during request handling and emits them in a Server-Timing response header,
which browsers surface in their developer tools' network panel. It lets a Go
service report how long individual server-side phases took — a database
query, a template render, an upstream call — directly to the browser
inspecting the response, in the same standardized format that Node middleware
such as server-timing produces.

Reach for it when you want per-request server observability that requires no
external tooling: because the numbers ride along on the response the client
already receives, a developer can open dev tools and see the breakdown next to
the request, and synthetic monitors or tests can parse the header to assert on
backend latency. Each metric is a named segment with a duration in
milliseconds and an optional human-readable description, so a single header
can carry several labelled measurements at once.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/servertiming"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./servertiming/`.

```go
app := express.New()
	app.Use(servertiming.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		m := servertiming.From(req)
		m.Add("db", 12500*time.Microsecond)
		m.AddWithDesc("render", "template render", 3*time.Millisecond)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get(servertiming.HeaderName) != "")
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that installs a Metrics collector on the request and emits the accumulated Server-Timing header just before the response is… |

### Types

| Type | What it is |
| --- | --- |
| `Metrics` | Metrics collects Server-Timing entries for a single request. |

<details>
<summary><code>Metrics</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func From(req *express.Request) *Metrics` | From returns the Metrics collector associated with the request, allowing handlers to add timing entries. |
| `func (m *Metrics) Add(name string, dur time.Duration)` | Add records a metric with the given name and duration. |
| `func (m *Metrics) AddWithDesc(name, desc string, dur time.Duration)` | AddWithDesc records a metric with a human-readable description. |

</details>

### Constants

`HeaderName`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/servertiming).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
