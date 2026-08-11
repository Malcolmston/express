# realip — provides middleware that determines the true client IP address from the X-Forwarded-For

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/realip.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/realip)

Package realip provides middleware that determines the true client IP address
from the X-Forwarded-For and X-Real-IP headers, taking an optional list of
trusted proxies into account, and rewrites req.Raw.RemoteAddr to the resolved
value. It ports the "trust proxy" resolution that Express performs on req.ip
together with the header-walking logic of the Node request-ip and proxy-addr
packages, collapsing them into a single handler that leaves the rest of the
stack seeing the originating client rather than the nearest hop.

Use it whenever the application sits behind a reverse proxy, load balancer,
CDN, or ingress controller that terminates the client connection and forwards
the real address in a header. Without it, req.Raw.RemoteAddr — and therefore
anything derived from it such as req.IP(), logging, and IP-keyed rate limiting
— would report the proxy's address instead of the end user's. Placing this
middleware first lets every downstream consumer read a corrected RemoteAddr
transparently, which is exactly what the RewritesRemoteAddr test asserts.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/realip"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./realip/`.

```go
app := express.New()
	app.Use(realip.New(realip.Options{
		TrustedProxies: []string{"5.6.7.8"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Printf("client: %s\n", realip.ClientIP(req))
		fmt.Printf("remote: %s\n", req.Raw.RemoteAddr)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	app.ServeHTTP(rec, req)
```

```
client: 1.2.3.4
remote: 1.2.3.4
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func ClientIP(req *express.Request) string` | ClientIP returns the resolved client IP for the request. |
| `func New(opts ...Options) express.Handler` | New returns middleware that resolves the client IP from forwarding headers and stores it on the request, also updating req.Raw.RemoteAddr. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the real-IP middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/realip).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
