// Package plimit provides a faithful port of the npm p-limit library: it runs
// an unbounded number of scheduled functions while capping the number that
// execute concurrently.
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
