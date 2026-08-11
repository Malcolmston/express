# rewrite — provides URL-rewriting middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/rewrite.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/rewrite)

Package rewrite provides URL-rewriting middleware. It transparently changes
the request path before it reaches the routing layer, supporting regular
expression matching with $1-style capture-group substitution. It is the
express framework's Go analogue of Node URL rewriters such as
express-urlrewrite and connect-modrewrite, and of the internal-rewrite subset
of Apache's mod_rewrite: an express.Handler that quietly maps an incoming URL
onto a different internal path without the client ever seeing a redirect.

Reach for this middleware to decouple your public URL surface from your route
definitions — serving legacy or vanity paths from a new handler, collapsing
a family of URLs onto one route, versioning an API prefix, or normalizing
paths — all without issuing 3xx redirects or teaching every handler about
the old shape. Because the rewrite is internal, the browser's address bar and
the client's view of the URL are unchanged; only the server's routing sees the
new path. Use a redirect (res.Redirect) instead when you actually want the
client to learn the canonical URL.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/rewrite"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./rewrite/`.

```go
app := express.New()

	app.Use(rewrite.New(rewrite.Options{
		Rules: []rewrite.Rule{
			{Pattern: `^/old/(.*)$`, To: "/new/$1"},
		},
	}))

	app.Get("/new/widgets", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("served by /new/widgets")
	})

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest("GET", "/old/widgets", nil))
	fmt.Printf("%d %s\n", w.Code, w.Body.String())
```

```
200 served by /new/widgets
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that rewrites req.Raw.URL.Path according to the first matching rule, then calls next. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the rewrite middleware. |
| `Rule` | Rule describes a single rewrite. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/rewrite).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
