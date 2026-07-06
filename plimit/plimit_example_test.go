package plimit_test

import (
	"fmt"
	"sync/atomic"

	"github.com/malcolmston/express/plimit"
)

// ExampleLimiter demonstrates bounded fan-out with the non-blocking Go method. We
// create a limiter that allows at most two functions to run concurrently and then
// schedule one hundred tasks, each of which atomically adds its index to a shared
// sum. Go never blocks the caller, so all tasks are registered in a tight loop and
// the goroutines contend for the two slots. Wait then blocks until every scheduled
// function has finished. We print the total, which is the deterministic sum 1..100
// regardless of the order in which the tasks actually ran.
func ExampleLimiter() {
	l := plimit.New(2)
	var sum int64
	for i := 1; i <= 100; i++ {
		i := i
		l.Go(func() {
			atomic.AddInt64(&sum, int64(i))
		})
	}
	l.Wait()
	fmt.Println(sum)
	// Output: 5050
}

// ExampleNew shows that New clamps any concurrency less than 1 up to 1, so a
// limiter always makes progress and never deadlocks on a zero or negative
// argument. We build a limiter with New(0) — which behaves like New(1) — and use
// it to serialize a handful of counter increments. Because a concurrency of 1
// runs the tasks one at a time, the final count is simply the number of tasks we
// scheduled. Wait ensures all of them have completed before we read the result.
// The printed value is fully deterministic.
func ExampleNew() {
	l := plimit.New(0) // clamped up to 1
	var count int64
	for i := 0; i < 5; i++ {
		l.Go(func() {
			atomic.AddInt64(&count, 1)
		})
	}
	l.Wait()
	fmt.Println(count)
	// Output: 5
}

// ExampleLimiter_Run demonstrates the back-pressuring Run entry point. Unlike Go,
// Run acquires a concurrency slot in the calling goroutine, so it blocks while the
// limiter is at capacity and returns only once the task has secured a slot and
// started. Here we run three tasks through a limiter of size two; each appends to
// a shared counter guarded by atomics. Because Run applies back-pressure, the
// producer loop naturally paces itself against the running work. After Wait we
// print the number of completed tasks, which is deterministic.
func ExampleLimiter_Run() {
	l := plimit.New(2)
	var done int64
	for i := 0; i < 3; i++ {
		l.Run(func() {
			atomic.AddInt64(&done, 1)
		})
	}
	l.Wait()
	fmt.Println(done)
	// Output: 3
}
