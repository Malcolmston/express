// Package maintenance provides middleware that puts an application into
// maintenance mode. While enabled, every request is short-circuited with a 503
// Service Unavailable response and a Retry-After header. It plays the role that
// the Node/Express ecosystem fills with packages such as express-maintenance
// or a hand-rolled "under construction" gate, or that operators implement at
// the reverse proxy with an HTTP 503 maintenance page — here delivered as a
// dependency-free handler that also exposes a programmatic runtime toggle.
//
// Reach for this middleware when you need to take an application offline
// gracefully: during a database migration, a deploy that must not receive
// traffic, an incident where you want to shed load while keeping the process
// alive, or a scheduled maintenance window. Returning a well-formed 503 with a
// Retry-After header is friendlier to clients, crawlers, and monitoring than
// dropping connections, and because the gate flips at runtime you can enable or
// disable it without restarting or redeploying the server.
//
// Install it as one of the first handlers via app.Use so that it fronts the
// entire route table. On each request the handler evaluates whether maintenance
// mode is active: if EnabledFunc is set it calls that predicate, otherwise it
// atomically loads the *int32 flag. When active, it sets the Retry-After header,
// writes the configured status 503 with the message body, and returns without
// calling next() — short-circuiting every downstream handler. When inactive it
// simply calls next() and the request proceeds normally. The middleware reads
// no request state and, in the pass-through case, writes nothing.
//
// New returns both the handler and a *Toggle. Toggle.Set(true/false) and
// Toggle.Enabled() drive and inspect the internal flag using sync/atomic, so it
// is safe to call them concurrently from another goroutine (for example an
// admin HTTP endpoint or a signal handler) while requests are in flight. If you
// supply Options.Enabled with your own *int32 the Toggle wraps that same flag;
// if you supply Options.EnabledFunc it takes precedence over both the flag and
// the Toggle, letting you derive the state from any source (a config value, a
// feature-flag service, a clock). Options.Message defaults to "Service
// Unavailable for maintenance" and Options.RetryAfter defaults to 300 seconds
// when set to a value <= 0.
//
// Semantically the gate is all-or-nothing: it does not exempt health checks or
// specific paths, so if you need the load balancer's probe to keep succeeding,
// mount the middleware on a sub-router or wrap it with your own path check.
// From a security and correctness standpoint 503 with Retry-After is the
// standards-correct signal that keeps well-behaved crawlers from de-indexing the
// site. Compared with the Node originals, this port keeps the core behavior
// (flip a flag, answer everything with 503 + Retry-After) while trading
// JavaScript conventions for Go ones: an atomic *int32 flag and Toggle instead
// of a mutable boolean on the app, an optional EnabledFunc predicate for custom
// sources, and plain integer-seconds Retry-After rather than a formatted date.
package maintenance

import (
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/malcolmston/express"
)

// Options configures the maintenance middleware.
type Options struct {
	// Enabled is an atomically-accessed flag: a non-zero value activates
	// maintenance mode. When nil an internal flag is used, toggled via the
	// returned Toggle. Ignored if EnabledFunc is set.
	Enabled *int32
	// EnabledFunc, when non-nil, is consulted on each request to decide whether
	// maintenance mode is active. It takes precedence over Enabled.
	EnabledFunc func() bool
	// Message is the response body sent while in maintenance mode. Defaults to
	// "Service Unavailable for maintenance".
	Message string
	// RetryAfter is the number of seconds advertised in the Retry-After header.
	// Values <= 0 default to 300.
	RetryAfter int
}

// Toggle enables or disables maintenance mode for middleware constructed
// without a custom EnabledFunc.
type Toggle struct {
	flag *int32
}

// Set turns maintenance mode on (true) or off (false).
func (t *Toggle) Set(on bool) {
	var v int32
	if on {
		v = 1
	}
	atomic.StoreInt32(t.flag, v)
}

// Enabled reports whether maintenance mode is currently on.
func (t *Toggle) Enabled() bool { return atomic.LoadInt32(t.flag) != 0 }

// New returns maintenance middleware and a Toggle for controlling it. When
// Options.EnabledFunc is supplied the Toggle still operates its own flag but the
// predicate governs the middleware.
func New(opts ...Options) (express.Handler, *Toggle) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Message == "" {
		o.Message = "Service Unavailable for maintenance"
	}
	if o.RetryAfter <= 0 {
		o.RetryAfter = 300
	}
	flag := o.Enabled
	if flag == nil {
		flag = new(int32)
	}
	toggle := &Toggle{flag: flag}
	retry := strconv.Itoa(o.RetryAfter)

	isEnabled := func() bool {
		if o.EnabledFunc != nil {
			return o.EnabledFunc()
		}
		return atomic.LoadInt32(flag) != 0
	}

	handler := func(req *express.Request, res *express.Response, next express.Next) {
		if isEnabled() {
			res.Set("Retry-After", retry)
			res.Status(http.StatusServiceUnavailable).Send(o.Message)
			return
		}
		next()
	}
	return handler, toggle
}
