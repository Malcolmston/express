package throttle

import (
	"sort"
	"sync"
	"testing"
	"time"
)

// fakeClock is a manually-advanced clock for deterministic tests.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
	nextID int
}

type fakeTimer struct {
	id      int
	fireAt  time.Time
	fn      func()
	stopped bool
	clock   *fakeClock
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Unix(0, 0)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) AfterFunc(d time.Duration, fn func()) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID++
	t := &fakeTimer{id: c.nextID, fireAt: c.now.Add(d), fn: fn, clock: c}
	c.timers = append(c.timers, t)
	return t
}

func (t *fakeTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	if t.stopped {
		return false
	}
	t.stopped = true
	return true
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	target := c.now.Add(d)
	c.mu.Unlock()

	for {
		c.mu.Lock()
		var due []*fakeTimer
		for _, t := range c.timers {
			if !t.stopped && !t.fireAt.After(target) {
				due = append(due, t)
			}
		}
		if len(due) == 0 {
			c.now = target
			c.mu.Unlock()
			return
		}
		sort.Slice(due, func(i, j int) bool {
			if due[i].fireAt.Equal(due[j].fireAt) {
				return due[i].id < due[j].id
			}
			return due[i].fireAt.Before(due[j].fireAt)
		})
		next := due[0]
		next.stopped = true
		c.now = next.fireAt
		remaining := c.timers[:0]
		for _, t := range c.timers {
			if !t.stopped {
				remaining = append(remaining, t)
			}
		}
		c.timers = remaining
		c.mu.Unlock()
		next.fn()
	}
}

func TestLeadingInvokeImmediately(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	th.Call()
	if count != 1 {
		t.Fatalf("leading call should fire immediately, got %d", count)
	}
}

func TestAtMostOncePerInterval(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	// Burst of calls within a single interval.
	for i := 0; i < 10; i++ {
		th.Call()
		clk.Advance(5 * time.Millisecond)
	}
	// Leading fired once; only one invocation so far in this interval.
	if count != 1 {
		t.Fatalf("want 1 during interval (leading only), got %d", count)
	}
	// Advancing past the interval fires the trailing call.
	clk.Advance(100 * time.Millisecond)
	if count != 2 {
		t.Fatalf("want 2 (leading+trailing), got %d", count)
	}
}

func TestTrailingFiresAtEnd(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	th.Call() // leading -> 1
	th.Call() // schedules trailing
	clk.Advance(100 * time.Millisecond)
	if count != 2 {
		t.Fatalf("trailing should fire at end, got %d", count)
	}
}

func TestLeadingDisabled(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ },
		WithClock(clk), WithLeading(false), WithTrailing(true))

	th.Call()
	if count != 0 {
		t.Fatalf("leading disabled should not fire immediately, got %d", count)
	}
	clk.Advance(100 * time.Millisecond)
	if count != 1 {
		t.Fatalf("trailing should fire, got %d", count)
	}
}

func TestTrailingDisabled(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ },
		WithClock(clk), WithLeading(true), WithTrailing(false))

	th.Call() // leading -> 1
	th.Call() // would be trailing but disabled
	clk.Advance(200 * time.Millisecond)
	if count != 1 {
		t.Fatalf("trailing disabled, want 1, got %d", count)
	}
}

func TestSpacedCallsEachInvoke(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	th.Call() // t=0 -> 1
	clk.Advance(150 * time.Millisecond)
	th.Call() // t=150, interval elapsed -> leading fires -> 2
	if count != 2 {
		t.Fatalf("spaced calls should each invoke, got %d", count)
	}
}

func TestCancel(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	th.Call() // leading -> 1
	th.Call() // schedule trailing
	th.Cancel()
	clk.Advance(200 * time.Millisecond)
	if count != 1 {
		t.Fatalf("cancel should drop trailing, got %d", count)
	}
}

func TestFlush(t *testing.T) {
	clk := newFakeClock()
	var count int
	th := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	th.Call() // leading -> 1
	th.Call() // schedule trailing
	th.Flush()
	if count != 2 {
		t.Fatalf("flush should invoke pending trailing, got %d", count)
	}
	clk.Advance(200 * time.Millisecond)
	if count != 2 {
		t.Fatalf("no double invoke after flush, got %d", count)
	}
}
