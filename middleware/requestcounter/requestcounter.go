// Package requestcounter provides middleware that maintains a per-process
// sequential counter of the total number of requests handled. It is the
// express analogue of the lightweight in-memory request tallies people wire up
// in Node with a closure over a counter variable (the pattern behind minimal
// hit counters and the request-count portion of ad-hoc metrics middleware),
// packaged here as a constructor that returns both the middleware and a
// thread-safe accessor for the current total.
//
// Use it when you want a cheap, dependency-free measure of load — a health or
// debug endpoint that reports how many requests the process has served, a smoke
// test that asserts traffic reached the app, or a coarse building block for
// homegrown metrics. It is intentionally minimal: for labelled counters,
// histograms, or Prometheus-style export you should reach for the metrics
// middleware instead. The counter lives only in memory and is per-process, so
// it resets to zero on restart and is not shared across instances.
//
// Mechanically the middleware sits early in the chain and, for every request,
// performs a single atomic increment (atomic.AddInt64) and then calls next() to
// continue. It reads and writes no headers, does not touch the request or
// response body, and never short-circuits, so its position in the chain only
// affects which requests are counted — placing it before a filter that may
// abort means aborted requests are still counted, since the increment happens
// before next() runs. The returned accessor performs an atomic load, so it may
// be called concurrently from other goroutines (for example an HTTP handler
// exposing the count) without additional locking.
//
// The counter is scoped to a single New call: each invocation closes over its
// own private int64, so independent counters do not interfere and you can run
// several (for example one global and one mounted on a subset of routes). There
// are no options, no defaults to configure, and no upper bound short of int64
// overflow, which is not a practical concern. Because the count includes every
// request the handler observes, it reflects raw ingress rather than successful
// responses; nothing decrements it.
//
// Parity with the Node original is behavioral: like the closure-based counters
// it mirrors, this package guarantees a monotonically increasing, concurrency-
// safe total exposed through an accessor. The two-return-value shape (handler
// plus accessor) is the idiomatic Go expression of the JavaScript closure that
// keeps the count private while still allowing reads, and it avoids the shared
// mutable global that a naive port would use.
package requestcounter

import (
	"sync/atomic"

	"github.com/malcolmston/express"
)

// New returns request-counting middleware together with an accessor that
// reports the number of requests observed so far.
func New() (express.Handler, func() int64) {
	var count int64
	handler := func(req *express.Request, res *express.Response, next express.Next) {
		atomic.AddInt64(&count, 1)
		next()
	}
	accessor := func() int64 { return atomic.LoadInt64(&count) }
	return handler, accessor
}
