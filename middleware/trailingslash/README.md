# trailingslash — provides middleware that enforces a consistent trailing-slash policy on request paths

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/trailingslash.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/trailingslash)

Package trailingslash provides middleware that enforces a consistent
trailing-slash policy on request paths by redirecting requests that do not
conform. It is the express framework's Go analogue of Node middleware such as
express-slash and connect-slashes: a drop-in express.Handler that either adds
a trailing slash to every path or strips it, issuing an HTTP redirect so that
each resource is reachable at exactly one canonical URL.

Reach for this middleware when you care about URL canonicalization -- for SEO
(search engines treat /about and /about/ as distinct URLs and split link
equity between them), for cache-hit consistency, or simply to keep your route
table from having to register both forms of every path. Choosing one policy
and redirecting the other form guarantees clients and crawlers converge on a
single spelling of each URL.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/trailingslash"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./trailingslash/`.

```go
app := express.New()
	app.Use(trailingslash.New(trailingslash.Options{Enforce: true}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	call := func(path string) {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		if loc := w.Header().Get("Location"); loc != "" {
			fmt.Printf("%-12s -> %d %s\n", path, w.Code, loc)
		} else {
			fmt.Printf("%-12s -> %d %s\n", path, w.Code, w.Body.String())
		}
	}
	call("/about")
	call("/about/")
	call("/report?y=1")
```

```
/about       -> 301 /about/
/about/      -> 200 ok
/report?y=1  -> 301 /report/?y=1
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware implementing the configured trailing-slash policy. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the trailing-slash middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/trailingslash).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
