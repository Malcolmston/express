# express — Overview

`express` is a fast, minimalist web framework for Go, modeled after
[Express.js](https://expressjs.com/). It reproduces the familiar Express routing
and middleware model on top of the Go standard library's `net/http`, using only
the standard library and no third-party dependencies.

This document explains how the framework works, how to use it, and how it
compares to its Node.js predecessor.

---

## How it works

### The handler/middleware model

Everything in express is a **handler** with the same three-argument shape as
Express:

```go
type Handler func(req *Request, res *Response, next Next)
```

- `req` wraps the incoming `*http.Request` with chainable accessors
  (`req.Params`, `req.Query`, `req.Body`, `req.Get`, ...). The raw request is
  always reachable as `req.Raw`.
- `res` wraps the `http.ResponseWriter` with chainable helpers (`res.Status`,
  `res.JSON`, `res.Send`, `res.Render`, ...). The raw writer is `res.Writer`.
- `next` advances the chain: call `next()` to run the next matching handler, or
  `next(err)` to skip to error-handling middleware.

Handlers registered with `Use` are **middleware**; handlers registered with a
method (`Get`, `Post`, ...) are **route handlers**. They are the same type and
run in registration order, forming a single ordered chain per request. Any
handler may short-circuit the chain simply by writing a response and not calling
`next()`.

Error handling uses a four-argument variant, mounted with `Use`, that runs only
when an upstream handler calls `next(err)`:

```go
func(err error, req *Request, res *Response, next Next)
```

### Routing with path parameters

Routes are registered on an `*Application` or a standalone `*Router` (an
`Application` embeds a `*Router`, so both expose the identical registration
surface described by the `RouteRegistrar` interface):

```go
app.Get("/users/:id", handler)      // named parameter -> req.Params("id")
app.Get("/users/:id?", handler)     // optional parameter (matches /users too)
app.Get(`/items/:id(\d+)`, handler) // regex-constrained parameter
app.Get("/files/*", handler)        // wildcard captured as the "*" parameter
app.All("/health", handler)         // any HTTP method
```

Routers are **mountable and nestable**. Mounting a sub-router under a prefix
rewrites its routes relative to that prefix, and `RouterOptions{MergeParams:
true}` lets a sub-router see parameters captured by its parent:

```go
api := express.NewRouter()
api.Get("/users/:id", getUser)
app.Use("/api", api) // exposes GET /api/users/:id
```

`app.Param(name, fn)` registers a callback that preprocesses a captured
parameter (e.g. loading a record) before the route handler runs, and
`app.Route(path)` groups several methods on a single path.

### The http.Handler surface

`*Application` implements `http.Handler`. Its `ServeHTTP` constructs the
`*Request`/`*Response` pair and drives the router's handler chain. Because the
app *is* a standard handler, it drops into anything that speaks `net/http`
— `http.Server`, `httptest`, TLS servers, or other middleware:

```go
srv := &http.Server{Addr: ":8080", Handler: app}
srv.ListenAndServe()
```

Conversely, `express.WrapHandler` / `express.WrapHandlerFunc` adapt a plain
`http.Handler` *into* an express `Handler`, so existing net/http code (including
connection-hijacking handlers like WebSocket or Socket.IO servers) can be
mounted with `app.Use`.

### Views, negotiation, and streaming

- **Views.** `app.Engine(ext, fn)` registers a template engine keyed by file
  extension; `res.Render(view, data)` renders a named view with data. A built-in
  `html/template` engine is registered for `.html` and `.tmpl` out of the box,
  and the template directory / default extension come from the `"views"` and
  `"view engine"` settings.
- **Content negotiation.** `res.Format(map)` dispatches by the request's
  `Accept` header, and `req.Accepts`, `req.AcceptsLanguages`,
  `req.AcceptsCharsets`, and `req.AcceptsEncodings` pick the best match from a
  list of offers. Conditional-GET helpers (`res.ETag`, `res.LastModified`,
  `res.NotModified`, `req.Fresh`) support caching.
- **Streaming.** `*Response` implements `io.Writer`, so it works with `io.Copy`
  and `fmt.Fprintf`. `res.Stream`, `res.SendStream`, and `res.SendChunked` write
  and flush output incrementally without buffering the whole body, and
  `res.SSE()` returns an `SSEWriter` for Server-Sent Events (`event:`/`data:`
  framing, keep-alive comments, and `Last-Event-ID` resumption).

---

## How to use it

### 1. Hello world

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

### 2. A JSON API with middleware, params, and error handling

```go
package main

import (
	"log"

	"github.com/malcolmston/express"
)

func main() {
	app := express.New()

	app.Use(express.Logger())  // log method, path, status, duration
	app.Use(express.Recover()) // turn panics into 500s
	app.Use(express.JSON())    // parse JSON bodies into req.Body()

	app.Get("/users/:id", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]string{"id": req.Params("id")})
	})

	app.Post("/users", func(req *express.Request, res *express.Response, next express.Next) {
		body, ok := req.Body().(map[string]any)
		if !ok {
			res.Status(400).JSON(map[string]string{"error": "invalid body"})
			return
		}
		res.Status(201).JSON(body)
	})

	// Four-argument error handler: runs only after next(err).
	app.Use(func(err error, req *express.Request, res *express.Response, next express.Next) {
		res.Status(500).JSON(map[string]string{"error": err.Error()})
	})

	log.Fatal(app.Listen(":3000"))
}
```

### 3. Mountable routers plus bundled middleware ports

```go
package main

import (
	"log"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/cors"
	"github.com/malcolmston/express/middleware/ratelimit"
)

func main() {
	app := express.New()
	app.Use(cors.New(cors.Options{AllowedOrigins: []string{"https://example.com"}}))
	app.Use(ratelimit.New(ratelimit.Options{Max: 100, Window: time.Minute}))

	api := express.NewRouter()
	api.Get("/ping", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("pong")
	})
	app.Use("/api", api) // GET /api/ping -> "pong"

	log.Fatal(app.Listen(":3000"))
}
```

> The exact constructor options above (`cors.Options`, `ratelimit.Options`)
> follow each subpackage's `New(...)` signature; see `MIDDLEWARE.md` for the full
> catalog.

---

## Why it's better than its predecessor

This is an honest comparison to Node's `expressjs/express`. The goal is parity of
*developer experience* with the operational and safety advantages of Go — not to
claim Express.js is bad.

- **Single static binary.** `go build` produces one self-contained executable
  with no runtime, no interpreter, and no `node_modules` to ship. Deployment is a
  file copy; there is no separate runtime to install or keep patched.

- **Standard-library-first, tiny dependency graph.** The module declares **zero
  third-party dependencies** — there is no `go.sum`, and every one of its ~190
  packages (the framework plus 100+ middleware/util ports) is built on the Go
  standard library alone. Express.js's minimalism is real, but a typical app
  still pulls a transitive tree of hundreds of npm packages. Here the supply
  chain is effectively the Go standard library.

- **Compile-time type-safe handlers.** Handlers, `req`, and `res` are concrete
  Go types checked at build time. Passing the wrong argument, misspelling a
  method, or returning the wrong shape is a compile error, not a runtime crash
  discovered in production. Express.js's `(req, res, next)` contract is enforced
  only at runtime.

- **Wire-compatible behavior.** The routing semantics (named/optional/regex
  params, wildcards, mountable routers, `MergeParams`, param preprocessing), the
  request/response helper surface, and the HTTP output aim for parity with
  Express.js — so the mental model transfers directly. See `COMPATIBILITY.md` for
  the feature-by-feature table and known gaps.

- **No node_modules supply chain.** Because the dependency set is the standard
  library, there is no `postinstall` script surface, no lockfile drift, and no
  transitive-dependency CVE churn to audit. Fewer moving parts to trust.

- **100+ bundled middleware / utility ports.** Security headers (helmet, CSP,
  CORS), auth/access control, body & response transforms (compression,
  body-parsers), rate limiting and traffic control, cookies/CSRF/sessions, plus a
  large set of utility ports (ids, hashing, string/collection helpers) ship *in
  this module*, each standard-library-only. In the Node ecosystem these are
  separate, independently versioned npm packages.

### Honest tradeoffs

- **Smaller ecosystem and community.** Express.js has a vast catalog of
  third-party middleware and years of accumulated answers, examples, and battle
  testing. This project ports a curated subset; anything outside it you write
  yourself.
- **Maturity.** Express.js is one of the most widely deployed web frameworks in
  existence. This is a re-implementation targeting parity, and some edges/gaps
  remain (documented in `COMPATIBILITY.md`).
- **Language reach.** If your team and stack are JavaScript/TypeScript
  end-to-end, sharing code and hiring around Node may outweigh Go's runtime
  advantages. The choice is contextual, not absolute.
- **Idiomatic differences.** Some Go developers prefer the standard `net/http`
  `Handler`/`ServeMux` style or routers like chi/gin; express deliberately trades
  a little Go idiom for Express.js familiarity.
