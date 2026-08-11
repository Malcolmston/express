# subdomain — provides middleware that extracts the subdomain portion of the request's hostname

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/subdomain.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/subdomain)

Package subdomain provides middleware that extracts the subdomain portion of
the request's hostname and stores it on the request for downstream handlers
under the key "subdomain". It is the express framework's Go analogue of
Express's built-in req.subdomains array and its "subdomain offset" setting,
packaged as an explicit middleware so multi-tenant routing can key off the
host without every handler re-parsing req.Hostname itself.

Reach for this middleware when a single application serves many tenants or
sections addressed by host — acme.example.com and globex.example.com routed
to different customers, or api.example.com, www.example.com, and
admin.example.com dispatched to different handler trees. Running it once at
the front of the chain computes the subdomain a single time and hands every
downstream handler a ready-made value to switch on, avoiding duplicated and
inconsistent host parsing scattered through the codebase.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/subdomain"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./subdomain/`.

```go
app := express.New()
	app.Use(subdomain.New(subdomain.Options{BaseHost: "example.com"}))

	var tenant string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(subdomain.Key); ok {
			tenant, _ = v.(string)
		}
		res.Send("ok")
	})

	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://api.example.com/", nil))

	fmt.Println(tenant)
```

```
api
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that computes the subdomain and stores it via req.Set(Key, sub). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the subdomain middleware. |

### Constants

`Key`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/subdomain).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
