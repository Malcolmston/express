// Package csrf provides Cross-Site Request Forgery protection using the
// double-submit-cookie pattern. It is the Go analogue of the classic csurf
// middleware from the Node ecosystem, packaged as a drop-in express.Handler. A
// random token is issued in a cookie and, for state-changing requests, must be
// echoed back in a request header or form field; requests whose submitted token
// does not match the cookie are rejected with 403 Forbidden and never reach the
// downstream handler.
//
// Use this middleware to defend endpoints that mutate server state — form posts,
// account changes, deletions — against forged requests triggered by a malicious
// third-party site while the victim is authenticated. The double-submit-cookie
// scheme works because a foreign origin can cause the browser to send the
// victim's cookies but cannot read them to reproduce the token in a header or
// body, so a matching submission proves the request originated from your own
// pages. Mount it with app.Use for a global guard, or attach it to the router
// subtree that owns your state-changing routes.
//
// Operationally the middleware sits at the front of the chain. On every request
// it looks for the token cookie (Options.CookieName, default "csrf"); if absent
// it generates a fresh 32-byte token, base64url-encodes it, and issues it as a
// Lax, non-HTTPOnly cookie so client scripts and templates can read it back. The
// active token is stored on the request via req.Set and can be retrieved with
// Token, which lets your GET handlers embed it in forms or expose it to
// JavaScript. Safe methods (GET, HEAD, OPTIONS, TRACE) always call next() after
// ensuring the cookie exists; only the unsafe methods POST, PUT, PATCH, and
// DELETE are verified.
//
// For unsafe methods the middleware reads the submitted token from the
// configured header (Options.Header, default "X-CSRF-Token") and, if that is
// empty, from the form field (Options.FormField, default "_csrf"). It compares
// the submission against the cookie value with crypto/subtle.ConstantTimeCompare
// to avoid leaking timing information. If the submitted token is missing or does
// not match, the request is short-circuited with a 403 Forbidden JSON body
// ({"error":"invalid CSRF token"}) and next() is never invoked. Options.MaxAge
// controls the cookie lifetime in seconds (0 yields a session cookie) and
// Options.Secure marks the cookie HTTPS-only; the zero-value Options is usable
// and applies all defaults.
//
// Compared with the Node csurf original, this port keeps the double-submit
// contract — token in a cookie, echoed in a header or field, constant-time
// compared, unsafe methods enforced — but is deliberately minimal. It does not
// integrate with server-side sessions or offer csurf's signed-cookie or
// session-backed token storage, does not expose a req.csrfToken() function
// (use Token instead), and always sets HTTPOnly=false because the token must be
// readable by the client to be resubmitted. Because the token is not tied to a
// server-side secret, deploy it over TLS and rely on the SameSite=Lax cookie
// attribute as defense in depth.
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
