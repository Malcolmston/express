// Package flash implements one-time "flash" messages backed by the express
// session. It is a stdlib-only port of the Node "connect-flash" middleware
// (the flash helper popularized by Express and Connect): a message is stored
// under a reserved session key during one request/response cycle and then read
// — and atomically cleared — on a subsequent one. The classic use is the
// post/redirect/get flow, where a handler that mutates state records a "saved"
// or "invalid input" notice, issues a redirect, and the following GET renders
// and consumes that notice exactly once.
//
// Reach for flash messages whenever you need to carry a short-lived,
// user-facing status string across a redirect without threading it through the
// URL or a query parameter. Because the data lives in the session it survives
// the redirect but is scoped to the one browser session, and because Get
// deletes it after reading, a refresh of the destination page will not show the
// message a second time. Messages are grouped by a free-form category string
// (for example "error", "info", or "success") so templates can style them
// differently.
//
// The middleware itself is deliberately thin. New returns a handler that does
// nothing but call next(); it exists so the package can be mounted like any
// other middleware and to leave room for future request-scoped setup. The real
// work is done by the package functions Add and Get, which operate directly on
// req.Session(). This means flash requires the express session middleware
// (express.Session()) to be installed earlier in the chain — mount flash.New
// after it. Neither the middleware nor the helpers write to the response or set
// headers; persistence is handled entirely by the session layer, which writes
// the session cookie on its own.
//
// The helpers are written to be safe when no session is attached: if
// req.Session() returns nil, Add silently does nothing and Get returns nil
// rather than panicking, so forgetting the session middleware degrades to a
// no-op instead of a crash. Add appends to the existing slice of messages under
// the reserved key ("_flash"), preserving insertion order across multiple calls
// within a request. Get returns the full slice and then calls Delete on the
// session key, so a second Get in the same request (or any later request)
// observes an empty queue; it returns nil — not an empty non-nil slice — when
// there is nothing pending.
//
// Compared with connect-flash the API is trimmed to two operations. connect-flash
// exposes a single req.flash(type, msg) that both adds and, when called with no
// message argument, reads-and-clears a category; here those roles are split into
// the explicit Add and Get functions, and Get always drains every category at
// once rather than one category per call. The category/text pair is modeled by
// the exported Message type. Storage fidelity depends on the session store: the
// messages must round-trip through whatever serialization the session backend
// uses, and the in-memory default store preserves the []Message slice as-is.
package flash

import "github.com/malcolmston/express"

// sessionKey is the session key under which flash messages are stored.
const sessionKey = "_flash"

// Message is a single flash message with a category (e.g. "error", "info")
// and its text.
type Message struct {
	// Category labels the message, typically for styling or filtering in a
	// template — common values are "error", "info", "warning", and "success".
	// It is free-form and never interpreted by this package.
	Category string
	// Message is the human-readable text to surface to the user.
	Message string
}

// New returns middleware that enables flash-message helpers for downstream
// handlers. It requires the express session middleware to be installed earlier
// in the chain; if no session is present the helpers degrade to no-ops.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		next()
	}
}

// Add appends a flash message under the given category to the session. It is a
// no-op if no session is attached to the request.
func Add(req *express.Request, category, msg string) {
	s := req.Session()
	if s == nil {
		return
	}
	msgs := read(s)
	msgs = append(msgs, Message{Category: category, Message: msg})
	s.Set(sessionKey, msgs)
}

// Get returns all pending flash messages and clears them from the session. It
// returns nil if no session is attached or no messages are pending.
func Get(req *express.Request) []Message {
	s := req.Session()
	if s == nil {
		return nil
	}
	msgs := read(s)
	if len(msgs) == 0 {
		return nil
	}
	s.Delete(sessionKey)
	return msgs
}

func read(s *express.SessionData) []Message {
	if v, ok := s.Get(sessionKey); ok {
		if m, ok := v.([]Message); ok {
			return m
		}
	}
	return nil
}
