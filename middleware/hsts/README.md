# hsts — provides middleware that sets the HTTP Strict-Transport-Security (HSTS) response

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/hsts.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/hsts)

Package hsts provides middleware that sets the HTTP Strict-Transport-Security
(HSTS) response header, instructing conforming browsers to access the site
only over HTTPS. It is a stdlib-only port of the Node.js "hsts" package (also
exposed as helmet.hsts) to this Express-style framework, and produces the same
header values for equivalent options.

Use this middleware on any site served over TLS to defend against protocol
downgrade attacks and cookie hijacking. Once a browser has seen the
Strict-Transport-Security header it will, for the lifetime of max-age,
automatically upgrade http:// requests to https:// and refuse to connect if
the certificate is invalid, closing the window in which a network attacker
could intercept a plaintext request. It does not itself perform any redirect;
pair it with a force-HTTPS redirect for the very first, still-plaintext hit.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/hsts"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./hsts/`.

```go
app := express.New()
	app.Use(hsts.New(hsts.Options{
		MaxAge:            31536000,
		IncludeSubDomains: true,
		Preload:           true,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Strict-Transport-Security"))
```

```
max-age=31536000; includeSubDomains; preload
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Strict-Transport-Security header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the HSTS middleware. |

### Constants

`DefaultMaxAge`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/hsts).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
