# tokenheader — provides generic middleware that validates an opaque token carried in a single

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/tokenheader.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/tokenheader)

Package tokenheader provides generic middleware that validates an opaque token
carried in a single, configurable request header. It is the express
framework's Go take on the family of Node "API key" / "shared secret" header
guards -- the hand-rolled middleware people write around headers such as
X-API-Key, X-Auth-Token, or X-Session-Token, and the header mode of packages
like express-api-key-auth -- packaged as a drop-in express.Handler. It is a
close sibling of the bearerauth middleware, differing chiefly in that the
header name is caller-chosen and no scheme prefix is parsed.

Reach for this middleware when your credential does not live in the standard
Authorization: Bearer form but in a custom header instead: internal
service-to-service calls that pass a shared secret, webhook receivers that
authenticate with a signing token in a vendor-specific header, or simple
API-key gates on machine-consumed endpoints. Because the token is opaque to
the middleware, it works equally well with random API keys, session
identifiers, or signed tokens; the meaning and validity of the token are
defined entirely by the caller-supplied Verify callback.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/tokenheader"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./tokenheader/`.

```go
app := express.New()
	app.Use(tokenheader.New(tokenheader.Options{
		Header: "X-API-Key",
		Verify: func(token string) bool { return token == "s3cr3t" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	call := func(key string) {
		r := httptest.NewRequest("GET", "/", nil)
		if key != "" {
			r.Header.Set("X-API-Key", key)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("key=%-8q -> %d %s\n", key, w.Code, w.Body.String())
	}
	call("s3cr3t")
	call("nope")
```

```
key="s3cr3t" -> 200 welcome
key="nope"   -> 401 Unauthorized
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that reads the configured header and rejects the request with 401 when the token is missing or fails verification. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the token-header middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/tokenheader).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
