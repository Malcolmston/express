// Package circuitbreaker provides middleware implementing the circuit-breaker
// pattern. It observes the status of downstream responses through a
// ResponseWriter wrapper; once a configured number of consecutive 5xx
// responses occur, the circuit "opens" and subsequent requests are
// short-circuited with a 503 response without invoking the downstream chain.
// After a cooldown period elapses the circuit moves to a half-open state,
// allowing a single trial request whose outcome closes or re-opens it. It is
// the express framework's Go analogue of Node resiliency middleware such as
// opossum and express-circuit-breaker.
//
// Use this middleware to stop a failing dependency from taking the rest of the
// service down with it. When a route proxies to a database, a third-party API,
// or another microservice that has started returning errors, hammering it with
// more requests only deepens the outage and ties up goroutines waiting on
// timeouts. Opening the circuit sheds that load instantly, returns a fast 503
// to callers instead of a slow failure, and gives the downstream dependency
// room to recover. Mount it around the specific routes whose failures you want
// to isolate rather than globally, since one breaker tracks one failure stream.
//
// Operationally the breaker wraps the downstream chain. Before each request it
// consults its state: while closed (and while half-open trials are permitted)
// it installs a statusWriter over res.Writer, calls next(), restores the
// original writer, and records the first status code the handler wrote. A
// status of 500 or greater counts as a failure and increments a counter; any
// status below 500 is treated as success and immediately resets the counter
// and closes the circuit. Once the consecutive-failure counter reaches the
// threshold the circuit opens and stamps the current time. The breaker is
// guarded by a mutex, so a single instance may be shared safely across
// concurrent requests.
//
// While the circuit is open the middleware short-circuits: it responds with 503
// Service Unavailable carrying Options.Message and never calls next(), so the
// ailing dependency sees no traffic. When at least Cooldown has elapsed since
// the circuit opened, the next request transitions the breaker to half-open and
// is allowed through as a single trial. If that trial succeeds the circuit
// closes and normal traffic resumes; if it fails the circuit re-opens
// immediately and the cooldown clock restarts. Note that the failure signal is
// the HTTP status code the handler writes — a handler that recovers from a
// panic into a 500, or explicitly sends a 5xx, trips the breaker, whereas a
// handler that never writes a status is recorded as 200 OK.
//
// Configuration lives in Options. Threshold is the number of consecutive 5xx
// responses that trips the breaker and defaults to 5 when set to zero or less;
// Cooldown is how long the circuit stays open before permitting a trial and
// defaults to 30 seconds; Message is the 503 body and defaults to "Service
// Unavailable"; and Now supplies the clock, defaulting to time.Now but
// overridable so time-based behaviour is fully deterministic in tests. Compared
// with richer Node libraries this port is intentionally minimal: it keys solely
// off response status rather than latency, timeouts, or thrown errors, exposes
// no rolling window, fallback function, or per-breaker metrics and events, and
// admits exactly one trial request per cooldown in the half-open state.
package circuitbreaker

import (
	"net/http"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the circuit breaker.
type Options struct {
	// Threshold is the number of consecutive 5xx responses that trips the
	// breaker. Values <= 0 default to 5.
	Threshold int
	// Cooldown is how long the circuit stays open before allowing a trial
	// request. Values <= 0 default to 30 seconds.
	Cooldown time.Duration
	// Message is the body returned while the circuit is open. Defaults to
	// "Service Unavailable".
	Message string
	// Now returns the current time. When nil it defaults to time.Now.
	Now func() time.Time
}

type state int

const (
	closed state = iota
	open
	halfOpen
)

type breaker struct {
	mu       sync.Mutex
	state    state
	failures int
	openedAt time.Time
	opts     Options
	now      func() time.Time
}

// allow reports whether a request may proceed to the downstream chain.
func (b *breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.state {
	case open:
		if b.now().Sub(b.openedAt) >= b.opts.Cooldown {
			b.state = halfOpen
			return true
		}
		return false
	default:
		return true
	}
}

// observe records the outcome of a completed request.
func (b *breaker) observe(status int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	failed := status >= 500
	if failed {
		b.failures++
		if b.state == halfOpen || b.failures >= b.opts.Threshold {
			b.state = open
			b.openedAt = b.now()
		}
		return
	}
	// Success resets the breaker.
	b.failures = 0
	b.state = closed
}

// statusWriter records the first status code written to the response.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

// New returns circuit-breaker middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Threshold <= 0 {
		o.Threshold = 5
	}
	if o.Cooldown <= 0 {
		o.Cooldown = 30 * time.Second
	}
	if o.Message == "" {
		o.Message = "Service Unavailable"
	}
	now := o.Now
	if now == nil {
		now = time.Now
	}
	b := &breaker{state: closed, opts: o, now: now}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if !b.allow() {
			res.Status(http.StatusServiceUnavailable).Send(o.Message)
			return
		}
		sw := &statusWriter{ResponseWriter: res.Writer, status: http.StatusOK}
		orig := res.Writer
		res.Writer = sw
		next()
		res.Writer = orig
		b.observe(sw.status)
	}
}
