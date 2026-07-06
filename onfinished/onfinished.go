// Package onfinished registers callbacks that run when an HTTP response
// finishes, mirroring the behavior of the npm on-finished library
// (https://www.npmjs.com/package/on-finished) using only the Go standard
// library. In Node, on-finished lets middleware attach a listener that fires
// exactly once when a response has been fully sent or a request has been fully
// read — or when either is torn down early by an error — so that cleanup logic
// (closing files, releasing locks, writing access logs, flushing metrics) runs
// deterministically at the end of the exchange.
//
// The Node API centers on a single function, onFinished(msg, listener), that
// hooks into the stream's "finish", "end", "close" and "error" events and
// guarantees the listener is invoked one time with an optional error argument.
// Go's net/http does not expose those lifecycle events directly, so this port
// re-frames the same contract around an explicit Tracker value that wraps an
// http.ResponseWriter. You create a Tracker with New, register any number of
// completion callbacks with OnFinished, and signal completion yourself by
// calling Done when the handler returns or the connection is closed.
//
// A Tracker embeds the http.ResponseWriter it wraps, so it is a drop-in
// replacement inside a handler: WriteHeader, Write and Flush all pass straight
// through to the underlying writer, and the Tracker satisfies both
// http.ResponseWriter and http.Flusher (Flush is a no-op when the underlying
// writer is not a Flusher). This lets you thread the Tracker through existing
// middleware unchanged while still observing when the response is done.
//
// The completion semantics match on-finished closely. Done records the final
// error (nil for a clean finish) and invokes every registered callback exactly
// once, in registration order; it is idempotent, so extra Done calls after the
// first are ignored and any error they carry is discarded. Registering a
// callback with OnFinished after the response has already finished does not drop
// it on the floor — the callback runs immediately with the previously recorded
// error, exactly as Node invokes a late listener synchronously. IsFinished
// reports whether Done has been called, paralleling on-finished's isFinished
// helper.
//
// Because completion is signaled explicitly rather than inferred from stream
// events, the package is easy to drive in tests with httptest and free of the
// races that come from hooking low-level socket state. The trade-off versus the
// Node original is that the caller is responsible for calling Done at the right
// moment (typically deferred at the top of a handler, or from a wrapping
// middleware once the inner handler returns); in exchange the behavior is fully
// deterministic and the Tracker is safe for concurrent use, guarding its
// finished flag, recorded error and callback list with a mutex.
package onfinished

import (
	"net/http"
	"sync"
)

// Tracker wraps an http.ResponseWriter and tracks whether the response has
// finished. Callbacks registered via OnFinished are invoked once when Done is
// called.
type Tracker struct {
	http.ResponseWriter

	mu        sync.Mutex
	finished  bool
	err       error
	callbacks []func(err error)
}

// New wraps w in a Tracker. The returned Tracker implements
// http.ResponseWriter and passes WriteHeader, Write, and Flush through to w.
func New(w http.ResponseWriter) *Tracker {
	return &Tracker{ResponseWriter: w}
}

// OnFinished registers fn to be called when the response finishes. If the
// response has already finished, fn is invoked immediately with the recorded
// error.
func (t *Tracker) OnFinished(fn func(err error)) {
	t.mu.Lock()
	if t.finished {
		err := t.err
		t.mu.Unlock()
		fn(err)
		return
	}
	t.callbacks = append(t.callbacks, fn)
	t.mu.Unlock()
}

// Done marks the response as finished and invokes all registered callbacks
// with err. It is idempotent: subsequent calls do nothing.
func (t *Tracker) Done(err error) {
	t.mu.Lock()
	if t.finished {
		t.mu.Unlock()
		return
	}
	t.finished = true
	t.err = err
	cbs := t.callbacks
	t.callbacks = nil
	t.mu.Unlock()

	for _, fn := range cbs {
		fn(err)
	}
}

// IsFinished reports whether the response has finished.
func (t *Tracker) IsFinished() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.finished
}

// Flush implements http.Flusher when the underlying ResponseWriter supports
// it; otherwise it is a no-op.
func (t *Tracker) Flush() {
	if f, ok := t.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
