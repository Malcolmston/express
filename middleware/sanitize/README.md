# sanitize — provides middleware that strips HTML tags from every query-string value on the incoming

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/sanitize.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/sanitize)

Package sanitize provides middleware that strips HTML tags from every
query-string value on the incoming request, mitigating trivial reflected
cross-site-scripting (XSS) vectors that flow through query parameters. It is
the express framework's Go analogue of the input-cleaning middlewares Node
developers reach for — such as express-sanitizer, sanitize-html wired into a
middleware, or the deprecated express-validator sanitizers — reduced to a
single, dependency-free tag-stripping pass over the request's query.

Reach for this middleware when query parameters are reflected back into HTML
responses, log lines, or templates and you want a cheap, always-on backstop
against markup injection. It is best understood as defence in depth rather
than a primary control: it removes the angle-bracketed tags that carry most
naive <script> and event-handler payloads before your handlers ever see the
value, so a parameter echoed straight into a page cannot smuggle a live
element. It should sit alongside, not instead of, proper contextual output
encoding at render time.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/sanitize"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./sanitize/`.

```go
app := express.New()
	app.Use(sanitize.New())

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("Hello, " + req.Query("name"))
	})

	r := httptest.NewRequest("GET", "/?name=%3Cscript%3Ealert(1)%3C%2Fscript%3EAnn", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
```

```
Hello, alert(1)Ann
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New() express.Handler` | New returns middleware that removes HTML tags from every query-string value in place, rewriting the request's RawQuery so downstream handlers see the… |
| `func StripTags(s string) string` | StripTags removes anything that looks like an HTML tag from s. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/sanitize).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
