// Package cors provides configurable Cross-Origin Resource Sharing (CORS)
// middleware for the express framework. It sets the appropriate
// Access-Control-* response headers and answers CORS preflight (OPTIONS)
// requests with a 204 No Content response.
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
