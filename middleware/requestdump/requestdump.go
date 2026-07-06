// Package requestdump provides development middleware that records a snapshot
// of each incoming request (method, path, headers, and capture time) into a
// bounded, thread-safe ring buffer for later inspection. It is the express
// analogue of the request-dump/inspection helpers used while developing Node
// apps — the kind of thing you reach for alongside morgan or a console.dir of
// req — but instead of writing to a log it retains the most recent requests in
// memory so tests or a debug endpoint can read them back structurally.
//
// Use it during development and testing when you need to assert on, or eyeball,
// what actually arrived at the server: which method and path were hit and which
// headers the client sent. A test can drive the app and then call Last to
// inspect the final request, or a debug route can call All to render a table of
// recent traffic. It is deliberately not a production tool — captured headers
// may contain sensitive material (cookies, Authorization), and the buffer is a
// process-local, non-persistent side channel — so it should be mounted behind
// development guards rather than enabled globally in production.
//
// Mechanically the middleware runs early and, for each request, builds a Dump
// from req.Method(), req.Path(), a clone of req.Raw.Header, and the current
// time, appends it to the shared ring under a mutex, trims the ring to its
// maximum length, and then calls next() to continue. It writes no headers,
// reads no response, and never short-circuits. The header map is cloned so a
// later handler mutating req headers cannot retroactively alter a captured
// snapshot. All access to the ring is serialized by a package-level mutex, so
// the middleware and the Last, All, and Reset accessors are safe under
// concurrent traffic.
//
// State and options: the ring, its size, and the mutex are package-level, so
// the capture history is global and shared across every New in the process
// rather than per-instance. Options.Size sets the maximum number of retained
// snapshots (default 32); passing a Size on any New resizes the shared buffer
// and immediately trims older entries if it shrank. When the ring is full the
// oldest snapshot is dropped as new ones arrive (a FIFO ring). Last returns the
// newest snapshot, or the zero Dump when nothing has been captured yet; All
// returns a copy of the retained snapshots oldest-first; Reset clears the
// history, which tests typically call to isolate cases given the shared state.
//
// Parity with the Node original is behavioral rather than API-identical: like
// the ad-hoc request-dump middleware it echoes, this package captures the
// salient request metadata for inspection. It diverges deliberately by storing
// structured Dump values in a bounded ring instead of formatting log lines, and
// by capturing only method, path, and headers (not the request body), which
// keeps it non-consuming and safe to place anywhere in the chain.
package requestdump

import (
	"net/http"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Dump is a captured snapshot of a single request. All fields are populated at
// capture time and are never mutated afterward, so a Dump can be read safely
// once obtained from Last or All.
type Dump struct {
	// Method is the HTTP request method (for example "GET" or "POST") as
	// reported by req.Method() when the request was captured.
	Method string

	// Path is the request URL path (without query string) as reported by
	// req.Path() when the request was captured.
	Path string

	// Headers is an independent clone of the request headers taken at capture
	// time. Because it is cloned, later mutation of the live request headers
	// does not affect this snapshot. It may be nil only for a zero Dump.
	Headers http.Header

	// Time is the wall-clock time at which the request was captured, set to
	// time.Now() as the middleware recorded the snapshot.
	Time time.Time
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
