# Compatibility

`express` is a Go re-implementation of the Express.js framework. Because Go and
JavaScript are different runtimes, it is not a drop-in for the Node package —
you cannot run Node middleware in it. What it targets is **API and behavior
parity**: the same routing model, the same `(req, res, next)` handler shape, the
same chainable helpers, and standards-compliant HTTP output (status codes,
headers, cookies, JSON) that any HTTP client — including a Node one —
interoperates with.

## Routing parity

| Express.js | This library |
| ---------- | ------------ |
| `app.get/post/put/delete/patch/head/options` | `app.Get/Post/Put/Delete/Patch/Head/Options` |
| `app.all(path, h)` | `app.All(path, h)` |
| `app.use(mw)` / `app.use(path, mw)` | `app.Use(mw)` / `app.Use(path, mw)` |
| `express.Router()` + `app.use(path, router)` | `express.NewRouter()` + `app.Use(path, router)` |
| `:param`, `*` wildcard | `:param`, `*` wildcard (`req.Params`) |
| error middleware `(err, req, res, next)` | `func(err error, req, res, next)` via `Use` |

## Request / Response parity

| Express.js | This library |
| ---------- | ------------ |
| `req.params`, `req.query` | `req.Params(k)`, `req.Query(k)` |
| `req.get(h)`, `req.is(t)` | `req.Get(h)`, `req.Is(t)` |
| `req.body` | `req.Body()` / `req.BodyJSON(&v)` |
| `req.cookies`, `req.ip`, `req.hostname`, `req.protocol`, `req.secure` | `req.Cookie`, `req.IP`, `req.Hostname`, `req.Protocol`, `req.Secure` |
| `res.status().json()/send()` | `res.Status().JSON()/Send()` (chainable) |
| `res.set`, `res.type`, `res.redirect`, `res.cookie`, `res.sendStatus`, `res.end` | same, capitalized |
| `res.locals` | `res.Locals` |

## Middleware parity

| Express.js | This library |
| ---------- | ------------ |
| `express.json()` | `express.JSON()` |
| `express.urlencoded()` | `express.URLEncoded()` |
| `express.text()` | `express.Text()` |
| `express.static(dir)` | `express.Static(dir)` |
| `multer` (uploads) | `express.Multipart()` + `req.FormFile` |
| `express-session` | `express.Session()` (pluggable `SessionStore`) |
| `morgan` (logging) | `express.Logger()` |
| `express-validator` | `validator` subpackage |

## HTTP wire compatibility

The framework is a thin layer over `net/http`, and an `*express.Application` is
an `http.Handler`. All output is standard HTTP:

- `res.JSON` sets `application/json; charset=utf-8` and marshals with
  `encoding/json`.
- `res.Cookie` emits a standard `Set-Cookie` (with `HttpOnly`/`Secure`/
  `SameSite` via `CookieOptions`); `Session()` cookies match `express-session`'s
  default cookie name (`connect.sid`).
- Redirects use the standard `Location` header and 3xx status.

Any client — a browser, `curl`, or a Node HTTP client — interoperates.

## Known gaps

Express.js is large; this port covers the routing/middleware core. Not (yet)
implemented:

- View/template engines (`res.render`, `app.set('view engine', ...)`).
- `res.sendFile` / `res.download` / range requests (use `express.Static` or
  `http.ServeFile` directly).
- Content negotiation helpers (`res.format`, `req.accepts`).
- ETag/`res.fresh`/conditional-GET helpers.
- Trust-proxy configuration nuances (a basic `X-Forwarded-*` read is provided).
