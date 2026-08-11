# nonce — provides middleware that generates a fresh, random

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/nonce.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/nonce)

Package nonce provides middleware that generates a fresh, random,
base64-encoded nonce for each request. It mirrors the Node helpers built
around Content-Security-Policy nonces — for example express's res.locals
nonce pattern and the cspNonce/helmet-csp integration used to authorize
specific inline scripts and styles — by producing one unguessable token per
request that downstream code can attach to a CSP header and to markup.

Reach for it whenever you need a per-request nonce, most commonly to allow a
handful of trusted inline <script>/<style> blocks under a strict CSP without
resorting to 'unsafe-inline'. Because a nonce is only safe if it is fresh and
unpredictable for every response, this middleware regenerates it on each
request rather than reusing a value; templates and a CSP builder then read the
same token so the header and the markup agree.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/nonce"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./nonce/`.

```go
app := express.New()
	app.Use(nonce.New(nonce.Options{Bytes: 16}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fromReq := nonce.Nonce(req)
		fromLocals, _ := res.Locals[nonce.ContextKey].(string)
		fmt.Println(fromReq != "" && fromReq == fromLocals)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
```

```
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Generate(n int) string` | Generate returns a random base64-encoded nonce built from n random bytes. |
| `func New(opts ...Options) express.Handler` | New returns middleware that stores a per-request random nonce via req.Set("nonce", n) and res.Locals["nonce"]. |
| `func Nonce(req *express.Request) string` | Nonce returns the nonce generated for the request, or "" if the middleware did not run. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the nonce middleware. |

### Constants

`ContextKey`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/nonce).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
