# express

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
```

Route parameters are read with `req.Params("id")`. A `*` in a path is a
wildcard captured as the `*` parameter.

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
| `res.End()` | finish with no body |

Most response methods return `*Response` so they can be chained:
`res.Status(201).JSON(user)`.

## Bundled middleware

- `express.JSON()` — parse `application/json` bodies into `req.Body()`.
- `express.URLEncoded()` — parse form-encoded bodies.
- `express.Static(root)` — serve static files from a directory.
- `express.Logger()` — log method, path, status, and duration.
- `express.Recover()` — recover from panics and return a 500.

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

## Companion library

Pair this with [`passport`](https://github.com/malcolmston/passport) — a Go port
of Passport.js — for pluggable authentication.

## License

[MIT](LICENSE)
