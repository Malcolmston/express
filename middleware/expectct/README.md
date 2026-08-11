# expectct — provides express middleware that sets the Expect-CT response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/expectct.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/expectct)

Package expectct provides express middleware that sets the Expect-CT response
header, allowing sites to opt in to reporting and/or enforcement of
Certificate Transparency requirements. It ports Helmet's expect-ct middleware,
emitting the same max-age / enforce / report-uri directive set that
helmet.expectCt() produced.

Historically this header let a site instruct browsers to expect valid Signed
Certificate Timestamps on its TLS certificates, reporting or rejecting
connections that lacked them. Note that Expect-CT is effectively obsolete:
Certificate Transparency is now enforced by default in modern browsers and the
header is ignored by current versions, which is why Helmet itself removed it.
This middleware exists for parity and for serving legacy clients; new
deployments generally do not need it.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/expectct"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./expectct/`.

```go
app := express.New()
	app.Use(expectct.New(expectct.Options{
		MaxAge:    86400,
		Enforce:   true,
		ReportURI: "https://example.com/report",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("Expect-CT"))
```

```
max-age=86400, enforce, report-uri="https://example.com/report"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the Expect-CT header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the expectct middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/expectct).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
