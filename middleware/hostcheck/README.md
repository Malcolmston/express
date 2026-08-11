# hostcheck — provides middleware that guards against Host header attacks by permitting only requests

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/hostcheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/hostcheck)

Package hostcheck provides middleware that guards against Host header attacks
by permitting only requests whose hostname is in a configured allowlist. It
plays the role of the "allowed hosts" / host-validation middleware common in
the Node ecosystem (comparable to Django's ALLOWED_HOSTS or the host check in
webpack-dev-server), adapted to this Express-style framework with only the
standard library.

Use it whenever your application trusts the incoming Host header for anything
— generating absolute URLs, password-reset links, cache keys, or
virtual-host routing. An attacker who can set an arbitrary Host header can
otherwise poison those values (Host header injection / web cache poisoning)
even when the request reaches your app through a proxy. Restricting requests
to a fixed set of expected hostnames removes that class of attack.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/hostcheck"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./hostcheck/`.

```go
app := express.New()
	app.Use(hostcheck.New(hostcheck.Options{
		Allow: []string{"example.com", "www.example.com"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	send := func(host string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Host = host
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	fmt.Println("allowed:", send("www.example.com:8080"))
	fmt.Println("rejected:", send("evil.com"))
```

```
allowed: 200
rejected: 421
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that rejects requests whose Hostname is not present in the allowlist, defending against Host header spoofing. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the host-check middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/hostcheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
