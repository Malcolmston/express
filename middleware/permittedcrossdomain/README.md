# permittedcrossdomain — provides middleware that sets the X-Permitted-Cross-Domain-Policies response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/permittedcrossdomain.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/permittedcrossdomain)

Package permittedcrossdomain provides middleware that sets the
X-Permitted-Cross-Domain-Policies response header on every response,
controlling whether Adobe clients honor cross-domain policy files served by
the site. It is a port of Helmet's crossOriginResourcePolicy sibling — the
permittedCrossDomainPolicies middleware — which emits this single header
with a configurable policy value.

The header governs legacy Adobe technologies (Flash Player and Acrobat) that
look for a crossdomain.xml or similar policy file to decide whether they may
load data from the site on behalf of another origin. Setting the header to
"none" instructs those clients that no policy files are permitted anywhere on
the domain, closing a historical cross-origin data-theft vector. Use it as a
low-cost hardening measure on any site; it is harmless on modern browsers,
which ignore the header, and meaningful only to the shrinking set of Adobe
runtimes that still consult it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/permittedcrossdomain"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./permittedcrossdomain/`.

```go
app := express.New()
	app.Use(permittedcrossdomain.New(permittedcrossdomain.Options{
		Policy: "master-only",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Permitted-Cross-Domain-Policies"))
```

```
master-only
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the X-Permitted-Cross-Domain-Policies header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the permittedcrossdomain middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/permittedcrossdomain).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
