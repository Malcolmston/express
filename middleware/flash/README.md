# flash — implements one-time "flash" messages backed by the express session

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/flash.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/flash)

Package flash implements one-time "flash" messages backed by the express
session. It is a stdlib-only port of the Node "connect-flash" middleware (the
flash helper popularized by Express and Connect): a message is stored under a
reserved session key during one request/response cycle and then read — and
atomically cleared — on a subsequent one. The classic use is the
post/redirect/get flow, where a handler that mutates state records a "saved"
or "invalid input" notice, issues a redirect, and the following GET renders
and consumes that notice exactly once.

Reach for flash messages whenever you need to carry a short-lived, user-facing
status string across a redirect without threading it through the URL or a
query parameter. Because the data lives in the session it survives the
redirect but is scoped to the one browser session, and because Get deletes it
after reading, a refresh of the destination page will not show the message a
second time. Messages are grouped by a free-form category string (for example
"error", "info", or "success") so templates can style them differently.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/flash"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./flash/`.

```go
app := express.New()
	app.Use(express.Session())
	app.Use(flash.New())

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		flash.Add(req, "info", "profile saved")
		flash.Add(req, "error", "avatar too large")

		for _, m := range flash.Get(req) {
			fmt.Printf("%s: %s\n", m.Category, m.Message)
		}
		fmt.Println("second read empty:", flash.Get(req) == nil)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
```

```
info: profile saved
error: avatar too large
second read empty: true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Add(req *express.Request, category, msg string)` | Add appends a flash message under the given category to the session. |
| `func New() express.Handler` | New returns middleware that enables flash-message helpers for downstream handlers. |

### Types

| Type | What it is |
| --- | --- |
| `Message` | Message is a single flash message with a category (e.g. |

<details>
<summary><code>Message</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Get(req *express.Request) []Message` | Get returns all pending flash messages and clears them from the session. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/flash).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
