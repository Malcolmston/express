# origincheck — provides middleware that rejects requests whose Origin header is not in a configured

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/origincheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/origincheck)

Package origincheck provides middleware that rejects requests whose Origin
header is not in a configured allowlist, mitigating cross-site request forgery
(CSRF) and other cross-origin abuse from untrusted callers. It implements the
same defensive pattern as Node origin-verification middleware such as the
Origin/Referer checks in csurf-style and koa-csrf-style guards: rather than
issuing and validating tokens, it trusts the browser-set Origin header to
decide whether a request may proceed.

Use it on state-changing endpoints (or an entire app) that are consumed only
by a known set of first-party front ends. Because browsers attach the Origin
header automatically on cross-origin and non-GET requests and forbid pages
from forging it, an allowlist of expected origins is a cheap, stateless way to
block requests initiated by an attacker's site. It is complementary to, not a
replacement for, token-based CSRF protection and standard authentication.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/origincheck"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./origincheck/`.

```go
app := express.New()
	app.Use(origincheck.New(origincheck.Options{
		Allow: []string{"example.com", "app.example.com:8443"},
	}))
	app.Post("/transfer", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	good := httptest.NewRequest("POST", "/transfer", nil)
	good.Header.Set("Origin", "https://example.com")
	goodRec := httptest.NewRecorder()
	app.ServeHTTP(goodRec, good)

	bad := httptest.NewRequest("POST", "/transfer", nil)
	bad.Header.Set("Origin", "https://evil.example.net")
	badRec := httptest.NewRecorder()
	app.ServeHTTP(badRec, bad)

	fmt.Println(goodRec.Code)
	fmt.Println(badRec.Code)
```

```
200
403
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 unless the request's Origin header host appears in the allowlist. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the origin-check middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/origincheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
