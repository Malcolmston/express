// Package plimit provides a faithful port of the npm p-limit library: it runs an
// unbounded number of scheduled functions while capping how many execute
// concurrently at any instant. Where p-limit wraps promise-returning functions
// and resolves them under a concurrency ceiling, this port wraps ordinary Go
// funcs and runs them on goroutines under the same ceiling, using only the
// standard library (sync and sync/atomic).
//
// The motivation is the same one that makes p-limit popular in Node: you often
// have many independent units of work — HTTP requests, file reads, database
// queries — and firing all of them at once would exhaust sockets, file handles,
// memory or a remote service's rate limit. A limiter lets you enqueue every task
// up front for readability while guaranteeing that no more than N run at the same
// moment. The remaining tasks queue and start automatically as running ones
// finish.
//
// Internally the Limiter uses a buffered channel of capacity N as a counting
// semaphore: acquiring a slot sends into the channel (which blocks once N slots
// are taken) and releasing receives from it. A sync.WaitGroup tracks outstanding
// work so Wait can block until everything has finished, and two atomic counters
// expose live introspection via ActiveCount (running now) and PendingCount
// (scheduled but still waiting for a slot).
//
// Two scheduling entry points model p-limit's behaviour from different angles. Go
// is non-blocking: it enqueues the function and returns immediately, so the
// caller can register many tasks in a tight loop and only the goroutines contend
// for slots. Run is back-pressuring: it acquires a slot in the calling goroutine
// before returning, so Run blocks while the limiter is at capacity and returns as
// soon as the task has secured a slot and started. Both ultimately run the
// supplied function on its own goroutine; choose Go to fan out eagerly and Run
// when you want the producer to slow down under load. Call Wait once, after
// scheduling, to join all tasks.
//
// Edge cases and Node parity: New clamps any concurrency argument less than 1
// (including 0 and negatives) up to 1, so the limiter always makes progress and a
// concurrency of 1 fully serializes execution — mirroring p-limit, which requires
// a concurrency of at least 1. Scheduling zero tasks and then calling Wait
// returns immediately. The counters are eventually consistent snapshots intended
// for observability, not for synchronization. Unlike p-limit's promise API there
// is no per-task return value or error channel: functions communicate results
// through variables they close over (guard shared state yourself), and Wait is
// the single completion signal rather than a Promise.all.
package plimit

import (
	"sync"
	"sync/atomic"
)

// Limiter runs scheduled functions with a bounded level of concurrency.
type Limiter struct {
	sem     chan struct{}
	wg      sync.WaitGroup
	active  int64
	pending int64
}

// New creates a Limiter that allows at most concurrency functions to run at
// the same time. A concurrency of 1 or less serializes execution.
func New(concurrency int) *Limiter {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Limiter{
		sem: make(chan struct{}, concurrency),
	}
}

// Go schedules fn to run. It never blocks the caller: the function is queued
// and starts as soon as a concurrency slot becomes available. Use Wait to
// block until all scheduled functions have completed.
func (l *Limiter) Go(fn func()) {
	l.wg.Add(1)
	atomic.AddInt64(&l.pending, 1)
	go func() {
		defer l.wg.Done()
		l.sem <- struct{}{} // acquire a slot (blocks here, not in Go)
		atomic.AddInt64(&l.pending, -1)
		atomic.AddInt64(&l.active, 1)
		defer func() {
			atomic.AddInt64(&l.active, -1)
			<-l.sem // release the slot
		}()
		fn()
	}()
}

// Run schedules fn the same way as Go but blocks the caller until a
// concurrency slot is available before returning. fn itself still runs in its
// own goroutine, so Run returning means fn has started (or is about to start).
func (l *Limiter) Run(fn func()) {
	l.wg.Add(1)
	atomic.AddInt64(&l.pending, 1)
	l.sem <- struct{}{} // acquire in the caller so Run blocks while at capacity
	atomic.AddInt64(&l.pending, -1)
	atomic.AddInt64(&l.active, 1)
	go func() {
		defer l.wg.Done()
		defer func() {
			atomic.AddInt64(&l.active, -1)
			<-l.sem
		}()
		fn()
	}()
}

// Wait blocks until all scheduled functions have finished executing.
func (l *Limiter) Wait() {
	l.wg.Wait()
}

// ActiveCount returns the number of functions currently executing.
func (l *Limiter) ActiveCount() int {
	return int(atomic.LoadInt64(&l.active))
}

// PendingCount returns the number of functions scheduled but not yet started.
func (l *Limiter) PendingCount() int {
	return int(atomic.LoadInt64(&l.pending))
}
