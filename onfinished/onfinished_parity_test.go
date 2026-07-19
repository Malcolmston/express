package onfinished

// Upstream parity tests for jshttp/on-finished.
//
// Source of truth (fetched via raw.githubusercontent.com):
//   https://raw.githubusercontent.com/jshttp/on-finished/master/test/test.js
//   https://raw.githubusercontent.com/jshttp/on-finished/master/index.js
//
// The Node library exposes two functions:
//   onFinished(msg, listener) — attaches a listener fired exactly once when the
//     message (request or response) finishes cleanly, errors, or is aborted; the
//     listener is invoked as listener(err, msg) where err is null on a clean
//     finish and an Error on teardown.
//   onFinished.isFinished(msg) — reports whether the message is already finished.
//
// Node infers state from stream "finish"/"end"/"close"/"error" events. Go's
// net/http exposes no such events, so this port re-frames the same contract
// around an explicit Tracker: New(w) wraps the writer, OnFinished(fn) registers
// a completion callback, Done(err) signals completion, and IsFinished() reports
// state. Each parity case below encodes a concrete input->output vector taken
// from an upstream it(...) block, translated to the Tracker API. Cases whose
// upstream behavior has no Tracker analogue (async setImmediate deferral, the
// standalone isFinished(msg) type-detection returning undefined / throwing
// TypeError, and the listener's second msg argument) are recorded as gaps in the
// task notes rather than invented here.

import (
	"errors"
	"net/http/httptest"
	"testing"
)

// Upstream: onFinished(res) "when the response finishes" -> "should fire the
// callback" and "should include the response object" (err is null / !err).
// Vector: Done(nil) invokes the registered listener exactly once with a nil
// error.
func TestParityResponseFinishFiresCallbackNilError(t *testing.T) {
	tr := New(httptest.NewRecorder())
	calls := 0
	var got error = errors.New("sentinel")
	tr.OnFinished(func(err error) {
		calls++
		got = err
	})
	tr.Done(nil)
	if calls != 1 {
		t.Fatalf("expected listener fired once, got %d", calls)
	}
	if got != nil {
		t.Fatalf("expected nil error on clean finish, got %v", got)
	}
}

// Upstream: onFinished(res) "when called after finish" -> "should fire when
// called after finish". A listener attached after the message has already
// finished still runs. Vector: OnFinished registered after Done runs
// immediately with the recorded error.
func TestParityListenerAfterFinishRunsImmediately(t *testing.T) {
	tr := New(httptest.NewRecorder())
	tr.Done(nil)
	fired := false
	tr.OnFinished(func(err error) { fired = true })
	if !fired {
		t.Fatal("listener registered after finish should fire immediately")
	}
}

// Upstream: onFinished(res) "when response errors" -> "should fire with error"
// (assert.ok(err)). Vector: Done(err) invokes the listener with a truthy
// (non-nil) error, and it is the exact recorded value.
func TestParityResponseErrorFiresWithError(t *testing.T) {
	tr := New(httptest.NewRecorder())
	want := errors.New("ECONNRESET")
	var got error
	fired := false
	tr.OnFinished(func(err error) {
		fired = true
		got = err
	})
	tr.Done(want)
	if !fired {
		t.Fatal("expected listener to fire on error finish")
	}
	if got != want {
		t.Fatalf("expected recorded error %v to propagate, got %v", want, got)
	}
}

// Upstream: isFinished(res) "should be false before response finishes"
// (assert.ok(!onFinished.isFinished(res))). Vector: IsFinished() is false
// before Done is called.
func TestParityIsFinishedFalseBeforeFinish(t *testing.T) {
	tr := New(httptest.NewRecorder())
	if tr.IsFinished() {
		t.Fatal("IsFinished should be false before Done")
	}
}

// Upstream: isFinished(res) "should be true after response finishes"
// (assert.ok(onFinished.isFinished(res)) inside the finish listener). Vector:
// IsFinished() is true once Done(nil) has run, including as observed from
// within the callback.
func TestParityIsFinishedTrueAfterFinish(t *testing.T) {
	tr := New(httptest.NewRecorder())
	insideCallback := false
	tr.OnFinished(func(err error) {
		insideCallback = tr.IsFinished()
	})
	tr.Done(nil)
	if !insideCallback {
		t.Fatal("IsFinished should be true when observed inside the finish callback")
	}
	if !tr.IsFinished() {
		t.Fatal("IsFinished should be true after Done")
	}
}

// Upstream: isFinished(res) "when response errors" -> "should return true"
// (assert.ok(err) && assert.ok(onFinished.isFinished(res))). Vector: after an
// error finish, IsFinished() is true.
func TestParityIsFinishedTrueAfterError(t *testing.T) {
	tr := New(httptest.NewRecorder())
	tr.OnFinished(func(err error) {})
	tr.Done(errors.New("boom"))
	if !tr.IsFinished() {
		t.Fatal("IsFinished should be true after an error finish")
	}
}

// Upstream: isFinished(res) "when the response aborts" -> "should return true"
// where the listener asserts ifError(err) (err null) yet isFinished is true.
// Vector: an abort that completes cleanly (Done(nil)) still marks the tracker
// finished, and the listener sees a nil error.
func TestParityAbortFinishesWithNilError(t *testing.T) {
	tr := New(httptest.NewRecorder())
	var got error = errors.New("sentinel")
	tr.OnFinished(func(err error) { got = err })
	tr.Done(nil)
	if got != nil {
		t.Fatalf("expected nil error on abort clean finish, got %v", got)
	}
	if !tr.IsFinished() {
		t.Fatal("IsFinished should be true after abort")
	}
}

// Upstream: onFinished(res) "when using keep-alive" -> "should fire for each
// response" whose listener fails with new Error('fired twice on same req') if
// invoked more than once per response. Vector: a redundant Done is a no-op, so
// each listener fires exactly once.
func TestParityFiresExactlyOncePerResponse(t *testing.T) {
	tr := New(httptest.NewRecorder())
	calls := 0
	tr.OnFinished(func(err error) { calls++ })
	tr.Done(nil)
	tr.Done(errors.New("ignored second finish"))
	if calls != 1 {
		t.Fatalf("listener must fire exactly once per response, got %d", calls)
	}
}

// Upstream: onFinished(res) "when calling many times on same response" ->
// "should not print warnings": 400 listeners are attached plus a final one, and
// all are expected to run without an EventEmitter max-listeners warning. Vector:
// every one of many registered listeners fires exactly once on a single Done.
func TestParityManyListenersEachFireOnce(t *testing.T) {
	tr := New(httptest.NewRecorder())
	const n = 400
	counts := make([]int, n)
	for i := 0; i < n; i++ {
		i := i
		tr.OnFinished(func(err error) { counts[i]++ })
	}
	tr.Done(nil)
	for i, c := range counts {
		if c != 1 {
			t.Fatalf("listener %d fired %d times, want 1", i, c)
		}
	}
}

// Upstream: onFinished's ee-first backing invokes listeners in the order they
// were attached (see onFinished(req) pipelined ordering assertions). Vector:
// multiple listeners fire in registration order on Done.
func TestParityListenersFireInRegistrationOrder(t *testing.T) {
	tr := New(httptest.NewRecorder())
	var order []int
	for i := 0; i < 3; i++ {
		i := i
		tr.OnFinished(func(err error) { order = append(order, i) })
	}
	tr.Done(nil)
	if len(order) != 3 || order[0] != 0 || order[1] != 1 || order[2] != 2 {
		t.Fatalf("expected registration order [0 1 2], got %v", order)
	}
}
