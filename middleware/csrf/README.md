# csrf — provides Cross-Site Request Forgery protection using the double-submit-cookie pattern

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/csrf.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/csrf)

Package csrf provides Cross-Site Request Forgery protection using the
double-submit-cookie pattern. It is the Go analogue of the classic csurf
middleware from the Node ecosystem, packaged as a drop-in express.Handler. A
random token is issued in a cookie and, for state-changing requests, must be
echoed back in a request header or form field; requests whose submitted token
does not match the cookie are rejected with 403 Forbidden and never reach the
downstream handler.

Use this middleware to defend endpoints that mutate server state — form
posts, account changes, deletions — against forged requests triggered by a
malicious third-party site while the victim is authenticated. The
double-submit-cookie scheme works because a foreign origin can cause the
browser to send the victim's cookies but cannot read them to reproduce the
token in a header or body, so a matching submission proves the request
originated from your own pages. Mount it with app.Use for a global guard, or
attach it to the router subtree that owns your state-changing routes.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/csrf"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./csrf/`.

```go
app := express.New()
	app.Use(csrf.New())
	app.Get("/form", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(csrf.Token(req))
	})
	app.Post("/submit", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("accepted")
	})

	// Simulate a client that already holds a token and submits it correctly.
	const token = "example-token-value"
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: "csrf", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
```

```
200
accepted
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns CSRF protection middleware. |
| `func Token(req *express.Request) string` | Token returns the CSRF token associated with the current request, or "" if the middleware did not run. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the CSRF middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/csrf).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
