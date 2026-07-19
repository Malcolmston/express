package plimit

// Upstream parity tests ported from sindresorhus/p-limit.
//
// Source of the vectors encoded here (fetched 2026-07-19):
//   https://raw.githubusercontent.com/sindresorhus/p-limit/main/test.js
//   https://raw.githubusercontent.com/sindresorhus/p-limit/main/index.js
//
// p-limit is a Promise-based library; this Go port wraps ordinary funcs run on
// goroutines under the same concurrency ceiling. The vectors below reuse the
// EXACT inputs and expected outputs from upstream's AVA test suite and assert
// the concurrency-gating semantics the port implements: a limiter never exceeds
// its concurrency, activeCount/pendingCount report running vs. queued work, a
// concurrency of 1 fully serializes in scheduling order, and a concurrency-gated
// index-preserving map produces upstream's exact result arrays.
//
// Known divergence (recorded, not fixed): upstream's pLimit(0) / pLimit(-1)
// THROW a TypeError ("Expected `concurrency` to be a number from 1 and up").
// This port's New deliberately CLAMPS concurrency < 1 up to 1 (documented in
// plimit.go and relied upon by ExampleNew / TestNewClampsConcurrency). Reversing
// that to a panic would break the port's published API and its existing tests,
// so it is left as-is; see the package notes.

import (
	"sync/atomic"
	"testing"
	"time"
)

// waitUntil busy-waits (bounded) until cond() is true. Mirrors the polling used
// elsewhere in the package to observe counter snapshots deterministically.
func waitUntil(cond func() bool, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(time.Millisecond)
	}
	return cond()
}

// Upstream: test('concurrency: 1', ...) with input [[10,300],[20,200],[30,100]]
// expects Promise.all(...) === [10, 20, 30] — i.e. a concurrency of 1 runs the
// tasks one at a time in scheduling order. Run applies back-pressure (acquires
// the single slot in the caller), so tasks start strictly in order; we record
// each task's value and expect [10, 20, 30].
func TestParityConcurrency1SerializesInOrder(t *testing.T) {
	l := New(1)
	values := []int{10, 20, 30}

	var maxActive int64
	var active int64
	got := make([]int, 0, len(values))

	for _, v := range values {
		v := v
		l.Run(func() {
			cur := atomic.AddInt64(&active, 1)
			for {
				old := atomic.LoadInt64(&maxActive)
				if cur <= old || atomic.CompareAndSwapInt64(&maxActive, old, cur) {
					break
				}
			}
			got = append(got, v) // safe: limit 1 serializes, Run is in-order
			atomic.AddInt64(&active, -1)
		})
	}
	l.Wait()

	if maxActive != 1 {
		t.Fatalf("concurrency 1 must serialize, got max active %d", maxActive)
	}
	if len(got) != 3 || got[0] != 10 || got[1] != 20 || got[2] != 30 {
		t.Fatalf("want [10 20 30], got %v", got)
	}
}

