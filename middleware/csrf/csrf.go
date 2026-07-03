// Package csrf provides Cross-Site Request Forgery protection using the
// double-submit-cookie pattern. A random token is issued in a cookie and, for
// state-changing requests, must be echoed back in a request header or form
// field. Requests whose submitted token does not match the cookie are rejected
// with 403 Forbidden.
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/malcolmston/express"
)

const contextKey = "csrf"

// Options configures the CSRF middleware.
type Options struct {
	// CookieName is the name of the cookie that stores the token
	// (default "csrf").
	CookieName string
	// Header is the request header inspected for the submitted token
	// (default "X-CSRF-Token").
	Header string
	// FormField is the form field inspected for the submitted token when the
	// header is absent (default "_csrf").
	FormField string
	// MaxAge sets the token cookie Max-Age in seconds (0 = session cookie).
	MaxAge int
	// Secure marks the token cookie Secure (HTTPS only).
	Secure bool
}

func (o *Options) applyDefaults() {
	if o.CookieName == "" {
		o.CookieName = "csrf"
	}
	if o.Header == "" {
		o.Header = "X-CSRF-Token"
	}
	if o.FormField == "" {
		o.FormField = "_csrf"
	}
}

var unsafeMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

// New returns CSRF protection middleware. It ensures a token cookie exists on
// every request and, for unsafe methods (POST/PUT/PATCH/DELETE), verifies that
// the token submitted in the configured header or form field matches the
// cookie. On mismatch it responds 403 and does not call next.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		token := req.Cookie(o.CookieName)
		if token == "" {
			token = generateToken()
			res.Cookie(o.CookieName, token, &express.CookieOptions{
				Path:     "/",
				MaxAge:   o.MaxAge,
				Secure:   o.Secure,
				HTTPOnly: false, // must be readable by client scripts
				SameSite: http.SameSiteLaxMode,
			})
		}
		req.Set(contextKey, token)

		if unsafeMethods[req.Method()] {
			submitted := req.Get(o.Header)
			if submitted == "" {
				submitted = req.FormValue(o.FormField)
			}
			if submitted == "" || subtle.ConstantTimeCompare([]byte(submitted), []byte(token)) != 1 {
				res.Status(http.StatusForbidden).JSON(map[string]string{
					"error": "invalid CSRF token",
				})
				return
			}
		}
		next()
	}
}

// Token returns the CSRF token associated with the current request, or "" if
// the middleware did not run.
func Token(req *express.Request) string {
	if v, ok := req.Value(contextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// rand.Read from crypto/rand does not fail in practice.
		return "insecure-fallback-token"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
