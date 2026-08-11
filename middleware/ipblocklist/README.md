# ipblocklist — provides middleware that rejects requests whose client IP address matches a configured

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/ipblocklist.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/ipblocklist)

Package ipblocklist provides middleware that rejects requests whose client IP
address matches a configured blocklist of exact addresses or CIDR ranges,
responding with 403 Forbidden, and lets every other request through. It
mirrors the Node "express-ipfilter" package in its deny/blacklist mode and the
common pattern of an IP-based denylist, reimplemented here on top of Go's
standard net package with no third-party dependencies.

Use this middleware when the default should be open access but a specific set
of clients must be turned away — banning abusive addresses, blocking known
bad networks or scanners, or dropping traffic from ranges you never intend to
serve. It is a negative/deny model: anything not on the list is allowed, which
is the complement of the sibling ipallowlist package. Choose blocklist when
the exceptions are the untrusted parties; choose allowlist when the exceptions
are the trusted ones.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/ipblocklist"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./ipblocklist/`.

```go
app := express.New()
	app.Use(ipblocklist.New(ipblocklist.Options{
		Block: []string{"1.2.3.4", "10.0.0.0/8"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	call := func(clientIP string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = clientIP + ":1234"
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, r)
		return rec.Code
	}

	fmt.Println(call("10.5.5.5"))
	fmt.Println(call("8.8.8.8"))
```

```
403
200
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 when the request's client IP matches one of the configured blocklist entries, and otherwise calls next. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the IP blocklist middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/ipblocklist).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
