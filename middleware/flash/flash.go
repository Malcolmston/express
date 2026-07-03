// Package flash implements one-time "flash" messages backed by the express
// session. Messages are added during one request/response cycle and read
// (and cleared) on a subsequent one, a common pattern for surfacing status
// notices after a redirect.
package flash

import "github.com/malcolmston/express"

// sessionKey is the session key under which flash messages are stored.
const sessionKey = "_flash"

// Message is a single flash message with a category (e.g. "error", "info")
// and its text.
type Message struct {
	Category string
	Message  string
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
