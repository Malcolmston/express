# redirectmap — provides middleware that redirects requests whose path matches an entry in a static

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/redirectmap.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/redirectmap)

Package redirectmap provides middleware that redirects requests whose path
matches an entry in a static lookup table. It ports the pattern of a static
redirect map — the kind of "old path -> new path" table often expressed as a
plain object handed to Node's express-urlrewrite, a hand-written redirect
middleware, or the redirects section of a hosting provider's config — into a
single express handler backed by a Go map.

Use it to migrate or consolidate URLs without registering an individual route
for each one: renamed pages, retired endpoints that should point at their
replacement, vanity links, or a bulk import of legacy paths. Because lookups
are a single map access, a table with thousands of entries costs the same as a
table with one, which makes it a convenient place to park large, data-driven
redirect sets that would be tedious to wire up as routes.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/redirectmap"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./redirectmap/`.

```go
app := express.New()
	app.Use(redirectmap.New(redirectmap.Options{
		Map:    map[string]string{"/old": "/new"},
		Status: http.StatusMovedPermanently,
	}))
	app.Get("/keep", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("kept")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/old", nil))
	fmt.Printf("redirect: %d -> %s\n", rec.Code, rec.Header().Get("Location"))

	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/keep", nil))
	fmt.Printf("fallthrough: %d %s\n", rec.Code, rec.Body.String())
```

```
redirect: 301 -> /new
fallthrough: 200 kept
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that redirects any request whose path is present in the map to the mapped destination; other requests fall through to next. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the redirect-map middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/redirectmap).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
