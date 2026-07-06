// Package cookiesession implements a lightweight, cookie-only session store as
// express middleware. The entire session payload is serialized to JSON,
// HMAC-signed with a secret, and stored in a single cookie on the client, so
// no server-side session store is required. It is the Go analogue of the Node
// cookie-session middleware (expressjs/cookie-session), packaged as a drop-in
// express.Handler, and like that library it keeps state in the cookie itself
// rather than in a backend keyed by a session id.
//
// Use this middleware when you want per-user session state without the
// operational cost of a shared store (Redis, a database, or sticky sessions),
// and when the data is small — a user id, a few flags, a CSRF secret. Because
// the payload travels in every request and response and browsers cap cookie
// size (roughly 4KB), it is unsuitable for large or sensitive-in-plaintext
// data. Mount it once near the top of the chain with app.Use so the session is
// loaded before your handlers run, then read and write it with the package
// Get and Set helpers.
//
// Operationally the middleware runs early. On each request it looks for the
// configured cookie (Options.CookieName, default "session"); if present it
// verifies the HMAC signature and, only on success, decodes the JSON into a
// per-request map. That map is attached to the request under an internal key
// and next() is called so handlers can access it. The middleware also
// registers a res.OnBeforeWrite callback that runs just before the response
// headers are committed: if and only if the session was modified during the
// request, it re-serializes the map, signs it, and writes the refreshed cookie
// via res.Cookie. An unmodified session produces no Set-Cookie header at all,
// which avoids needless cookie churn and keeps caches friendlier.
//
// The signed value is "base64url(json).base64url(hmacSHA256(json))". On read,
// a missing dot, undecodable segments, a signature mismatch, or malformed JSON
// all cause the cookie to be silently ignored and the request to start with an
// empty session, so a tampered or corrupt cookie can never inject forged
// values. Set marks the session dirty, which is what arms the write-back;
// Get reports both the value and whether the key was present. The emitted
// cookie is always Path="/", HTTPOnly, and SameSite=Lax; Options.MaxAge sets
// its Max-Age in seconds (0 means a session cookie that expires when the
// browser closes) and Options.Secure restricts it to HTTPS. Options.Secret is
// the signing key and must be set — an empty secret still produces a valid
// HMAC but offers no real tamper protection and is strongly discouraged.
//
// Compared with the Node original, this port keeps the same signed,
// stateless-server, write-on-change model and the same "reject silently on bad
// signature" behaviour, but is deliberately narrower. It stores a single
// session object rather than exposing a rotating array of signing keys, signs
// with a fixed HMAC-SHA256 scheme instead of a pluggable keygrip rotation, and
// does not encrypt the payload — the JSON is signed for integrity but remains
// readable by the client, so never place secrets that must stay hidden from
// the user inside the session.
package cookiesession

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/malcolmston/express"
)

const contextKey = "cookiesession"

// Options configures the cookie session middleware.
type Options struct {
	// Secret is the key used to HMAC-sign the cookie payload. It must be set;
	// an empty secret disables tamper protection and is strongly discouraged.
	Secret string
	// CookieName is the name of the session cookie (default "session").
	CookieName string
	// MaxAge sets the cookie Max-Age in seconds (0 = session cookie).
	MaxAge int
	// Secure marks the cookie Secure (HTTPS only).
	Secure bool
}

func (o *Options) applyDefaults() {
	if o.CookieName == "" {
		o.CookieName = "session"
	}
}

// store holds the decoded session map for a single request along with a dirty
// flag indicating whether it needs to be re-serialized to a cookie.
type store struct {
	values map[string]any
	dirty  bool
}

// New returns middleware that loads a signed session cookie into a per-request
// map and, if modified, writes the updated signed cookie before the response
// headers are committed. Access values with Get and Set.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		st := &store{values: map[string]any{}}
		if raw := req.Cookie(o.CookieName); raw != "" {
			if m, ok := decode(raw, o.Secret); ok {
				st.values = m
			}
		}
		req.Set(contextKey, st)

		res.OnBeforeWrite(func() {
			if !st.dirty {
				return
			}
			encoded, err := encode(st.values, o.Secret)
			if err != nil {
				return
			}
			res.Cookie(o.CookieName, encoded, &express.CookieOptions{
				Path:     "/",
				MaxAge:   o.MaxAge,
				Secure:   o.Secure,
				HTTPOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		})

		next()
	}
}

// Get returns a value from the loaded session and whether it was present.
func Get(req *express.Request, key string) (any, bool) {
	st := load(req)
	if st == nil {
		return nil, false
	}
	v, ok := st.values[key]
	return v, ok
}

// Set stores a value in the session and marks it for persistence.
func Set(req *express.Request, key string, value any) {
	st := load(req)
	if st == nil {
		return
	}
	st.values[key] = value
	st.dirty = true
}

func load(req *express.Request) *store {
	if v, ok := req.Value(contextKey); ok {
		if st, ok := v.(*store); ok {
			return st
		}
	}
	return nil
}

func encode(values map[string]any, secret string) (string, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	mac := sign(data, secret)
	return base64.RawURLEncoding.EncodeToString(data) + "." + base64.RawURLEncoding.EncodeToString(mac), nil
}

func decode(raw, secret string) (map[string]any, bool) {
	i := strings.IndexByte(raw, '.')
	if i < 0 {
		return nil, false
	}
	data, err := base64.RawURLEncoding.DecodeString(raw[:i])
	if err != nil {
		return nil, false
	}
	sig, err := base64.RawURLEncoding.DecodeString(raw[i+1:])
	if err != nil {
		return nil, false
	}
	if !hmac.Equal(sig, sign(data, secret)) {
		return nil, false
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, false
	}
	return m, true
}

func sign(data []byte, secret string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return h.Sum(nil)
}
