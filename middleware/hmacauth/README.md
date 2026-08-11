# hmacauth — provides middleware that authenticates requests by verifying an HMAC-SHA256 signature

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/hmacauth.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/hmacauth)

Package hmacauth provides middleware that authenticates requests by verifying
an HMAC-SHA256 signature of the request body against a shared secret. It fills
the role of Node webhook-verification helpers such as the signature checks
used by GitHub, Stripe, and Shopify webhooks, or the generic
"express-hmac-auth" style middleware, implemented here for this Express-style
framework using only crypto/hmac and crypto/sha256 from the standard library.

Use it to protect endpoints — most commonly inbound webhooks and
service-to-service APIs — where both peers share a secret key and the caller
signs each request. Verifying an HMAC over the body proves two things at once:
that the sender knew the secret (authentication) and that the body was not
altered in transit (integrity). Unlike a static bearer token, the signature
changes with every payload, so a captured signature cannot be replayed against
a different body.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/hmacauth"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./hmacauth/`.

```go
secret := []byte("shared-secret")

	app := express.New()
	app.Use(hmacauth.New(hmacauth.Options{Secret: secret}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	send := func(body, sig string) int {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r.Header.Set("X-Signature", sig)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	body := "payload"
	fmt.Println("valid:", send(body, hmacauth.Sign(secret, []byte(body))))
	fmt.Println("invalid:", send(body, "deadbeef"))
```

```
valid: 200
invalid: 401
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that computes the HMAC-SHA256 of the raw request body and compares it, in constant time, to the hex value supplied in the… |
| `func Sign(secret, body []byte) string` | Sign returns the hex-encoded HMAC-SHA256 of body using secret. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the HMAC authentication middleware. |

### Constants

`DefaultMaxBodyBytes`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/hmacauth).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
