# csp — provides middleware that builds and sets a Content-Security-Policy response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/csp.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/csp)

Package csp provides middleware that builds and sets a Content-Security-Policy
response header from a directive map. It is the Go analogue of Helmet's
contentSecurityPolicy middleware from the Node ecosystem, packaged as a
drop-in express.Handler. Where Helmet accepts a directives object and renders
it into the header string, this port accepts a Go map of directive names to
source lists and does the same rendering, then attaches the value to every
response that flows through it.

Use this middleware to instruct browsers about which origins are legitimate
sources of scripts, styles, images, fonts, frames, and other resources, which
is the single most effective defense against cross-site scripting and
resource-injection attacks. A policy of "default-src 'self'" alone already
blocks inline scripts and third-party code; tighten it further by adding
directives such as "script-src", "style-src", or "img-src". Mount it globally
with app.Use so the header is present on every response, or attach it to a
specific router or path prefix when only part of the application needs a
policy.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/csp"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./csp/`.

```go
app := express.New()
	app.Use(csp.New(csp.Options{Directives: map[string][]string{
		"default-src": {"'self'"},
		"script-src":  {"'self'", "https://cdn.example.com"},
	}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Content-Security-Policy"))
```

```
default-src 'self'; script-src 'self' https://cdn.example.com
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Build(directives map[string][]string) string` | Build renders a directive map into a Content-Security-Policy header value. |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets a Content-Security-Policy header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the csp middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/csp).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
