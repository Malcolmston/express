# ipallowlist — provides middleware that only permits requests whose client IP address matches

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/ipallowlist.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/ipallowlist)

Package ipallowlist provides middleware that only permits requests whose
client IP address matches a configured allowlist of exact addresses or CIDR
ranges, rejecting everything else with 403 Forbidden. It fills the same role
as the Node "express-ipfilter" package (used in "allow"/whitelist mode) and
the general pattern of an IP-based access gate, reimplemented here on top of
Go's standard net package with no third-party dependencies.

Reach for this middleware when a route or an entire application must only be
reachable from a known set of clients — an internal admin panel, a metrics
or health endpoint scraped by fixed collectors, a webhook receiver locked to a
provider's published ranges, or a service fronted by a load balancer whose
egress IPs are known. It is a positive/allow model: anything not explicitly
listed is denied, which is the safer default for restricting access. For the
inverse (deny a few, allow the rest) use the sibling ipblocklist package.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/ipallowlist"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./ipallowlist/`.

```go
app := express.New()
	app.Use(ipallowlist.New(ipallowlist.Options{
		Allow: []string{"10.0.0.5", "192.168.0.0/24"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	call := func(clientIP string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = clientIP + ":1234"
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, r)
		return rec.Code
	}

	fmt.Println(call("192.168.0.42"))
	fmt.Println(call("8.8.8.8"))
```

```
200
403
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 unless the request's client IP matches one of the configured allowlist entries. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the IP allowlist middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/ipallowlist).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
