# slowdown — provides middleware that progressively delays responses for a client once it exceeds

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/slowdown.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/slowdown)

Package slowdown provides middleware that progressively delays responses for a
client once it exceeds a configured request threshold within a fixed window.
It is the express framework's Go analogue of the Node express-slow-down
package (express-rate-limit/express-slow-down): rather than rejecting bursts
outright the way a rate limiter does, it lets the first requests through
untouched and then adds a growing pause to each further request, gently
discouraging abuse while keeping legitimate traffic working.

Reach for this middleware to blunt brute-force and scraping attempts, to
smooth spikes against an expensive endpoint, or as a softer companion to a
hard rate limiter — the slow-down absorbs the shoulder of a burst so fewer
requests ever reach the limiter's cutoff. It pairs naturally with a login or
search route where a slightly slower response is an acceptable price for
throttling an attacker but a limiter's outright 429 would be too blunt.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/slowdown"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./slowdown/`.

```go
var delays []time.Duration
	app := express.New()
	app.Use(slowdown.New(slowdown.Options{
		Threshold: 2,
		Delay:     20 * time.Millisecond,
		Sleep:     func(d time.Duration) { delays = append(delays, d) },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	for i := 0; i < 6; i++ {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "203.0.113.7:9000"
		app.ServeHTTP(httptest.NewRecorder(), r)
	}

	fmt.Println("delayed requests:", len(delays))
	_ = delays
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns slow-down middleware configured by opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the slow-down middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/slowdown).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
