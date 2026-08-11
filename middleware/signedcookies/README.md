# signedcookies — provides middleware that verifies a tamper-evident

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/signedcookies.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/signedcookies)

Package signedcookies provides middleware that verifies a tamper-evident,
HMAC-signed cookie and exposes its plaintext value to downstream handlers. It
is the express framework's Go analogue of the signed-cookie support built into
Express via the cookie-parser middleware (expressjs/cookie-parser) and the
underlying cookie-signature module (tj/node-cookie-signature): a value is
stored on the client alongside a keyed HMAC so the server can detect any
modification without keeping server-side session state.

Reach for this middleware when you place a small, non-secret identifier in a
cookie — a user id, a tenant slug, a feature flag — and need to trust it
on the way back in without a database round trip. The signature guarantees
integrity and authenticity (only a holder of Secret could have produced it),
so a client cannot forge or edit the value. It does not provide
confidentiality: the plaintext value is visible in the cookie, so never place
a password or other secret in it; encrypt separately if the payload must be
hidden.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/signedcookies"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./signedcookies/`.

```go
secret := []byte("s3cr3t")

	app := express.New()
	app.Use(signedcookies.New(signedcookies.Options{
		Secret: secret,
		Name:   "uid",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		v, _ := req.Value("uid")
		res.Send("hello " + v.(string))
	})

	good := httptest.NewRequest(http.MethodGet, "/", nil)
	good.AddCookie(&http.Cookie{Name: "uid", Value: signedcookies.Sign(secret, "alice")})
	gw := httptest.NewRecorder()
	app.ServeHTTP(gw, good)
	fmt.Printf("valid:    %d %s\n", gw.Code, gw.Body.String())

	bad := httptest.NewRequest(http.MethodGet, "/", nil)
	bad.AddCookie(&http.Cookie{Name: "uid", Value: signedcookies.Sign(secret, "alice") + "x"})
	bw := httptest.NewRecorder()
	app.ServeHTTP(bw, bad)
	fmt.Printf("tampered: %d %s\n", bw.Code, bw.Body.String())
```

```
valid:    200 hello alice
tampered: 401 Unauthorized
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that verifies the configured signed cookie. |
| `func Sign(secret []byte, value string) string` | Sign returns a signed cookie value of the form "value.hexHMAC" produced from value and secret. |
| `func Verify(secret []byte, signed string) (string, bool)` | Verify checks a signed cookie value and returns the embedded value when the signature is valid. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the signed-cookie middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/signedcookies).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
