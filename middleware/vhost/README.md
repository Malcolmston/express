# vhost — provides virtual-host middleware that dispatches requests to a dedicated handler based

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/vhost.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/vhost)

Package vhost provides virtual-host middleware that dispatches requests to a
dedicated handler based on the request's hostname. It supports exact matches
and a leading "*." wildcard that matches any single-or-multi-label subdomain.
It is a port of the Node "vhost" connect/Express middleware: a way to run
several logical sites — an API host, a docs host, per-tenant subdomains —
behind a single server and route each incoming request to the right handler
purely by its Host header.

Reach for it when one process must serve multiple domains or subdomains and
you want a clean split between their handler trees rather than branching on
the hostname inside every route. Typical uses are directing "api.example.com"
to an API sub-application, giving "*.tenant.example.com" a shared multi-tenant
handler, or carving "admin.example.com" off from the public site. Because
matching is driven by the client-supplied Host header, vhost is a routing
convenience, not an authentication or trust boundary.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/vhost"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./vhost/`.

```go
app := express.New()
	app.Use(vhost.New(vhost.Options{
		Host: "api.example.com",
		Handler: func(req *express.Request, res *express.Response, next express.Next) {
			res.Send("api response")
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("main site")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "http://api.example.com/", nil))

	fmt.Println(rec.Body.String())
```

```
api response
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that invokes Options.Handler for requests whose hostname matches Options.Host; non-matching requests fall through to next. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the vhost middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/vhost).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
