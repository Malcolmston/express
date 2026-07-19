// Package throttle rate-limits how often a function may run, invoking it at
// most once per a fixed wait duration. It is a stdlib-only Go port of lodash's
// throttle (https://lodash.com/docs/#throttle), matching that function's
// leading/trailing edge behavior along with its cancel and flush controls; the
// same leading/trailing model is shared by the npm "throttle-debounce" package.
// Throttling is used to tame high-frequency events — scroll and resize
// handlers, keystrokes, progress callbacks, log flushing — so expensive work
// runs at a bounded rate instead of on every event.
//
// A Throttler is created with New(wait, fn, opts...). Calling its Call method
// stands in for "the event fired"; the Throttler decides whether to run fn now,
// schedule it for the end of the current window, or ignore it because fn already
// ran recently. Regardless of how many times Call is invoked inside one wait
// window, fn runs at most once for the leading edge and at most once for the
// trailing edge of that window. This is distinct from debouncing, which resets
// its timer on every call and only fires after activity stops; a throttle
// guarantees steady progress during a sustained burst.
//
// Leading and trailing behavior is configurable and both default to enabled,
// exactly as in lodash. With leading enabled, the first Call in an idle period
// invokes fn immediately. With trailing enabled, if one or more further Calls
// arrive during the wait window, fn is invoked once more when the window
// elapses, so the most recent activity is not lost. Disabling leading via
// WithLeading(false) delays the first invocation to the trailing edge;
// disabling trailing via WithTrailing(false) drops the follow-up call. If both
// are disabled fn never runs. When calls are spaced farther apart than wait,
// each one is treated as a fresh leading edge and invokes fn immediately.
//
// Cancel discards any pending trailing invocation and resets the Throttler to
// an idle state, so the next Call is again treated as a leading edge. Flush
// immediately invokes a pending trailing call (if trailing is enabled and one is
// scheduled) rather than waiting for the timer, and then clears the pending
// state so no double invocation occurs. Pending reports whether a timer is
// currently scheduled. These three methods mirror lodash's cancel, flush, and
// pending helpers.
//
// The clock is injectable via WithClock, which supplies a Clock implementation
// providing Now and AfterFunc plus a Timer with Stop. The default clock uses
// time.Now and time.AfterFunc, so real timers fire fn on their own goroutines;
// a fake clock can be advanced manually to make tests and examples fully
// deterministic without sleeping. A Throttler serializes its own state with an
// internal mutex, so Call, Cancel, Flush, and Pending are safe to invoke from
// multiple goroutines; the wrapped fn must itself be safe for the concurrency
// it may see.
package throttle

import (
	"sync"
	"time"
)

// Clock abstracts time so that throttling can be tested with a fake clock.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
	// AfterFunc waits for the duration to elapse and then calls fn in its own
	// goroutine. It returns a Timer that can be used to cancel the call.
	AfterFunc(d time.Duration, fn func()) Timer
}

// Timer represents a single scheduled call created by Clock.AfterFunc.
type Timer interface {
	// Stop prevents the Timer from firing. It reports whether the call was
	// stopped before it fired.
	Stop() bool
}

type realClock struct{}

// Now implements Clock; it returns the current time from time.Now.
func (realClock) Now() time.Time { return time.Now() }

// AfterFunc implements Clock; it schedules fn to run after d using
// time.AfterFunc and returns a Timer wrapping the resulting *time.Timer.
func (realClock) AfterFunc(d time.Duration, fn func()) Timer {
	return realTimer{time.AfterFunc(d, fn)}
}

type realTimer struct{ t *time.Timer }

// Stop implements Timer; it stops the underlying *time.Timer and reports
// whether the call was stopped before it fired.
func (r realTimer) Stop() bool { return r.t.Stop() }

// Option configures a Throttler.
type Option func(*Throttler)

