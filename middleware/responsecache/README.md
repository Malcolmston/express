# responsecache — provides simple in-memory caching of successful GET responses

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/responsecache.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/responsecache)

Package responsecache provides simple in-memory caching of successful GET
responses. It is the express framework's Go analogue of Node caching
middleware such as apicache, express-cache-response, and the memory store
behind route-level res.set('Cache-Control') helpers: a drop-in express.Handler
that memoizes a handler's output and replays it to later callers without
re-running the handler. The cache lives entirely in process memory, is keyed
by request URI, and expires entries after a configurable time-to-live.

Reach for this middleware to shed load from expensive but idempotent GET
endpoints — a rendered page, an aggregation query, a third-party API proxy
— whose responses can tolerate being a few seconds stale. It is best suited
to read-only, anonymous, or otherwise cache-friendly routes; it is not a CDN,
a shared cache, or a substitute for HTTP validators, and it does not honor
client Cache-Control or per-user variation. Because it is a single unbounded
map, use it for a bounded set of hot URIs rather than as a general store.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/responsecache"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./responsecache/`.

```go
app := express.New()
	app.Use(responsecache.New(responsecache.Options{TTL: time.Minute}))

	calls := 0
	app.Get("/report", func(req *express.Request, res *express.Response, next express.Next) {
		calls++
		res.Type("text").Send("expensive result")
	})

	do := func() {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/report", nil))
		fmt.Printf("%s %d %s\n", w.Header().Get("X-Cache"), w.Code, w.Body.String())
	}

	do()
	do()

	fmt.Printf("handler ran %d time(s)\n", calls)
```

```
MISS 200 expensive result
HIT 200 expensive result
handler ran 1 time(s)
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that caches successful GET responses in memory for the configured TTL. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the response cache. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/responsecache).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