// Upstream: test('concurrency: 4', ...) actually uses concurrency = 5 and 100
// tasks, asserting running <= concurrency at every moment. Port equivalent:
// max observed active must never exceed 5, and all 100 tasks complete.
func TestParityConcurrency5NeverExceedsLimit(t *testing.T) {
	const concurrency = 5
	const tasks = 100
	l := New(concurrency)

	var active, maxActive, completed int64
	for i := 0; i < tasks; i++ {
		l.Go(func() {
			cur := atomic.AddInt64(&active, 1)
			for {
				old := atomic.LoadInt64(&maxActive)
				if cur <= old || atomic.CompareAndSwapInt64(&maxActive, old, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond) // force overlap so slots stay contended
			atomic.AddInt64(&active, -1)
			atomic.AddInt64(&completed, 1)
		})
	}
	l.Wait()

	if maxActive > concurrency {
		t.Fatalf("max active %d exceeded concurrency %d", maxActive, concurrency)
	}
	if maxActive == 0 {
		t.Fatal("expected some concurrency")
	}
	if completed != tasks {
		t.Fatalf("want %d completed, got %d", tasks, completed)
	}
}

// Upstream: test('activeCount and pendingCount properties', ...) with pLimit(5).
// The load-bearing vector: after scheduling 5 immediate + 3 delayed tasks,
// activeCount === 5 and pendingCount === 3; after everything drains both are 0.
func TestParityActiveAndPendingCount(t *testing.T) {
	l := New(5)

	if got := l.ActiveCount(); got != 0 {
		t.Fatalf("initial ActiveCount want 0, got %d", got)
	}
	if got := l.PendingCount(); got != 0 {
		t.Fatalf("initial PendingCount want 0, got %d", got)
	}

	release := make(chan struct{})
	var started int64
	for i := 0; i < 8; i++ { // 5 that will run + 3 that will queue
		l.Go(func() {
			atomic.AddInt64(&started, 1)
			<-release
		})
	}

	if !waitUntil(func() bool { return atomic.LoadInt64(&started) >= 5 }, 500*time.Millisecond) {
		t.Fatal("expected 5 tasks to start")
	}
	if got := l.ActiveCount(); got != 5 {
		t.Fatalf("want ActiveCount 5, got %d", got)
	}
	if got := l.PendingCount(); got != 3 {
		t.Fatalf("want PendingCount 3, got %d", got)
	}

	close(release)
	l.Wait()

	if got := l.ActiveCount(); got != 0 {
		t.Fatalf("want ActiveCount 0 after wait, got %d", got)
	}
	if got := l.PendingCount(); got != 0 {
		t.Fatalf("want PendingCount 0 after wait, got %d", got)
	}
}

// Upstream: test('shared context with a limited provider helper', ...) with
// pLimit(1) schedules two tasks then asserts activeCount === 1, pendingCount === 1.
func TestParitySharedContextActivePending(t *testing.T) {
	l := New(1)
	release := make(chan struct{})
	var started int64

	for i := 0; i < 2; i++ {
		l.Go(func() {
			atomic.AddInt64(&started, 1)
			<-release
		})
	}

	if !waitUntil(func() bool { return atomic.LoadInt64(&started) >= 1 }, 500*time.Millisecond) {
		t.Fatal("expected 1 task to start")
	}
	if got := l.ActiveCount(); got != 1 {
		t.Fatalf("want ActiveCount 1, got %d", got)
	}
	if got := l.PendingCount(); got != 1 {
		t.Fatalf("want PendingCount 1, got %d", got)
	}

	close(release)
	l.Wait()
}

// Upstream: test('runs all tasks asynchronously', ...) with pLimit(3) schedules
// two tasks then asserts activeCount === 2 (both admitted, neither blocked).
func TestParityRunsAllTasksAsynchronously(t *testing.T) {
	l := New(3)
	release := make(chan struct{})
	var started int64

	for i := 0; i < 2; i++ {
		l.Go(func() {
			atomic.AddInt64(&started, 1)
			<-release
		})
	}

	if !waitUntil(func() bool { return atomic.LoadInt64(&started) >= 2 }, 500*time.Millisecond) {
		t.Fatal("expected 2 tasks to start")
	}
	if got := l.ActiveCount(); got != 2 {
		t.Fatalf("want ActiveCount 2, got %d", got)
	}

	close(release)
	l.Wait()
}

// Upstream: test('map', ...) => limit.map([1..7], x => x+1) === [2..8] with
// pLimit(1). The port has no map method, so we drive the same fan-out through
// the limiter, writing each result to its own index to preserve order, and
// assert upstream's exact result array.
func TestParityMapPlusOne(t *testing.T) {
	l := New(1)
	inputs := []int{1, 2, 3, 4, 5, 6, 7}
	results := make([]int, len(inputs))

	for i, v := range inputs {
		i, v := i, v
		l.Go(func() { results[i] = v + 1 })
	}
	l.Wait()

	want := []int{2, 3, 4, 5, 6, 7, 8}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("want %v, got %v", want, results)
		}
	}
}

// Upstream: test('map passes index and preserves order with concurrency', ...)
// pLimit(3), inputs [10,10,10,10,10], mapper (value,index) => value+index with
// shuffled completion order, expects [10,11,12,13,14] in input order.
func TestParityMapPreservesOrderWithConcurrency(t *testing.T) {
	l := New(3)
	inputs := []int{10, 10, 10, 10, 10}
	results := make([]int, len(inputs))

	for i, v := range inputs {
		i, v := i, v
		l.Go(func() {
			// Shuffle completion order like upstream's variable delay.
			time.Sleep(time.Duration(len(inputs)-i) * time.Millisecond)
			results[i] = v + i
		})
	}
	l.Wait()

	want := []int{10, 11, 12, 13, 14}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("want %v, got %v", want, results)
		}
	}
}

// Upstream: test('map accepts an iterable (set)', ...) pLimit(2), inputs {1,2,3,4},
// mapper x => x*2, expects [2,4,6,8]. (Also covers the array-iterator variant,
// which has identical inputs and expected output.)
func TestParityMapIterableDouble(t *testing.T) {
	l := New(2)
	inputs := []int{1, 2, 3, 4}
	results := make([]int, len(inputs))

	for i, v := range inputs {
		i, v := i, v
		l.Go(func() { results[i] = v * 2 })
	}
	l.Wait()

	want := []int{2, 4, 6, 8}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("want %v, got %v", want, results)
		}
	}
}
