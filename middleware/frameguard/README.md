# frameguard — provides middleware that sets the X-Frame-Options response header to control whether

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/frameguard.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/frameguard)

Package frameguard provides middleware that sets the X-Frame-Options response
header to control whether the page may be rendered inside a <frame>, <iframe>,
<embed>, or <object>, helping to mitigate clickjacking attacks. It is a
stdlib-only port of Helmet's "frameguard" module (historically the standalone
"frameguard" npm package), which is one of the header-setting middlewares
Helmet enables by default.

Clickjacking tricks a user into interacting with your site while it is framed
invisibly beneath attacker-controlled UI. Setting X-Frame-Options tells the
browser to refuse to render the response inside a frame from another origin,
defeating the attack for legacy browsers. Use this middleware on any HTML
endpoint that does not need to be embedded cross-origin. Modern browsers also
honor the frame-ancestors directive of Content-Security-Policy, which
supersedes X-Frame-Options; frameguard remains valuable for older clients and
as defense in depth, and the two can be sent together.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/frameguard"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./frameguard/`.

```go
app := express.New()
	app.Use(frameguard.New(frameguard.Options{Action: "deny"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println("X-Frame-Options:", rec.Header().Get("X-Frame-Options"))
```

```
X-Frame-Options: DENY
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns middleware that sets the X-Frame-Options header. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the frameguard middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/frameguard).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
