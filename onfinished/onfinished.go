// Package onfinished registers callbacks that run when an HTTP response
// finishes, mirroring the behavior of the npm on-finished library in an
// idiomatic, httptest-friendly way.
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