// WithLeading enables or disables invoking on the leading edge of the timeout.
// It is enabled by default.
func WithLeading(leading bool) Option {
	return func(t *Throttler) { t.leading = leading }
}

// WithTrailing enables or disables invoking on the trailing edge of the
// timeout. It is enabled by default.
func WithTrailing(trailing bool) Option {
	return func(t *Throttler) { t.trailing = trailing }
}

// WithClock injects a custom Clock, primarily for testing.
func WithClock(c Clock) Option {
	return func(t *Throttler) { t.clock = c }
}

// Throttler is a throttled wrapper around a function.
type Throttler struct {
	mu       sync.Mutex
	wait     time.Duration
	fn       func()
	clock    Clock
	leading  bool
	trailing bool

	timer      Timer
	lastCall   time.Time
	lastInvoke time.Time
	hasLast    bool
	pending    bool
}

// New creates a Throttler that invokes fn at most once per wait. By default it
// invokes on both the leading and trailing edges.
func New(wait time.Duration, fn func(), opts ...Option) *Throttler {
	t := &Throttler{
		wait:     wait,
		fn:       fn,
		clock:    realClock{},
		leading:  true,
		trailing: true,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// shouldInvoke reports whether wait has elapsed since the last invocation.
func (t *Throttler) shouldInvoke(now time.Time) bool {
	if !t.hasLast {
		return true
	}
	sinceInvoke := now.Sub(t.lastInvoke)
	sinceCall := now.Sub(t.lastCall)
	// Throttle: leading edge, or wait elapsed since last invoke.
	return sinceInvoke >= t.wait || sinceCall < 0
}

func (t *Throttler) remainingWait(now time.Time) time.Duration {
	remaining := t.wait - now.Sub(t.lastInvoke)
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (t *Throttler) invoke(now time.Time) {
	t.lastInvoke = now
	t.pending = false
	fn := t.fn
	if fn != nil {
		fn()
	}
}

func (t *Throttler) startTimer(d time.Duration) {
	if d < 0 {
		d = 0
	}
	t.timer = t.clock.AfterFunc(d, t.timerExpired)
}

func (t *Throttler) timerExpired() {
	t.mu.Lock()
	now := t.clock.Now()
	if t.shouldInvoke(now) {
		t.timer = nil
		if t.trailing && t.pending {
			t.invoke(now)
			t.hasLast = false
			t.mu.Unlock()
			return
		}
		t.pending = false
		t.hasLast = false
		t.mu.Unlock()
		return
	}
	t.startTimer(t.remainingWait(now))
	t.mu.Unlock()
}

// Call invokes the throttled behavior. It fires at most once per wait.
func (t *Throttler) Call() {
	t.mu.Lock()
	now := t.clock.Now()
	invoking := t.shouldInvoke(now)

	t.lastCall = now
	t.hasLast = true
	t.pending = true

	if invoking {
		if t.timer == nil {
			// Leading edge.
			t.lastInvoke = now
			if t.leading {
				t.invoke(now)
			}
			t.startTimer(t.wait)
			t.mu.Unlock()
			return
		}
	}
	if t.timer == nil {
		t.startTimer(t.wait)
	}
	t.mu.Unlock()
}

// Cancel discards any pending trailing invocation.
func (t *Throttler) Cancel() {
	t.mu.Lock()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	t.hasLast = false
	t.pending = false
	t.lastInvoke = time.Time{}
	t.mu.Unlock()
}

// Flush immediately invokes any pending trailing call.
func (t *Throttler) Flush() {
	t.mu.Lock()
	if t.timer == nil && !t.pending {
		t.mu.Unlock()
		return
	}
	now := t.clock.Now()
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	if t.trailing && t.pending {
		t.invoke(now)
	}
	t.hasLast = false
	t.pending = false
	t.mu.Unlock()
}

// Pending reports whether an invocation is scheduled.
func (t *Throttler) Pending() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.timer != nil
}
