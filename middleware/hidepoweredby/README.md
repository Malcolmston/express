# hidepoweredby — provides middleware that removes the X-Powered-By response header (or replaces it

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/hidepoweredby.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/hidepoweredby)

Package hidepoweredby provides middleware that removes the X-Powered-By
response header (or replaces it with a decoy value) so the server does not
advertise the technology stack it runs on. It is a stdlib-only port of the
Node.js "hide-powered-by" package (exposed in helmet as hidePoweredBy) to this
Express-style framework.

Reach for this middleware whenever you want to reduce information leakage in
your response headers. By default many frameworks — including this one,
whose "x-powered-by" application setting is enabled by default — advertise
themselves with an X-Powered-By header. That value gives attackers a free hint
about the server software and version to target, so security hardening guides
routinely recommend stripping it. Supplying a decoy value instead (for example
pretending to be a different stack) can additionally mislead automated
scanners.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/hidepoweredby"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./hidepoweredby/`.

```go
app := express.New()
	app.Use(hidepoweredby.New(hidepoweredby.Options{SetTo: "PHP/8.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Powered-By", "Express")
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-Powered-By"))
```

```
PHP/8.0
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that hides or spoofs the X-Powered-By header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the hidepoweredby middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/hidepoweredby).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
