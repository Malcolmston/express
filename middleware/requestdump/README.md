# requestdump — provides development middleware that records a snapshot of each incoming request

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/requestdump.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestdump)

Package requestdump provides development middleware that records a snapshot of
each incoming request (method, path, headers, and capture time) into a
bounded, thread-safe ring buffer for later inspection. It is the express
analogue of the request-dump/inspection helpers used while developing Node
apps — the kind of thing you reach for alongside morgan or a console.dir of
req — but instead of writing to a log it retains the most recent requests in
memory so tests or a debug endpoint can read them back structurally.

Use it during development and testing when you need to assert on, or eyeball,
what actually arrived at the server: which method and path were hit and which
headers the client sent. A test can drive the app and then call Last to
inspect the final request, or a debug route can call All to render a table of
recent traffic. It is deliberately not a production tool — captured headers
may contain sensitive material (cookies, Authorization), and the buffer is a
process-local, non-persistent side channel — so it should be mounted behind
development guards rather than enabled globally in production.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/requestdump"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./requestdump/`.

```go
requestdump.Reset()

	app := express.New()
	app.Use(requestdump.New())
	app.Get("/users", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users", nil))

	d := requestdump.Last()
	fmt.Println(d.Method, d.Path)
```

```
GET /users
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that records each request into the ring buffer. |
| `func Reset()` | Reset clears the captured request history. |

### Types

| Type | What it is |
| --- | --- |
| `Dump` | Dump is a captured snapshot of a single request. |
| `Options` | Options configures the request dump middleware. |

<details>
<summary><code>Dump</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func All() []Dump` | All returns a copy of the currently retained request snapshots, oldest first. |
| `func Last() Dump` | Last returns the most recently captured request. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/requestdump).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
