# Backlog — missing features & gaps

An honest, prioritized list of what's still missing in this Express port. This is
curated real work, not padding — see the note at the bottom about the "10,000"
target.

## Core framework gaps

- [ ] Template engine ecosystem: only `html/template` is built in. Adapters for
      pug/jade, ejs, handlebars, mustache, liquid, and a layout/partials system.
- [ ] `res.render` callback form and streaming template output.
- [ ] View caching (`view cache` setting) and view resolution across multiple
      directories.
- [ ] `res.sendFile` options object (root, dotfiles, maxAge, lastModified,
      headers, acceptRanges) — currently a single path.
- [ ] `res.download` callback + Content-Disposition RFC 5987 (UTF-8 filenames).
- [ ] `express.static` full option set: `index`, `redirect`, `extensions`,
      `fallthrough`, `immutable`, `maxAge`, `setHeaders`, `dotfiles`.
- [ ] Directory traversal hardening tests + symlink policy for static serving.
- [ ] `req.accepts*` quality edge cases (RFC 7231 precedence, `*` vs specific).
- [ ] `res.vary` deduplication against existing header tokens.
- [ ] ETag generation strategies (`weak`/`strong`/custom fn) as an app setting,
      not just the middleware.
- [ ] `req.fresh`/`res.fresh` integration into `res.send` (auto-304).
- [ ] `app.set('etag')`, `app.set('query parser')`, `app.set('trust proxy')`
      with the full Express semantics (subnets, hop counts, functions).
- [ ] `req.subdomains`, `req.baseUrl`, `req.originalUrl` vs `req.url` after mounts.
- [ ] `req.route` (the matched route object) exposed to handlers.
- [ ] `req.xhr`, `req.secure` behind proxies, `req.ips` (the full XFF chain).
- [ ] Path patterns: named wildcards `*name`, `{optional}` groups, regex-literal
      routes, arrays of paths, `(a|b)` alternation.
- [ ] `app.mountpath` and `mount` events for sub-apps (vs sub-routers).
- [ ] `app.engine`, `app.locals`, `res.locals` inheritance chain parity.
- [ ] `res.jsonp` app callback-name setting; `res.links`; `res.location` with
      back/relative resolution.
- [ ] Content negotiation for `res.format` with `default` + `types` + charset.
- [ ] Async handler support with automatic error forwarding (panic→next(err)).
- [ ] Router `caseSensitive`/`strict` propagation to nested routers by default.
- [ ] HEAD auto-handling from GET routes; OPTIONS auto-Allow header.
- [ ] Trailing-slash redirect option; `strict routing` 308 redirects.
- [ ] `req.is()` with `+suffix` and wildcard media ranges.
- [ ] Signed cookies at the framework level (cookie secret rotation).
- [ ] Range requests for `res.send` of large buffers.

## Bundled/ecosystem middleware still to port

Real npm packages with no Go equivalent here yet:

- [ ] `serve-favicon` (full, with caching) — have a basic `favicon`.
- [ ] `connect-timeout` (per-route) — have `timeout`.
- [ ] `express-rate-limit` store adapters (redis, memcached, mongo).
- [ ] `express-slow-down` full parity.
- [ ] `helmet` remaining pieces: `contentSecurityPolicy` reporting, `hsts`
      preload submission helper, `crossOriginResourcePolicy` per-route.
- [ ] `cors` preflight caching + dynamic origin function + `optionsSuccessStatus`.
- [ ] `compression` brotli (`br`), deflate, per-type filters, threshold tuning.
- [ ] `body-parser` limits per content-type, `verify` hook, `inflate`, `strict`.
- [ ] `multer` disk storage engine, file filters, limits, field parsing.
- [ ] `express-session` store adapters (redis, mongo, postgres, memcached),
      rolling sessions, `resave`/`saveUninitialized`/`touch` semantics.
- [ ] `cookie-session` (stateless) full parity.
- [ ] `csurf` double-submit + synchronizer token + SameSite integration.
- [ ] `method-override` from body/query/header variants.
- [ ] `response-time` digits/suffix options.
- [ ] `morgan` predefined formats (combined, common, dev, short, tiny) + tokens.
- [ ] `serve-index` styling/icons/templates.
- [ ] `vhost` for full sub-apps.
- [ ] `errorhandler` HTML/JSON/text negotiation with stack traces (dev only).
- [ ] `express-http-proxy` / reverse proxy middleware.
- [ ] `connect-history-api-fallback` full option set.
- [ ] `express-validator` sanitizers, nested/wildcard fields, custom async
      validators, i18n messages, `oneOf`.
- [ ] `express-fileupload`, `express-formidable`.
- [ ] `express-jwt` + `jwks-rsa` (RS256 via JWKS).
- [ ] `express-openid-connect`, `express-basic-auth`, `passport` wiring helpers.
- [ ] `express-async-errors` behavior.
- [ ] `express-ws` (WebSocket routes) beyond `WrapHandler`.
- [ ] `express-sslify`, `express-enforces-ssl`.
- [ ] `express-status-monitor` (metrics dashboard).
- [ ] `express-graphql` / GraphQL handler.
- [ ] `express-prometheus-middleware`.
- [ ] `serve-static` + `finalhandler` exact behavior.
- [ ] `express-list-endpoints` (introspection).
- [ ] `express-paginate`.
- [ ] `connect-redis`, `connect-mongo` session stores.
- [ ] i18n middleware (`i18next-http-middleware`).
- [ ] `express-http-context` (per-request context via context.Context).

## Testing / tooling

- [ ] Benchmarks for routing, middleware chains, and JSON encoding.
- [ ] Fuzz tests for the path matcher and body parsers.
- [ ] `golangci-lint` config + staticcheck.
- [ ] HTTP/2 and TLS example + graceful shutdown helper.
- [ ] Supertest-style test helper package.

---

### On the "10,000 items" request

A literal 10,000 hand-written entries would be ~95% synthetic padding
("edge-case N of feature X"), which wouldn't help anyone. This file lists real,
actionable gaps. The single largest genuine source of volume is the npm
middleware ecosystem (thousands of packages) — porting each is one real item.
Ask and I'll expand any section into an exhaustive per-package checklist.
