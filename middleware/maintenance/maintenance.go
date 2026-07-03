// Package maintenance provides middleware that puts an application into
// maintenance mode. While enabled, every request is short-circuited with a 503
// Service Unavailable response and a Retry-After header. The mode can be
// toggled at runtime, either through a shared *int32 flag or a custom
// predicate.
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
