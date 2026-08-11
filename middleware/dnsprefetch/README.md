# dnsprefetch — provides middleware that sets the X-DNS-Prefetch-Control response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/dnsprefetch.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/dnsprefetch)

Package dnsprefetch provides middleware that sets the X-DNS-Prefetch-Control
response header, controlling whether browsers speculatively resolve the DNS
names of links and resources on a page. It is the Go analogue of Helmet's
dnsPrefetchControl middleware from the Node ecosystem, packaged as a drop-in
express.Handler that writes the same header with the same "on"/"off"
semantics.

Use this middleware to make a deliberate choice about DNS prefetching rather
than leaving it to browser defaults. Prefetching improves perceived
performance by resolving hostnames before the user clicks, but it also leaks a
signal to DNS resolvers about which links a page contains, which can be a
minor privacy concern for pages that reference sensitive third-party hosts.
Turning it off (the default here) prioritizes privacy; turning it on
prioritizes latency. Mount it with app.Use for a site-wide policy, or attach
it to a specific router when only some pages warrant a different setting.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/dnsprefetch"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./dnsprefetch/`.

```go
app := express.New()
	app.Use(dnsprefetch.New(dnsprefetch.Options{Allow: true}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-DNS-Prefetch-Control"))
```

```
on
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the X-DNS-Prefetch-Control header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the dnsprefetch middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/dnsprefetch).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
