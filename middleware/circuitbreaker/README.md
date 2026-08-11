# circuitbreaker — provides middleware implementing the circuit-breaker pattern

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/circuitbreaker.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/circuitbreaker)

Package circuitbreaker provides middleware implementing the circuit-breaker
pattern. It observes the status of downstream responses through a
ResponseWriter wrapper; once a configured number of consecutive 5xx responses
occur, the circuit "opens" and subsequent requests are short-circuited with a
503 response without invoking the downstream chain. After a cooldown period
elapses the circuit moves to a half-open state, allowing a single trial
request whose outcome closes or re-opens it. It is the express framework's Go
analogue of Node resiliency middleware such as opossum and
express-circuit-breaker.

Use this middleware to stop a failing dependency from taking the rest of the
service down with it. When a route proxies to a database, a third-party API,
or another microservice that has started returning errors, hammering it with
more requests only deepens the outage and ties up goroutines waiting on
timeouts. Opening the circuit sheds that load instantly, returns a fast 503 to
callers instead of a slow failure, and gives the downstream dependency room to
recover. Mount it around the specific routes whose failures you want to
isolate rather than globally, since one breaker tracks one failure stream.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/circuitbreaker"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./circuitbreaker/`.

```go
now := time.Unix(0, 0)
	healthy := false

	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 2,
		Cooldown:  30 * time.Second,
		Message:   "circuit open",
		Now:       func() time.Time { return now },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if healthy {
			res.Send("ok")
			return
		}
		res.Status(500).Send("upstream down")
	})

	do := func() {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("%d %q\n", w.Code, w.Body.String())
	}

	do()
	do()
	do()

	now = now.Add(31 * time.Second)
	healthy = true
	do()
```

```
500 "upstream down"
500 "upstream down"
503 "circuit open"
200 "ok"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns circuit-breaker middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the circuit breaker. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/circuitbreaker).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
