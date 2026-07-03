// Package circuitbreaker provides middleware implementing the circuit-breaker
// pattern. It observes the status of downstream responses through a
// ResponseWriter wrapper; once a configured number of consecutive 5xx
// responses occur, the circuit "opens" and subsequent requests are
// short-circuited with a 503 response without invoking the downstream chain.
// After a cooldown period elapses the circuit moves to a half-open state,
// allowing a single trial request whose outcome closes or re-opens it.
//
// The clock is injectable via Options.Now, so the time-based behaviour is fully
// deterministic in tests.
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
