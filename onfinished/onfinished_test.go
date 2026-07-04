package onfinished

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPassThrough(t *testing.T) {
	rec := httptest.NewRecorder()
	tr := New(rec)
	tr.WriteHeader(http.StatusTeapot)
	if _, err := tr.Write([]byte("hi")); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected 418, got %d", rec.Code)
	}
	if rec.Body.String() != "hi" {
		t.Fatalf("expected body hi, got %q", rec.Body.String())
	}
}

func TestImplementsResponseWriter(t *testing.T) {
	var _ http.ResponseWriter = New(httptest.NewRecorder())
}

func TestCallbackOnDone(t *testing.T) {
	tr := New(httptest.NewRecorder())
	called := 0
	tr.OnFinished(func(err error) {
		called++
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}
	})
	if tr.IsFinished() {
		t.Fatal("should not be finished yet")
	}
	tr.Done(nil)
	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
	if !tr.IsFinished() {
		t.Fatal("should be finished")
	}
}

func TestDoneIdempotent(t *testing.T) {
	tr := New(httptest.NewRecorder())
	called := 0
	tr.OnFinished(func(err error) { called++ })
	tr.Done(nil)
	tr.Done(errors.New("ignored"))
	if called != 1 {
		t.Fatalf("expected 1 call, got %d", called)
	}
}

func TestMultipleCallbacks(t *testing.T) {
	tr := New(httptest.NewRecorder())
	total := 0
	for i := 0; i < 3; i++ {
		tr.OnFinished(func(err error) { total++ })
	}
	tr.Done(nil)
	if total != 3 {
		t.Fatalf("expected 3, got %d", total)
	}
}

func TestCallbackAfterFinish(t *testing.T) {
	tr := New(httptest.NewRecorder())
	myErr := errors.New("boom")
	tr.Done(myErr)
	var got error
	called := false
	tr.OnFinished(func(err error) {
		called = true
		got = err
	})
	if !called {
		t.Fatal("callback registered after finish should run immediately")
	}
	if got != myErr {
		t.Fatalf("expected recorded error, got %v", got)
	}
}

func TestErrorPropagation(t *testing.T) {
	tr := New(httptest.NewRecorder())
	myErr := errors.New("fail")
	var got error
	tr.OnFinished(func(err error) { got = err })
	tr.Done(myErr)
	if got != myErr {
		t.Fatalf("expected %v, got %v", myErr, got)
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (f *flushRecorder) Flush() { f.flushed = true }

func TestFlushPassThrough(t *testing.T) {
	fr := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	tr := New(fr)
	tr.Flush()
	if !fr.flushed {
		t.Fatal("expected underlying Flush to be called")
	}
}

func TestFlushNoFlusher(t *testing.T) {
	// Should not panic when underlying writer has no Flush.
	tr := New(nopWriter{})
	tr.Flush()
}

type nopWriter struct{}

func (nopWriter) Header() http.Header         { return http.Header{} }
func (nopWriter) Write(b []byte) (int, error) { return len(b), nil }
func (nopWriter) WriteHeader(int)             {}
