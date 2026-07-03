// Package cookiesession implements a lightweight, cookie-only session store.
// The entire session payload is serialized to JSON, HMAC-signed with a secret,
// and stored in a single cookie on the client. No server-side storage is
// required. Because the payload travels with every request it is intended for
// small amounts of data.
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
