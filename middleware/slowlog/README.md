# slowlog — provides middleware that logs a warning whenever a request takes longer than

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/slowlog.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/slowlog)

Package slowlog provides middleware that logs a warning whenever a request
takes longer than a configured threshold to complete, helping surface slow
endpoints. It plays the same role as Node timing loggers such as response-time
paired with a threshold check, or the "slow request" warnings emitted by
morgan-style access logs, but distilled to a single concern: measure
wall-clock handler time and shout only when it crosses a line.

Reach for this middleware as lightweight, always-on performance triage. Mount
it globally so every route is timed, and the log fills only with the requests
that actually ran slowly — a cheap way to catch a database call that got
slow under load, an N+1 query, or an endpoint that degrades in production
without standing up a full tracing stack. Because the quiet path costs one
time.Now and one subtraction, it is inexpensive enough to leave enabled
everywhere.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/slowlog"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./slowlog/`.

```go
var buf bytes.Buffer
	app := express.New()
	app.Use(slowlog.New(slowlog.Options{
		Threshold: time.Millisecond,
		Logger:    log.New(&buf, "", 0),
	}))
	app.Get("/report", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(5 * time.Millisecond)
		res.Send("done")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/report", nil))

	line := strings.TrimSpace(buf.String())
	fmt.Println("warned:", strings.Contains(line, "WARNING"))
	fmt.Println("names route:", strings.Contains(line, "GET /report"))
```

```
warned: true
names route: true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that measures the time spent handling each request and logs a warning when it exceeds the threshold. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the slow-request logger. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/slowlog).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
