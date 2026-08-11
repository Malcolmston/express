# xssfilter — provides express middleware that sets the legacy X-XSS-Protection response header

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/xssfilter.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/xssfilter)

Package xssfilter provides express middleware that sets the legacy
X-XSS-Protection response header on every response. It is the express analogue
of Helmet's xssFilter/X-XSS-Protection helper for Node, packaged as a small
middleware so an application can control this one header in a single place at
the front of its chain.

X-XSS-Protection was a header that instructed a browser's built-in reflected-
XSS auditor how to behave. Modern security guidance — and the default here
— is to turn that auditor off by sending the value "0". The heuristic
auditors in older browsers were themselves a source of vulnerabilities and
subtle content-corruption bugs, every major engine has since removed the
feature, and real cross-site-scripting defense belongs in a
Content-Security-Policy and in correct output encoding. Emitting "0" documents
that intent explicitly and suppresses any residual legacy behavior.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/xssfilter"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./xssfilter/`.

```go
app := express.New()
	app.Use(xssfilter.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))

	fmt.Println(rec.Header().Get("X-XSS-Protection"))
```

```
0
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the X-XSS-Protection header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the xssfilter middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/xssfilter).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
