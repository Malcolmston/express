# Middleware catalog

`express` ships **102 middleware packages** under `middleware/`. Each is an
independent subpackage with a `New(...)` constructor returning `express.Handler`
(or `express.ErrorHandler`), configurable via an `Options` struct, standard-library only.

| Package | Import | Description |
| ------- | ------ | ----------- |
| `abtest` | `github.com/malcolmston/express/middleware/abtest` | returns middleware that assigns and persists a stable bucket for each |
| `acceptlanguage` | `github.com/malcolmston/express/middleware/acceptlanguage` | returns middleware that stores the negotiated language via |
| `accesslog` | `github.com/malcolmston/express/middleware/accesslog` | returns middleware that writes an Apache combined-log-format line after |
| `apikey` | `github.com/malcolmston/express/middleware/apikey` | returns middleware that requires a valid API key. Missing or invalid |
| `basepath` | `github.com/malcolmston/express/middleware/basepath` | returns middleware that strips Options.Prefix from req.Raw.URL.Path |
| `basicauth` | `github.com/malcolmston/express/middleware/basicauth` | returns middleware that enforces HTTP Basic authentication. Requests |
| `bearerauth` | `github.com/malcolmston/express/middleware/bearerauth` | returns middleware that reads an "Authorization: Bearer <token>" header |
| `bodylimit` | `github.com/malcolmston/express/middleware/bodylimit` | returns middleware enforcing a maximum request body size. If the |
| `cachecontrol` | `github.com/malcolmston/express/middleware/cachecontrol` | returns middleware that sets the Cache-Control response header built from |
| `circuitbreaker` | `github.com/malcolmston/express/middleware/circuitbreaker` | returns circuit-breaker middleware configured by opts |
| `compression` | `github.com/malcolmston/express/middleware/compression` | returns middleware that gzip-compresses eligible responses. A response is |
| `concurrencylimit` | `github.com/malcolmston/express/middleware/concurrencylimit` | returns concurrency-limiting middleware configured by opts |
| `contenttypedefault` | `github.com/malcolmston/express/middleware/contenttypedefault` | returns middleware that, just before the response headers are committed, |
| `cookieparser` | `github.com/malcolmston/express/middleware/cookieparser` | returns middleware that parses every cookie on the request into a |
| `cookiesession` | `github.com/malcolmston/express/middleware/cookiesession` | returns middleware that loads a signed session cookie into a per-request |
| `correlationid` | `github.com/malcolmston/express/middleware/correlationid` | returns middleware that ensures every request carries a correlation id |
| `cors` | `github.com/malcolmston/express/middleware/cors` | returns CORS middleware configured by the optional Options. Passing no |
| `crossoriginembedder` | `github.com/malcolmston/express/middleware/crossoriginembedder` | returns middleware that sets the Cross-Origin-Embedder-Policy header |
| `crossoriginopener` | `github.com/malcolmston/express/middleware/crossoriginopener` | returns middleware that sets the Cross-Origin-Opener-Policy header |
| `crossoriginresource` | `github.com/malcolmston/express/middleware/crossoriginresource` | returns middleware that sets the Cross-Origin-Resource-Policy header |
| `csp` | `github.com/malcolmston/express/middleware/csp` | returns middleware that sets a Content-Security-Policy header |
| `cspnonce` | `github.com/malcolmston/express/middleware/cspnonce` | returns middleware that generates a nonce, exposes it via |
| `csrf` | `github.com/malcolmston/express/middleware/csrf` | returns CSRF protection middleware. It ensures a token cookie exists on |
| `decompress` | `github.com/malcolmston/express/middleware/decompress` | returns middleware that, when the request's Content-Encoding is gzip, |
| `dnsprefetch` | `github.com/malcolmston/express/middleware/dnsprefetch` | returns middleware that sets the X-DNS-Prefetch-Control header |
| `downloadheader` | `github.com/malcolmston/express/middleware/downloadheader` | returns middleware that sets Content-Disposition: attachment so browsers |
| `errorjson` | `github.com/malcolmston/express/middleware/errorjson` | returns an express.ErrorHandler that writes the error's message as a JSON |
| `etag` | `github.com/malcolmston/express/middleware/etag` | returns middleware that sets a strong ETag header (the SHA-1 of the |
| `expectct` | `github.com/malcolmston/express/middleware/expectct` | returns middleware that sets the Expect-CT header |
| `expires` | `github.com/malcolmston/express/middleware/expires` | returns middleware that sets the Expires response header to the current |
| `favicon` | `github.com/malcolmston/express/middleware/favicon` | returns middleware that handles GET/HEAD requests for /favicon.ico and |
| `featureflag` | `github.com/malcolmston/express/middleware/featureflag` | returns middleware that stores the configured flags on each request |
| `flash` | `github.com/malcolmston/express/middleware/flash` | returns middleware that enables flash-message helpers for downstream |
| `forcessl` | `github.com/malcolmston/express/middleware/forcessl` | returns middleware that redirects http requests to https with a 301 |
| `frameguard` | `github.com/malcolmston/express/middleware/frameguard` | returns middleware that sets the X-Frame-Options header |
| `healthcheck` | `github.com/malcolmston/express/middleware/healthcheck` | returns healthcheck middleware configured by opts |
| `healthz` | `github.com/malcolmston/express/middleware/healthz` | returns healthz middleware configured by opts |
| `helmet` | `github.com/malcolmston/express/middleware/helmet` | returns middleware that sets helmet's bundle of default security headers |
| `hidepoweredby` | `github.com/malcolmston/express/middleware/hidepoweredby` | returns middleware that hides or spoofs the X-Powered-By header. The |
| `hmacauth` | `github.com/malcolmston/express/middleware/hmacauth` | returns middleware that computes the HMAC-SHA256 of the raw request body |
| `hostcheck` | `github.com/malcolmston/express/middleware/hostcheck` | returns middleware that rejects requests whose Hostname is not present in |
| `hsts` | `github.com/malcolmston/express/middleware/hsts` | returns middleware that sets the Strict-Transport-Security header |
| `ienoopen` | `github.com/malcolmston/express/middleware/ienoopen` | returns middleware that sets X-Download-Options: noopen |
| `ipallowlist` | `github.com/malcolmston/express/middleware/ipallowlist` | returns middleware that responds with 403 unless the request's client IP |
| `ipblocklist` | `github.com/malcolmston/express/middleware/ipblocklist` | returns middleware that responds with 403 when the request's client IP |
| `jsonp` | `github.com/malcolmston/express/middleware/jsonp` | returns middleware that, when a valid callback query parameter is present, |
| `jwtauth` | `github.com/malcolmston/express/middleware/jwtauth` | returns middleware that verifies an HS256 JWT from the |
| `latencyheader` | `github.com/malcolmston/express/middleware/latencyheader` | returns latency-header middleware configured by opts |
| `maintenance` | `github.com/malcolmston/express/middleware/maintenance` | returns maintenance middleware and a Toggle for controlling it. When |
| `methodallow` | `github.com/malcolmston/express/middleware/methodallow` | returns middleware that responds with 405 Method Not Allowed, including |
| `methodoverride` | `github.com/malcolmston/express/middleware/methodoverride` | returns middleware that rewrites req.Raw.Method from the configured |
| `metrics` | `github.com/malcolmston/express/middleware/metrics` | returns metrics middleware together with the *Metrics accumulator it |
| `nocache` | `github.com/malcolmston/express/middleware/nocache` | returns middleware that sets the standard set of no-cache headers: |
| `nonce` | `github.com/malcolmston/express/middleware/nonce` | returns middleware that stores a per-request random nonce via |
| `nosniff` | `github.com/malcolmston/express/middleware/nosniff` | returns middleware that sets X-Content-Type-Options: nosniff |
| `notfound` | `github.com/malcolmston/express/middleware/notfound` | returns a terminal handler that always responds 404. It does not call |
| `originagentcluster` | `github.com/malcolmston/express/middleware/originagentcluster` | returns middleware that sets Origin-Agent-Cluster: ?1 |
| `origincheck` | `github.com/malcolmston/express/middleware/origincheck` | returns middleware that responds with 403 unless the request's Origin |
| `pagination` | `github.com/malcolmston/express/middleware/pagination` | returns middleware that reads ?page and ?limit, clamps them to sensible |
| `panicjson` | `github.com/malcolmston/express/middleware/panicjson` | returns middleware that recovers from panics raised by later handlers |
| `permittedcrossdomain` | `github.com/malcolmston/express/middleware/permittedcrossdomain` | returns middleware that sets the X-Permitted-Cross-Domain-Policies header |
| `poweredby` | `github.com/malcolmston/express/middleware/poweredby` | returns middleware that sets the X-Powered-By response header. This |
| `querylimit` | `github.com/malcolmston/express/middleware/querylimit` | returns query-limit middleware configured by opts |
| `querynormalize` | `github.com/malcolmston/express/middleware/querynormalize` | returns middleware that rewrites req.Raw.URL.RawQuery in normalized form |
| `ratelimit` | `github.com/malcolmston/express/middleware/ratelimit` | returns rate-limiting middleware configured by opts |
| `rawbody` | `github.com/malcolmston/express/middleware/rawbody` | returns middleware that buffers the full request body into a []byte and |
| `readiness` | `github.com/malcolmston/express/middleware/readiness` | returns readiness middleware configured by opts |
| `realip` | `github.com/malcolmston/express/middleware/realip` | returns middleware that resolves the client IP from forwarding headers |
| `redirectmap` | `github.com/malcolmston/express/middleware/redirectmap` | returns middleware that redirects any request whose path is present in |
| `referer` | `github.com/malcolmston/express/middleware/referer` | returns middleware that stores a Referer via req.Set(Key, ref). Both the |
| `referercheck` | `github.com/malcolmston/express/middleware/referercheck` | returns middleware that responds with 403 unless the request's Referer |
| `referrerpolicy` | `github.com/malcolmston/express/middleware/referrerpolicy` | returns middleware that sets the Referrer-Policy header |
| `requestcontext` | `github.com/malcolmston/express/middleware/requestcontext` | returns middleware that attaches a *Ctx via req.Set(Key, ctx) and sets |
| `requestcounter` | `github.com/malcolmston/express/middleware/requestcounter` | returns request-counting middleware together with an accessor that |
| `requestdump` | `github.com/malcolmston/express/middleware/requestdump` | returns middleware that records each request into the ring buffer. If |
| `requestid` | `github.com/malcolmston/express/middleware/requestid` | returns middleware that ensures every request carries an id. If the |
| `requireauth` | `github.com/malcolmston/express/middleware/requireauth` | returns middleware that responds with 401 unless the configured request |
| `responsecache` | `github.com/malcolmston/express/middleware/responsecache` | returns middleware that caches successful GET responses in memory for the |
| `responsetime` | `github.com/malcolmston/express/middleware/responsetime` | returns middleware that records the time spent handling a request and |
| `retryafter` | `github.com/malcolmston/express/middleware/retryafter` | returns retry-after middleware configured by opts |
| `rewrite` | `github.com/malcolmston/express/middleware/rewrite` | returns middleware that rewrites req.Raw.URL.Path according to the first |
| `rolecheck` | `github.com/malcolmston/express/middleware/rolecheck` | returns middleware that responds with 403 unless the request holds at |
| `sanitize` | `github.com/malcolmston/express/middleware/sanitize` | returns middleware that removes HTML tags from every query-string value |
| `scopecheck` | `github.com/malcolmston/express/middleware/scopecheck` | returns middleware that responds with 403 unless every required scope is |
| `serveindex` | `github.com/malcolmston/express/middleware/serveindex` | returns middleware that serves directory listings from Options.Root |
| `servertiming` | `github.com/malcolmston/express/middleware/servertiming` | returns middleware that installs a Metrics collector on the request and |
| `signedcookies` | `github.com/malcolmston/express/middleware/signedcookies` | returns middleware that verifies the configured signed cookie. On |
| `slowdown` | `github.com/malcolmston/express/middleware/slowdown` | returns slow-down middleware configured by opts |
| `slowlog` | `github.com/malcolmston/express/middleware/slowlog` | returns middleware that measures the time spent handling each request and |
| `spa` | `github.com/malcolmston/express/middleware/spa` | returns middleware that serves files from Root, falling back to Index |
| `subdomain` | `github.com/malcolmston/express/middleware/subdomain` | returns middleware that computes the subdomain and stores it via |
| `throttle` | `github.com/malcolmston/express/middleware/throttle` | returns throttle middleware configured by opts |
| `timeout` | `github.com/malcolmston/express/middleware/timeout` | returns timeout middleware configured by opts |
| `tokenheader` | `github.com/malcolmston/express/middleware/tokenheader` | returns middleware that reads the configured header and rejects the |
| `trailingslash` | `github.com/malcolmston/express/middleware/trailingslash` | returns middleware implementing the configured trailing-slash policy |
| `uptime` | `github.com/malcolmston/express/middleware/uptime` | returns uptime middleware together with a Since accessor reporting the |
| `useragent` | `github.com/malcolmston/express/middleware/useragent` | returns middleware that parses the User-Agent header and stores the |
| `useragentblock` | `github.com/malcolmston/express/middleware/useragentblock` | returns middleware that responds with 403 when the request's User-Agent |
| `vary` | `github.com/malcolmston/express/middleware/vary` | returns middleware that appends each configured field to the response's |
| `version` | `github.com/malcolmston/express/middleware/version` | returns version middleware configured by opts |
| `vhost` | `github.com/malcolmston/express/middleware/vhost` | returns middleware that invokes Options.Handler for requests whose |
| `xssfilter` | `github.com/malcolmston/express/middleware/xssfilter` | returns middleware that sets the X-XSS-Protection header |
