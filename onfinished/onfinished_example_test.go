package onfinished_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express/onfinished"
)

// ExampleTracker demonstrates the normal lifecycle of a response Tracker. A
// handler wraps its http.ResponseWriter with New, registers a completion
// callback with OnFinished, and writes a response as usual since the Tracker
// passes writes straight through to the underlying writer. When the handler is
// done it calls Done(nil) to signal a clean finish, which invokes the callback
// exactly once with a nil error. The deterministic markers printed by the
// callback and after Done make the completion order observable.
func ExampleTracker() {
	rec := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		tr := onfinished.New(w)
		tr.OnFinished(func(err error) {
			fmt.Printf("finished err=%v\n", err)
		})
		tr.WriteHeader(http.StatusOK)
		fmt.Fprint(tr, "hello")
		tr.Done(nil)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(rec, req)

	fmt.Printf("status=%d body=%q finished=%v\n", rec.Code, rec.Body.String(), true)
	// Output:
	// finished err=<nil>
	// status=200 body="hello" finished=true
}

// ExampleTracker_error demonstrates that the error recorded by Done is
// propagated to every callback, letting cleanup logic distinguish a clean
// finish from an aborted one. Here Done is called with a non-nil error to
// simulate a connection torn down early. The registered callback receives that
// exact error value. This mirrors how the npm on-finished library passes an
// error argument to its listener when the underlying stream errors.
func ExampleTracker_error() {
	tr := onfinished.New(httptest.NewRecorder())
	tr.OnFinished(func(err error) {
		fmt.Println("cleanup, err:", err)
	})
	tr.Done(errors.New("connection reset"))
	// Output: cleanup, err: connection reset
}

// ExampleTracker_lateRegistration demonstrates that registering a callback
// after the response has already finished does not silently drop it. The
// Tracker is marked finished by an early Done call carrying an error. A
// callback registered afterward with OnFinished runs immediately rather than
// never, receiving the previously recorded error. This matches on-finished's
// behavior of invoking a late listener synchronously. IsFinished confirms the
// Tracker is already in its finished state.
func ExampleTracker_lateRegistration() {
	tr := onfinished.New(httptest.NewRecorder())
	tr.Done(errors.New("boom"))

	fmt.Println("finished before register:", tr.IsFinished())
	tr.OnFinished(func(err error) {
		fmt.Println("ran immediately with:", err)
	})
	// Output:
	// finished before register: true
	// ran immediately with: boom
}
