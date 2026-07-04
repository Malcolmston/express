// Package throttle provides a faithful port of lodash's throttle.
//
// A throttled function invokes the wrapped function at most once per every
// wait duration. It supports leading and trailing invocation, cancellation,
// and immediate flushing.
//
// The clock is injectable via WithClock so that tests can advance time
// manually instead of relying on real timers.
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

func (realClock) Now() time.Time { return time.Now() }

func (realClock) AfterFunc(d time.Duration, fn func()) Timer {
	return realTimer{time.AfterFunc(d, fn)}
}

type realTimer struct{ t *time.Timer }

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
