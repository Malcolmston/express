// Package requestdump provides development middleware that records a snapshot
// of each incoming request (method, path, and headers) into a bounded,
// thread-safe ring buffer. It is handy for inspecting the most recent requests
// from tests or a debug endpoint.
package requestdump

import (
	"net/http"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Dump is a captured snapshot of a single request.
type Dump struct {
	Method  string
	Path    string
	Headers http.Header
	Time    time.Time
}

// Options configures the request dump middleware.
type Options struct {
	// Size is the maximum number of recent requests retained (default 32).
	Size int
}

var (
	mu   sync.Mutex
	ring []Dump
	maxN = 32
)

// New returns middleware that records each request into the ring buffer. If
// Options.Size is provided it resizes the buffer. Inspect captured requests
// with Last and All.
func New(opts ...Options) express.Handler {
	if len(opts) > 0 && opts[0].Size > 0 {
		mu.Lock()
		maxN = opts[0].Size
		if len(ring) > maxN {
			ring = ring[len(ring)-maxN:]
		}
		mu.Unlock()
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		d := Dump{
			Method:  req.Method(),
			Path:    req.Path(),
			Headers: req.Raw.Header.Clone(),
			Time:    time.Now(),
		}
		mu.Lock()
		ring = append(ring, d)
		if len(ring) > maxN {
			ring = ring[len(ring)-maxN:]
		}
		mu.Unlock()
		next()
	}
}

// Last returns the most recently captured request. The zero Dump is returned
// if none have been captured yet.
func Last() Dump {
	mu.Lock()
	defer mu.Unlock()
	if len(ring) == 0 {
		return Dump{}
	}
	return ring[len(ring)-1]
}

// All returns a copy of the currently retained request snapshots, oldest first.
func All() []Dump {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Dump, len(ring))
	copy(out, ring)
	return out
}

// Reset clears the captured request history.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	ring = nil
}
