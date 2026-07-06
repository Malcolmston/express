// Package cors provides configurable Cross-Origin Resource Sharing (CORS)
// middleware for the express framework. It sets the appropriate
// Access-Control-* response headers and answers CORS preflight (OPTIONS)
// requests with a 204 No Content response. It is the Go analogue of the
// ubiquitous Express cors package (expressjs/cors), packaged as a drop-in
// express.Handler.
//
// Use this middleware whenever a browser-based client hosted on one origin must
// call an API served from another origin. The same-origin policy blocks such
// requests by default; CORS is the mechanism by which the server opts specific
// origins back in and tells the browser which methods, request headers, and
// response headers are permitted, and whether credentials (cookies,
// Authorization headers) may be sent. Typical deployments mount it globally
// with app.Use so both the simple requests and the preflight probes for every
// route are handled, but it can equally guard a single router or path prefix.
//
// Operationally the middleware runs near the front of the chain. On every
// request it reads the Origin request header and, via resolveOrigin, decides
// whether that origin is permitted; when it is, it writes
// Access-Control-Allow-Origin (adding a Vary: Origin header whenever the value
// is a concrete origin rather than "*", so shared caches do not leak one
// origin's response to another). It then conditionally emits
// Access-Control-Allow-Credentials and Access-Control-Expose-Headers. If the
// request method is OPTIONS the middleware treats it as a preflight: it writes
// Access-Control-Allow-Methods, reflects or sets Access-Control-Allow-Headers,
// optionally sets Access-Control-Max-Age, sets Content-Length: 0, and
// short-circuits with a 204 No Content — next() is never called. For all other
// methods the middleware sets its headers and calls next() so the route handler
// runs normally.
//
// Behavior is driven by Options, whose zero value is usable and yields an open
// policy: AllowedOrigins defaults to []string{"*"} (any origin) and
// AllowedMethods defaults to GET, HEAD, PUT, PATCH, POST, DELETE. AllowedHeaders,
// when empty, causes the incoming Access-Control-Request-Headers value to be
// reflected back verbatim (and Vary: Access-Control-Request-Headers to be set).
// ExposedHeaders populates Access-Control-Expose-Headers. MaxAge sets the
// preflight cache lifetime in seconds; a non-positive value omits the header.
// The important edge case concerns credentials: the CORS specification forbids
// combining a wildcard origin with credentials, so when AllowCredentials is
// true and the resolved origin would be "*", the request's concrete Origin is
// echoed instead. An origin that is not in the allow-list simply receives no
// Access-Control-Allow-Origin header — the request is not rejected server-side;
// the browser enforces the policy by refusing the response.
//
// Compared with the Node original, this port keeps the same header-setting and
// preflight-terminating semantics and the same permissive defaults, but is
// intentionally leaner. It does not support per-request dynamic origin
// functions, regular-expression origin matching, the configurable
// preflightContinue / optionsSuccessStatus knobs, or custom preflight status
// codes; preflight always ends with 204. Origin matching is a case-insensitive
// exact comparison against the configured list (or an unconditional match on
// "*"), with no pattern support.
package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the CORS middleware. The zero value is usable and
// behaves as an open policy that allows any origin ("*") with the default set
// of methods.
type Options struct {
	// AllowedOrigins is the list of origins permitted to make cross-origin
	// requests. The special value "*" allows any origin. When empty it
	// defaults to []string{"*"}.
	AllowedOrigins []string

	// AllowedMethods is the list of HTTP methods advertised in preflight
	// responses. When empty a sensible default set is used.
	AllowedMethods []string

	// AllowedHeaders is the list of request headers advertised in preflight
	// responses. When empty the value of the incoming
	// Access-Control-Request-Headers is reflected back.
	AllowedHeaders []string

	// ExposedHeaders is the list of response headers exposed to the browser
	// via Access-Control-Expose-Headers.
	ExposedHeaders []string

	// AllowCredentials, when true, adds Access-Control-Allow-Credentials: true.
	// It cannot be combined with a wildcard origin; when credentials are
	// enabled the request's Origin is echoed instead.
	AllowCredentials bool

	// MaxAge is the number of seconds a preflight response may be cached. A
	// non-positive value omits the Access-Control-Max-Age header.
	MaxAge int
}

var defaultMethods = []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"}

// New returns CORS middleware configured by the optional Options. Passing no
// arguments yields an open policy allowing all origins.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if len(o.AllowedOrigins) == 0 {
		o.AllowedOrigins = []string{"*"}
	}
	if len(o.AllowedMethods) == 0 {
		o.AllowedMethods = defaultMethods
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		origin := req.Get("Origin")

		allowOrigin, ok := resolveOrigin(o.AllowedOrigins, origin)
		if ok {
			// When credentials are allowed we must echo the concrete origin
			// rather than "*".
			if o.AllowCredentials && allowOrigin == "*" && origin != "" {
				allowOrigin = origin
			}
			res.Set("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				res.Vary("Origin")
			}
		}

		if o.AllowCredentials {
			res.Set("Access-Control-Allow-Credentials", "true")
		}
		if len(o.ExposedHeaders) > 0 {
			res.Set("Access-Control-Expose-Headers", strings.Join(o.ExposedHeaders, ", "))
		}

		// Handle preflight requests.
		if strings.EqualFold(req.Method(), http.MethodOptions) {
			res.Set("Access-Control-Allow-Methods", strings.Join(o.AllowedMethods, ", "))

			if len(o.AllowedHeaders) > 0 {
				res.Set("Access-Control-Allow-Headers", strings.Join(o.AllowedHeaders, ", "))
			} else if reqHeaders := req.Get("Access-Control-Request-Headers"); reqHeaders != "" {
				res.Set("Access-Control-Allow-Headers", reqHeaders)
				res.Vary("Access-Control-Request-Headers")
			}

			if o.MaxAge > 0 {
				res.Set("Access-Control-Max-Age", strconv.Itoa(o.MaxAge))
			}

			res.Set("Content-Length", "0")
			res.Status(http.StatusNoContent).End()
			return
		}

		next()
	}
}

// resolveOrigin reports the value to use for Access-Control-Allow-Origin and
// whether the request origin is permitted.
func resolveOrigin(allowed []string, origin string) (string, bool) {
	for _, a := range allowed {
		if a == "*" {
			return "*", true
		}
		if origin != "" && strings.EqualFold(a, origin) {
			return origin, true
		}
	}
	return "", false
}
