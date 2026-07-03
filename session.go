package express

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

// DefaultSessionCookie is the cookie name used by the session middleware,
// matching the default used by Node's express-session.
const DefaultSessionCookie = "connect.sid"

// SessionStore persists session data keyed by an opaque id. Implementations
// must be safe for concurrent use.
type SessionStore interface {
	Get(id string) (map[string]any, bool)
	Set(id string, data map[string]any)
	Destroy(id string)
}

// MemorySessionStore is an in-memory SessionStore for development and
// single-process use.
type MemorySessionStore struct {
	mu   sync.RWMutex
	data map[string]map[string]any
}

// NewMemorySessionStore creates an empty in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{data: make(map[string]map[string]any)}
}

func (m *MemorySessionStore) Get(id string) (map[string]any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.data[id]
	if !ok {
		return nil, false
	}
	cp := make(map[string]any, len(d))
	for k, v := range d {
		cp[k] = v
	}
	return cp, true
}

func (m *MemorySessionStore) Set(id string, data map[string]any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make(map[string]any, len(data))
	for k, v := range data {
		cp[k] = v
	}
	m.data[id] = cp
}

func (m *MemorySessionStore) Destroy(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, id)
}

// SessionOptions configures the session middleware.
type SessionOptions struct {
	// Name is the session cookie name (default "connect.sid").
	Name string
	// Store persists sessions (default an in-memory store).
	Store SessionStore
	// Secure marks the cookie Secure (HTTPS only).
	Secure bool
	// HTTPOnly marks the cookie HttpOnly (default true).
	HTTPOnly bool
	// SameSite sets the cookie SameSite policy (default Lax).
	SameSite http.SameSite
	// MaxAge sets the cookie Max-Age in seconds (0 = session cookie).
	MaxAge int
}

// SessionData is the per-request session object exposed to handlers via
// req.Session(). Values are read and written with Get/Set; changes are
// persisted automatically before the response is sent.
type SessionData struct {
	// Values holds the session data.
	Values map[string]any

	id      string
	store   SessionStore
	opts    *SessionOptions
	dirty   bool
	destroy bool
}

// Get returns a session value and whether it was present.
func (s *SessionData) Get(key string) (any, bool) {
	v, ok := s.Values[key]
	return v, ok
}

// GetString returns a string session value, or "" if missing / not a string.
func (s *SessionData) GetString(key string) string {
	if v, ok := s.Values[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// Set stores a session value and marks the session dirty.
func (s *SessionData) Set(key string, value any) {
	s.Values[key] = value
	s.dirty = true
}

// Delete removes a session value.
func (s *SessionData) Delete(key string) {
	delete(s.Values, key)
	s.dirty = true
}

// Regenerate issues a fresh session id, discarding the old one. Call this on
// privilege changes such as login to defeat session fixation.
func (s *SessionData) Regenerate() error {
	if s.id != "" {
		s.store.Destroy(s.id)
	}
	id, err := randomID()
	if err != nil {
		return err
	}
	s.id = id
	s.dirty = true
	return nil
}

// Destroy marks the session for removal; the cookie is cleared on response.
func (s *SessionData) Destroy() {
	s.destroy = true
	s.dirty = true
}

// Session returns the session attached to the request by the Session
// middleware, or nil if the middleware is not installed.
func (req *Request) Session() *SessionData {
	if v, ok := req.values["session"]; ok {
		if s, ok := v.(*SessionData); ok {
			return s
		}
	}
	return nil
}

// Session returns middleware that loads a per-request session from a cookie and
// persists changes automatically. Access it in handlers via req.Session().
func Session(opts ...SessionOptions) Handler {
	o := SessionOptions{HTTPOnly: true, SameSite: http.SameSiteLaxMode}
	if len(opts) > 0 {
		o = opts[0]
		if o.SameSite == 0 {
			o.SameSite = http.SameSiteLaxMode
		}
	}
	if o.Name == "" {
		o.Name = DefaultSessionCookie
	}
	if o.Store == nil {
		o.Store = NewMemorySessionStore()
	}

	return func(req *Request, res *Response, next Next) {
		sess := &SessionData{Values: map[string]any{}, store: o.Store, opts: &o}
		if c, err := req.Raw.Cookie(o.Name); err == nil && c.Value != "" {
			if data, ok := o.Store.Get(c.Value); ok {
				sess.id = c.Value
				sess.Values = data
			}
		}
		req.Set("session", sess)

		// Persist the session just before the response headers are written.
		res.OnBeforeWrite(func() {
			if sess.destroy {
				if sess.id != "" {
					o.Store.Destroy(sess.id)
				}
				clearCookie(res, &o)
				return
			}
			if !sess.dirty {
				return
			}
			if sess.id == "" {
				if id, err := randomID(); err == nil {
					sess.id = id
				}
			}
			o.Store.Set(sess.id, sess.Values)
			setSessionCookie(res, sess.id, &o)
		})

		next()
	}
}

func setSessionCookie(res *Response, id string, o *SessionOptions) {
	c := &http.Cookie{
		Name:     o.Name,
		Value:    id,
		Path:     "/",
		HttpOnly: o.HTTPOnly,
		Secure:   o.Secure,
		SameSite: o.SameSite,
		MaxAge:   o.MaxAge,
	}
	http.SetCookie(res.Writer, c)
}

func clearCookie(res *Response, o *SessionOptions) {
	http.SetCookie(res.Writer, &http.Cookie{
		Name:     o.Name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: o.HTTPOnly,
		Secure:   o.Secure,
		SameSite: o.SameSite,
	})
}

func randomID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
