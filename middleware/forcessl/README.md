# forcessl — provides middleware that redirects insecure HTTP requests to their HTTPS equivalent

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/forcessl.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/forcessl)

Package forcessl provides middleware that redirects insecure HTTP requests to
their HTTPS equivalent, preserving the host, path, and query string. It is a
stdlib-only port of the Node "express-force-ssl" (and the closely related
"express-sslify") middleware: when a request arrives over plain HTTP it is
bounced to the same URL under the https scheme so that browsers and clients
upgrade to an encrypted connection.

Use this middleware at the very top of the chain on any deployment that must
never serve content over cleartext HTTP — for example a public site behind a
load balancer or reverse proxy that also terminates or forwards TLS. Placing
it first ensures the redirect happens before any handler runs, so no
application logic executes and no cookies or bodies are processed on the
insecure request. It pairs naturally with an HSTS header set by a separate
security middleware once the client is on HTTPS.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/forcessl"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./forcessl/`.

```go
app := express.New()
	app.Use(forcessl.New())
	app.Get("/path", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("should not reach here over http")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/path?x=1", nil)
	app.ServeHTTP(rec, req)

	fmt.Println("status:", rec.Code)
	fmt.Println("location:", rec.Header().Get("Location"))
```

```
status: 301
location: https://example.com/path?x=1
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that redirects http requests to https with a 301 (Moved Permanently). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the force-SSL middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/forcessl).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
