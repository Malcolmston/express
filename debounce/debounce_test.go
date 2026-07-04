package debounce

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

// Advance moves the clock forward by d, firing any timers whose deadline is
// reached, in chronological order.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	target := c.now.Add(d)
	c.mu.Unlock()

	for {
		c.mu.Lock()
		// Find the earliest non-stopped timer due at or before target.
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
		// Remove stopped/fired timers to keep the slice small.
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

func TestTrailingSingleCall(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	d.Call()
	if count != 0 {
		t.Fatalf("should not fire immediately, got %d", count)
	}
	clk.Advance(100 * time.Millisecond)
	if count != 1 {
		t.Fatalf("want 1 after wait, got %d", count)
	}
}

func TestCoalesceRapidCalls(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	for i := 0; i < 5; i++ {
		d.Call()
		clk.Advance(10 * time.Millisecond)
	}
	if count != 0 {
		t.Fatalf("should not have fired yet, got %d", count)
	}
	clk.Advance(100 * time.Millisecond)
	if count != 1 {
		t.Fatalf("want single coalesced call, got %d", count)
	}
}

func TestLeadingEdge(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ },
		WithClock(clk), WithLeading(true), WithTrailing(false))

	d.Call()
	if count != 1 {
		t.Fatalf("leading should fire immediately, got %d", count)
	}
	d.Call()
	if count != 1 {
		t.Fatalf("second call within wait should not fire, got %d", count)
	}
	clk.Advance(100 * time.Millisecond)
	if count != 1 {
		t.Fatalf("trailing disabled, want 1, got %d", count)
	}
}

func TestLeadingAndTrailing(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ },
		WithClock(clk), WithLeading(true), WithTrailing(true))

	d.Call() // leading fires
	d.Call() // schedules trailing
	if count != 1 {
		t.Fatalf("want 1 after leading, got %d", count)
	}
	clk.Advance(100 * time.Millisecond)
	if count != 2 {
		t.Fatalf("want 2 (leading+trailing), got %d", count)
	}
}

func TestCancel(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	d.Call()
	d.Cancel()
	clk.Advance(200 * time.Millisecond)
	if count != 0 {
		t.Fatalf("cancel should prevent invocation, got %d", count)
	}
}

func TestFlush(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ }, WithClock(clk))

	d.Call()
	d.Flush()
	if count != 1 {
		t.Fatalf("flush should invoke immediately, got %d", count)
	}
	clk.Advance(200 * time.Millisecond)
	if count != 1 {
		t.Fatalf("no double-invoke after flush, got %d", count)
	}
}

func TestFlushNoPending(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ }, WithClock(clk))
	d.Flush()
	if count != 0 {
		t.Fatalf("flush with nothing pending should not invoke, got %d", count)
	}
}

func TestMaxWait(t *testing.T) {
	clk := newFakeClock()
	var count int
	d := New(100*time.Millisecond, func() { count++ },
		WithClock(clk), WithMaxWait(200*time.Millisecond))

	// Keep calling just before wait elapses; maxWait should force an invoke.
	for i := 0; i < 30; i++ {
		d.Call()
		clk.Advance(50 * time.Millisecond)
	}
	if count == 0 {
		t.Fatalf("maxWait should have forced at least one invocation")
	}
}

func TestPending(t *testing.T) {
	clk := newFakeClock()
	d := New(100*time.Millisecond, func() {}, WithClock(clk))
	if d.Pending() {
		t.Fatal("should not be pending before any call")
	}
	d.Call()
	if !d.Pending() {
		t.Fatal("should be pending after call")
	}
	clk.Advance(100 * time.Millisecond)
	if d.Pending() {
		t.Fatal("should not be pending after fire")
	}
}
