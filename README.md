# express

[![Go Test](https://github.com/Malcolmston/express/actions/workflows/go-test.yml/badge.svg)](https://github.com/Malcolmston/express/actions/workflows/go-test.yml)
[![Go Lint](https://github.com/Malcolmston/express/actions/workflows/go-lint.yml/badge.svg)](https://github.com/Malcolmston/express/actions/workflows/go-lint.yml)
[![Go Vuln](https://github.com/Malcolmston/express/actions/workflows/go-vuln.yml/badge.svg)](https://github.com/Malcolmston/express/actions/workflows/go-vuln.yml)
[![Web Unit](https://github.com/Malcolmston/express/actions/workflows/web-unit.yml/badge.svg)](https://github.com/Malcolmston/express/actions/workflows/web-unit.yml)
[![Web E2E](https://github.com/Malcolmston/express/actions/workflows/web-e2e.yml/badge.svg)](https://github.com/Malcolmston/express/actions/workflows/web-e2e.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express.svg)](https://pkg.go.dev/github.com/malcolmston/express)
[![Go Report Card](https://goreportcard.com/badge/github.com/malcolmston/express)](https://goreportcard.com/report/github.com/malcolmston/express)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Malcolmston/express)](go.mod)
[![Release](https://img.shields.io/github/v/release/Malcolmston/express?sort=semver)](https://github.com/Malcolmston/express/releases)
[![Last Commit](https://img.shields.io/github/last-commit/Malcolmston/express)](https://github.com/Malcolmston/express/commits)
[![Code Size](https://img.shields.io/github/languages/code-size/Malcolmston/express)](https://github.com/Malcolmston/express)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Docs](https://img.shields.io/badge/docs-pages-2f9bff)](https://malcolmston.github.io/express/)

**Node's Express, for Go.**

`express` is a fast, minimalist web framework for Go modeled after
[Express.js](https://expressjs.com/). It gives you the familiar Express routing
API — `app.Get`, `app.Post`, `app.Use`, route parameters, middleware chains,
mountable routers — and chainable request/response helpers, all built on top of
the standard library's `net/http`.

```go
package main

import (
	"log"

	"github.com/malcolmston/express"
)

func main() {
	app := express.New()

	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("Hello World")
	})

	log.Fatal(app.Listen(":3000"))
}
```

## Install

```sh
go get github.com/malcolmston/express
```

## Concepts

### The handler signature

Every handler has the same three-argument shape as Express:

```go
func(req *express.Request, res *express.Response, next express.Next)
```

Call `next()` to pass control to the next matching handler, or `next(err)` to
jump to error-handling middleware.

### Routing

```go
app.Get("/users/:id", handler)     // GET with a route parameter
app.Post("/users", handler)        // POST
app.Put("/users/:id", handler)
app.Delete("/users/:id", handler)
app.Patch("/users/:id", handler)
app.All("/health", handler)        // any method
app.Query("/search", handler)      // the new HTTP QUERY method (safe, with a body)
```

The `QUERY` method is the emerging IETF safe-with-body method; `app.Query`
mirrors Express's `app.query()`.

Route parameters are read with `req.Params("id")`. A `*` in a path is a
wildcard captured as the `*` parameter. Parameters can be **optional** (`:id?`)
or constrained by a **regular expression** (`:id(\d+)`):

```go
app.Get("/users/:id?", handler)      // matches /users and /users/42
app.Get(`/items/:id(\d+)`, handler)  // matches /items/42, not /items/abc
```

Register method handlers for one path with `app.Route`, and preprocess a
parameter with `app.Param`:

```go
app.Route("/users").Get(list).Post(create)
app.Param("id", func(req *express.Request, res *express.Response, next express.Next, id string) {
	user, err := db.Find(id)
	if err != nil { next(err); return }
	req.Set("user", user)
	next()
})
```

### Middleware

```go
app.Use(express.Logger())          // app-wide middleware
app.Use(express.Recover())         // recover from panics -> 500
app.Use(express.JSON())            // parse JSON bodies into req.Body()
app.Use("/admin", requireAuth)     // middleware scoped to a path prefix
```

Middleware runs in registration order. Any handler may short-circuit by
writing a response and simply not calling `next()`.

### Mountable routers

```go
api := express.NewRouter()
api.Get("/users/:id", getUser)
api.Post("/users", createUser)

app.Use("/api", api)   // routes become /api/users/:id, /api/users
```

Routers accept options and can be mounted at a parameterized path; use
`MergeParams` so the sub-router sees the parent's captured params:

```go
users := express.NewRouter(express.RouterOptions{
	MergeParams:   true,  // inherit parent params (e.g. :userId)
	CaseSensitive: false, // "/Foo" == "/foo" (default)
	Strict:        false, // "/foo" == "/foo/" (default)
})
users.Get("/profile", func(req *express.Request, res *express.Response, next express.Next) {
	res.Send("profile of " + req.Params("userId"))
})
app.Use("/users/:userId", users) // GET /users/42/profile -> "profile of 42"
```

Routers nest arbitrarily, each with its own middleware, params, and options.

### Error handling

Register a four-argument error handler with `Use`; it runs only when an
upstream handler calls `next(err)`:

```go
app.Use(func(err error, req *express.Request, res *express.Response, next express.Next) {
	res.Status(500).JSON(map[string]string{"error": err.Error()})
})
```

## Request helpers

| Method | Description |
| ------ | ----------- |
| `req.Params(name)` | route parameter |
| `req.SetPath(p)` | rewrite the path *and* the router's match path (re-routes) |
| `req.Query(name)` | query-string value |
| `req.Get(field)` | request header |
| `req.Body()` | parsed body (after a body-parser middleware) |
| `req.BodyJSON(&dst)` | read + unmarshal the JSON body into `dst` |
| `req.Is("json")` | content-type test |
| `req.Cookie(name)` | read a cookie |
| `req.IP()`, `req.Hostname()`, `req.Protocol()`, `req.Secure()` | connection info |

## Response helpers

| Method | Description |
| ------ | ----------- |
| `res.Status(code)` | set the status code (chainable) |
| `res.Send(body)` | send a string, `[]byte`, or JSON-serializable value |
| `res.JSON(v)` | send `v` as JSON |
| `res.SendStatus(code)` | send the status text for `code` |
| `res.Set(field, value)` | set a header |
| `res.Type(t)` | set Content-Type (`"json"`, `"html"`, ...) |
| `res.Redirect(url)` / `res.Redirect(code, url)` | redirect |
| `res.Cookie(name, value, opts)` | set a cookie |
| `res.SendFile(path)` | send a file (Range + conditional GET support) |
| `res.Download(path, name)` / `res.Attachment(name)` | send as a download |
| `res.Render(view, data)` | render a template (see Views) |
| `res.Format(map)` | content negotiation by Accept |
| `res.ETag(tag)` / `res.LastModified(t)` / `res.NotModified()` | conditional GET |
| `res.End()` | finish with no body |

Request negotiation helpers: `req.Accepts(...)`, `req.AcceptsLanguages(...)`,
`req.AcceptsCharsets(...)`, `req.AcceptsEncodings(...)`, `req.Ranges(size)`, and
`req.Fresh(res)` / `req.Stale(res)`.

## Views

Register a template engine and render views with `res.Render`. The built-in
engine uses `html/template` (for `.html` / `.tmpl`); plug in any engine with
`app.Engine`.

```go
app.Set("views", "./views")     // template directory (default "views")
app.Set("view engine", "html")  // default extension

app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
	res.Render("index", map[string]any{"Title": "Home"})
})

// Custom engine:
app.Engine(".mustache", func(path string, data any) (string, error) { ... })
```

Most response methods return `*Response` so they can be chained:
`res.Status(201).JSON(user)`.

## Bundled middleware

- `express.JSON()` — parse `application/json` bodies into `req.Body()`.
- `express.URLEncoded()` — parse form-encoded bodies.
- `express.Text()` — parse `text/plain` bodies into `req.Body()`.
- `express.Multipart(maxMemory)` — parse `multipart/form-data` (file uploads).
- `express.Static(root)` — serve static files from a directory.
- `express.Session(opts...)` — cookie-backed sessions (see below).
- `express.Logger()` — log method, path, status, and duration.
- `express.Recover()` — recover from panics and return a 500.

## Middleware suite (100+ packages)

Beyond the bundled middleware above, `express` ships a large catalog of
ready-to-use middleware under [`middleware/`](middleware/) — over 100 independent
subpackages spanning security headers, authentication/access control, body &
response transforms, rate limiting & traffic control, routing/static helpers,
cookies/CSRF/sessions, and dev utilities. Each has a `New(...)` constructor and
is standard-library only.

```go
import (
	"github.com/malcolmston/express/middleware/cors"
	"github.com/malcolmston/express/middleware/helmet"
	"github.com/malcolmston/express/middleware/ratelimit"
	"github.com/malcolmston/express/middleware/compression"
)

app.Use(helmet.New())
app.Use(cors.New(cors.Options{AllowOrigins: []string{"https://example.com"}}))
app.Use(compression.New())
app.Use(ratelimit.New(ratelimit.Options{Max: 100, Window: time.Minute}))
```

See [MIDDLEWARE.md](MIDDLEWARE.md) for the full catalog.

## Sessions

`express.Session()` adds a cookie-backed session, persisted through a pluggable
`SessionStore` (an in-memory store is the default). Read and write it with
`req.Session()`; changes are saved automatically just before the response is
sent.

```go
app.Use(express.Session(express.SessionOptions{
	Name:   "sid",
	Secure: true, // HTTPS only
	// Store: myRedisStore,
}))

app.Post("/login", func(req *express.Request, res *express.Response, next express.Next) {
	sess := req.Session()
	sess.Regenerate()          // new id on privilege change (anti-fixation)
	sess.Set("userID", "42")
	res.Send("logged in")
})

app.Get("/me", func(req *express.Request, res *express.Response, next express.Next) {
	res.Send("user " + req.Session().GetString("userID"))
})

app.Post("/logout", func(req *express.Request, res *express.Response, next express.Next) {
	req.Session().Destroy()
	res.Send("bye")
})
```

## File uploads & forms

```go
app.Use(express.Multipart(0)) // 0 = 32 MiB default in-memory buffer

app.Post("/avatar", func(req *express.Request, res *express.Response, next express.Next) {
	file, header, err := req.FormFile("avatar")
	if err != nil {
		next(err)
		return
	}
	defer file.Close()
	res.JSON(map[string]any{"filename": header.Filename, "caption": req.FormValue("caption")})
})
```

`req.Form()` returns all form values (query + body) as `url.Values`; `req.Files(name)`
returns every uploaded file header for a field.

## Input validation

The [`validator`](validator/) subpackage provides fluent request validation in
the spirit of `express-validator`. Build a `Schema` and mount it as middleware
that rejects invalid requests with a `400` JSON body — or call `Validate` on a
map directly.

```go
import "github.com/malcolmston/express/validator"

schema := validator.Schema{
	validator.Field("email").Required().Email(),
	validator.Field("age").Optional().IsInt().Min(0).Max(120),
	validator.Field("name").Required().MinLen(2).MaxLen(50),
	validator.Field("role").Required().In("admin", "user"),
}

app.Use(express.JSON())
app.Post("/users", schema.Body(), createUser) // 400 {"errors":[...]} on failure
```

Available rules: `Required`, `Optional`, `Email`, `MinLen`, `MaxLen`, `Min`,
`Max`, `IsInt`, `IsNumber`, `In`, `Matches`, and `Custom`. Use `schema.Query()`
to validate the query string instead of the body.

## Streaming & chunked responses

The response supports incremental, flushed output for large or open-ended
bodies. `*Response` implements `io.Writer`, so it drops into `io.Copy` and
`fmt.Fprintf`.

```go
// Stream a body chunk-by-chunk (each write is flushed to the client).
app.Get("/stream", func(req *express.Request, res *express.Response, next express.Next) {
	res.Type("text").Stream(func(w io.Writer) error {
		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "line %d\n", i)
		}
		return nil
	})
})

// Copy a large reader to the client in chunks without buffering it all.
res.SendStream(file)               // default 32 KiB chunks
res.SendChunked(bigBytes, 64<<10)  // fixed-size chunks from memory
res.WriteChunk([]byte("partial"))  // low-level: write + flush
res.Flush()                        // flush buffered data
```

### Server-Sent Events

```go
app.Get("/events", func(req *express.Request, res *express.Response, next express.Next) {
	sse := res.SSE() // sets text/event-stream headers and flushes them
	for {
		select {
		case <-req.Raw.Context().Done():
			return
		case ev := <-updates:
			sse.SendJSON("update", ev)   // event: update\ndata: {...}
		case <-ticker.C:
			sse.Comment("keep-alive")
		}
	}
})
```

`SSEWriter` provides `Send`, `SendData`, `SendJSON`, `SendID` (for
`Last-Event-ID` resumption), `Comment`, and `Retry`.

## API documentation (`app.Docs`)

`app.Docs()` introspects every route you have registered — including those on
mounted sub-routers — and serves live, standards-compliant API documentation
with no code generation step and no third-party dependencies. One call mounts an
OpenAPI 3.1 spec, interactive Swagger UI and ReDoc pages, a YAML rendering, an
AsyncAPI 2.6 document for socket/event channels, and a Postman collection.

```go
app := express.New()

app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Get("/users/:id", getUser)

// Introspection knows the method, path and :id parameter automatically.
// Describe adds the parts it can't infer.
app.Describe("POST", "/users", express.RouteDoc{
	Summary: "Create a user",
	Tags:    []string{"users"},
	RequestBody: &express.BodyDoc{
		Required: true,
		Schema:   map[string]any{"type": "object", "required": []any{"name"}},
	},
	Responses: map[string]express.ResponseDoc{
		"201": {Description: "Created", Schema: map[string]any{"type": "object"}},
	},
})

// Document socket/event channels for the AsyncAPI spec.
app.Channel("chat.message", express.ChannelDoc{
	Description: "Live chat messages",
	Subscribe:   &express.MessageDoc{Name: "messageReceived", Payload: map[string]any{"type": "object"}},
})

app.Docs(express.DocsOptions{
	Title:   "My API",
	Version: "1.0.0",
	Servers: []string{"https://api.example.com"},
})
```

This mounts, by default:

| Path             | Content                                         |
| ---------------- | ----------------------------------------------- |
| `/docs`          | Swagger UI                                       |
| `/redoc`         | ReDoc                                            |
| `/openapi.json`  | OpenAPI 3.1 specification (JSON)                  |
| `/openapi.yaml`  | OpenAPI 3.1 specification (YAML)                  |
| `/asyncapi.json` | AsyncAPI 2.6 document (event/socket channels)    |
| `/postman.json`  | Postman v2.1 collection                          |

Every path is configurable via `DocsOptions` (set any to `"-"` to disable), and
an `Enrich` hook can customise each generated operation programmatically. The
specs are rebuilt per request, so routes registered after `Docs()` still appear.
You can also obtain the documents directly — `app.Routes()`, `app.OpenAPI()`,
`app.OpenAPIYAML()`, `app.AsyncAPI()`, `app.PostmanCollection()` — to serve or
persist them however you like.

## Using with net/http

An `*express.Application` is an `http.Handler`, so it drops into anything that
speaks `net/http` — `http.Server`, `httptest`, or other middleware:

```go
srv := &http.Server{Addr: ":8080", Handler: app}
srv.ListenAndServe()
```

## Example

A runnable example lives in [`examples/basic`](examples/basic/main.go):

```sh
go run ./examples/basic
```

## Compatibility

This is a Go re-implementation modeled on Express.js, targeting API/behavior
parity and standards-compliant HTTP output. See [COMPATIBILITY.md](COMPATIBILITY.md)
for a feature-by-feature parity table and known gaps.

## Companion library

Pair this with [`passport`](https://github.com/malcolmston/passport) — a Go port
of Passport.js — for pluggable authentication.

## License

[MIT](LICENSE)
