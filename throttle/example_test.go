package throttle_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/throttle"
)

// fakeClock is a manually advanced Clock so the example is fully deterministic
// and does not sleep on real time.
type fakeClock struct {
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	at      time.Time
	fn      func()
	stopped bool
}

func (t *fakeTimer) Stop() bool {
	if t.stopped {
		return false
	}
	t.stopped = true
	return true
}

func (c *fakeClock) Now() time.Time { return c.now }

func (c *fakeClock) AfterFunc(d time.Duration, fn func()) throttle.Timer {
	t := &fakeTimer{at: c.now.Add(d), fn: fn}
	c.timers = append(c.timers, t)
	return t
}

// Advance moves the clock forward and fires any timers that are now due.
func (c *fakeClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
	for _, t := range c.timers {
		if !t.stopped && !t.at.After(c.now) {
			t.stopped = true
			t.fn()
		}
	}
}

// Example demonstrates the leading-and-trailing behavior of a throttle using an
// injected fake clock so the timing is deterministic. The first Call fires the
// wrapped function immediately on the leading edge, taking the counter to 1. A
// second Call inside the same wait window does not fire again but schedules a
// trailing invocation. Advancing the clock past the wait duration fires that
// trailing call, taking the counter to 2. This shows that no matter how many
// times Call is invoked in one window, the function runs at most once for the
// leading edge and once for the trailing edge.
func Example() {
	clock := &fakeClock{now: time.Unix(0, 0)}
	var count int
	th := throttle.New(100*time.Millisecond, func() { count++ }, throttle.WithClock(clock))

	th.Call() // leading edge fires immediately
	th.Call() // within the window: schedules a trailing call
	fmt.Println(count)

	clock.Advance(100 * time.Millisecond) // trailing call fires
	fmt.Println(count)
	// Output:
	// 1
	// 2
}
