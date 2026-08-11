# helmet — bundles a sensible set of security-related HTTP response headers into a single express

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/helmet.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/helmet)

Package helmet bundles a sensible set of security-related HTTP response
headers into a single express middleware, mirroring the popular Node.js Helmet
defaults. It is a stdlib-only port of helmet's default header set to this
Express-style framework, composing what upstream ships as several independent
sub-middleware into one handler so a single app.Use hardens every response.

Reach for helmet as a baseline security layer on any web application: it is
the quickest way to turn on a batch of widely recommended, low-risk response
headers without wiring each one individually. Each header addresses a specific
attack surface — MIME sniffing, clickjacking, protocol downgrade, referrer
leakage, DNS prefetch tracking, cross-domain policy abuse, stack
fingerprinting, and cross-origin process isolation — and applying them
together closes several classes of common web vulnerabilities at once. If you
need finer control over any single header you can still use the dedicated
sibling middleware (hsts, frameguard, hidepoweredby, dnsprefetch, and so on),
which this package mirrors.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/helmet"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./helmet/`.

```go
app := express.New()
	app.Use(helmet.New(helmet.Options{
		FrameguardAction:      "DENY",
		HSTSIncludeSubDomains: true,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("X-Frame-Options:", rec.Header().Get("X-Frame-Options"))
```

```
X-Frame-Options: DENY
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets helmet's bundle of default security headers. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the helmet middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/helmet).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
