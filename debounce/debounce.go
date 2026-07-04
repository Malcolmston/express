// Package debounce provides a faithful port of lodash's debounce.
//
// A debounced function delays invoking the wrapped function until after wait
// has elapsed since the last time the debounced function was invoked. It
// supports leading and trailing invocation, cancellation, and immediate
// flushing.
//
// The clock is injectable via WithClock so that tests can advance time
// manually instead of relying on real timers.
package debounce

import (
	"sync"
	"time"
)

// Clock abstracts time so that debouncing can be tested with a fake clock.
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

// realClock is the default Clock backed by the standard library time package.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (realClock) AfterFunc(d time.Duration, fn func()) Timer {
	return realTimer{time.AfterFunc(d, fn)}
}

type realTimer struct{ t *time.Timer }

func (r realTimer) Stop() bool { return r.t.Stop() }

// Option configures a Debouncer.
type Option func(*Debouncer)

// WithLeading enables or disables invoking on the leading edge of the timeout.
// It is disabled by default.
func WithLeading(leading bool) Option {
	return func(d *Debouncer) { d.leading = leading }
}

// WithTrailing enables or disables invoking on the trailing edge of the
// timeout. It is enabled by default.
func WithTrailing(trailing bool) Option {
	return func(d *Debouncer) { d.trailing = trailing }
}

// WithMaxWait sets the maximum time fn is allowed to be delayed before it is
// invoked, mirroring lodash's maxWait option. A zero value disables it.
func WithMaxWait(maxWait time.Duration) Option {
	return func(d *Debouncer) {
		d.maxWait = maxWait
		d.hasMaxWait = maxWait > 0
	}
}

// WithClock injects a custom Clock, primarily for testing.
func WithClock(c Clock) Option {
	return func(d *Debouncer) { d.clock = c }
}

// Debouncer is a debounced wrapper around a function.
type Debouncer struct {
	mu       sync.Mutex
	wait     time.Duration
	maxWait  time.Duration
	fn       func()
	clock    Clock
	leading  bool
	trailing bool

	hasMaxWait bool

	timer      Timer
	lastCall   time.Time // time of the most recent Call
	lastInvoke time.Time // time of the most recent invocation of fn
	hasLast    bool      // whether lastCall is set for a pending trailing call
	pending    bool      // whether there is a pending trailing invocation
}

// New creates a Debouncer that delays calling fn until wait has elapsed since
// the last Call. By default it invokes on the trailing edge only.
func New(wait time.Duration, fn func(), opts ...Option) *Debouncer {
	d := &Debouncer{
		wait:     wait,
		fn:       fn,
		clock:    realClock{},
		leading:  false,
		trailing: true,
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

// shouldInvoke reports whether enough time has passed to invoke fn.
func (d *Debouncer) shouldInvoke(now time.Time) bool {
	if !d.hasLast {
		return true
	}
	sinceCall := now.Sub(d.lastCall)
	sinceInvoke := now.Sub(d.lastInvoke)
	if sinceCall >= d.wait || sinceCall < 0 {
		return true
	}
	return d.hasMaxWait && sinceInvoke >= d.maxWait
}

// remainingWait returns the time until the trailing edge should fire.
func (d *Debouncer) remainingWait(now time.Time) time.Duration {
	sinceCall := now.Sub(d.lastCall)
	remaining := d.wait - sinceCall
	if d.hasMaxWait {
		maxRemaining := d.maxWait - now.Sub(d.lastInvoke)
		if maxRemaining < remaining {
			remaining = maxRemaining
		}
	}
	return remaining
}

// invoke runs fn and records the invocation time.
func (d *Debouncer) invoke(now time.Time) {
	d.lastInvoke = now
	fn := d.fn
	d.pending = false
	// Call fn without holding the lock to avoid re-entrancy deadlocks.
	if fn != nil {
		fn()
	}
}

// timerExpired is called when the scheduled timer fires.
func (d *Debouncer) timerExpired() {
	d.mu.Lock()
	now := d.clock.Now()
	if d.shouldInvoke(now) {
		d.timer = nil
		if d.trailing && d.pending {
			d.invoke(now)
			d.hasLast = false
			d.mu.Unlock()
			return
		}
		d.pending = false
		d.hasLast = false
		d.mu.Unlock()
		return
	}
	// Restart the timer for the remaining time.
	d.startTimer(d.remainingWait(now))
	d.mu.Unlock()
}

// startTimer schedules the trailing timer. Callers must hold d.mu.
func (d *Debouncer) startTimer(dur time.Duration) {
	if dur < 0 {
		dur = 0
	}
	d.timer = d.clock.AfterFunc(dur, d.timerExpired)
}

// Call invokes the debounced behavior. Rapid successive calls are coalesced.
func (d *Debouncer) Call() {
	d.mu.Lock()
	now := d.clock.Now()
	isInvoking := d.shouldInvoke(now)

	d.lastCall = now
	d.hasLast = true
	d.pending = true

	if isInvoking {
		if d.timer == nil {
			// Leading edge.
			d.lastInvoke = now
			if d.leading {
				d.invoke(now)
			}
			d.startTimer(d.wait)
			d.mu.Unlock()
			return
		}
		if d.hasMaxWait {
			// Handle maxWait by restarting and invoking immediately.
			d.startTimerOrReplace(d.wait)
			d.invoke(now)
			d.mu.Unlock()
			return
		}
	}
	if d.timer == nil {
		d.startTimer(d.wait)
	}
	d.mu.Unlock()
}

// startTimerOrReplace stops any existing timer and starts a new one. Callers
// must hold d.mu.
func (d *Debouncer) startTimerOrReplace(dur time.Duration) {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.startTimer(dur)
}

// Cancel discards any pending trailing invocation.
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.hasLast = false
	d.pending = false
	d.lastInvoke = time.Time{}
	d.mu.Unlock()
}

// Flush immediately invokes any pending trailing call.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	if d.timer == nil && !d.pending {
		d.mu.Unlock()
		return
	}
	now := d.clock.Now()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	if d.trailing && d.pending {
		d.invoke(now)
	}
	d.hasLast = false
	d.pending = false
	d.mu.Unlock()
}

// Pending reports whether an invocation is scheduled.
func (d *Debouncer) Pending() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.timer != nil
}
