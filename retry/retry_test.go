package retry

import (
	"errors"
	"testing"
	"time"
)

// recordSleep returns a Sleep function that records durations without sleeping.
func recordSleep(rec *[]time.Duration) func(time.Duration) {
	return func(d time.Duration) { *rec = append(*rec, d) }
}

func TestSuccessFirstTry(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	err := Do(func(attempt int) error {
		calls++
		return nil
	}, Options{Retries: 3, Sleep: recordSleep(&sleeps)})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("want 1 call, got %d", calls)
	}
	if len(sleeps) != 0 {
		t.Fatalf("no sleeps expected, got %v", sleeps)
	}
}

func TestSuccessAfterNFailures(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	err := Do(func(attempt int) error {
		calls++
		if attempt < 3 {
			return errors.New("boom")
		}
		return nil
	}, Options{
		Retries:    5,
		Factor:     2,
		MinTimeout: 10 * time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("want 3 calls, got %d", calls)
	}
	// Two failures -> two backoff sleeps: 10ms, 20ms.
	want := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
	if len(sleeps) != len(want) {
		t.Fatalf("want %v sleeps, got %v", want, sleeps)
	}
	for i := range want {
		if sleeps[i] != want[i] {
			t.Fatalf("sleep[%d] = %v, want %v", i, sleeps[i], want[i])
		}
	}
}

func TestBackoffSequence(t *testing.T) {
	var sleeps []time.Duration
	_ = Do(func(attempt int) error {
		return errors.New("always fails")
	}, Options{
		Retries:    4,
		Factor:     2,
		MinTimeout: 100 * time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	// 4 retries -> delays 100, 200, 400, 800.
	want := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	if len(sleeps) != len(want) {
		t.Fatalf("want %d sleeps, got %v", len(want), sleeps)
	}
	for i := range want {
		if sleeps[i] != want[i] {
			t.Fatalf("sleep[%d] = %v, want %v", i, sleeps[i], want[i])
		}
	}
}

func TestMaxTimeoutCap(t *testing.T) {
	var sleeps []time.Duration
	_ = Do(func(attempt int) error {
		return errors.New("fail")
	}, Options{
		Retries:    4,
		Factor:     10,
		MinTimeout: 10 * time.Millisecond,
		MaxTimeout: 100 * time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	// Uncapped would be 10, 100, 1000, 10000ms; capped at 100ms.
	want := []time.Duration{
		10 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
		100 * time.Millisecond,
	}
	if len(sleeps) != len(want) {
		t.Fatalf("want %d sleeps, got %v", len(want), sleeps)
	}
	for i := range want {
		if sleeps[i] != want[i] {
			t.Fatalf("sleep[%d] = %v, want %v", i, sleeps[i], want[i])
		}
	}
}

func TestGivesUpAfterRetries(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	sentinel := errors.New("still failing")
	err := Do(func(attempt int) error {
		calls++
		return sentinel
	}, Options{
		Retries:    3,
		MinTimeout: time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel error, got %v", err)
	}
	// 1 initial + 3 retries = 4 calls; 3 sleeps.
	if calls != 4 {
		t.Fatalf("want 4 calls, got %d", calls)
	}
	if len(sleeps) != 3 {
		t.Fatalf("want 3 sleeps, got %d", len(sleeps))
	}
}

func TestAbortStopsImmediately(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	fatal := errors.New("do not retry")
	err := Do(func(attempt int) error {
		calls++
		return Abort(fatal)
	}, Options{
		Retries:    5,
		MinTimeout: time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	if !errors.Is(err, fatal) {
		t.Fatalf("want unwrapped fatal error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("abort should stop after 1 call, got %d", calls)
	}
	if len(sleeps) != 0 {
		t.Fatalf("abort should not sleep, got %v", sleeps)
	}
}

func TestAbortAfterSomeRetries(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	fatal := errors.New("fatal")
	err := Do(func(attempt int) error {
		calls++
		if attempt < 3 {
			return errors.New("transient")
		}
		return Abort(fatal)
	}, Options{
		Retries:    10,
		MinTimeout: time.Millisecond,
		Sleep:      recordSleep(&sleeps),
	})

	if !errors.Is(err, fatal) {
		t.Fatalf("want fatal, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("want 3 calls, got %d", calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("want 2 sleeps before abort, got %d", len(sleeps))
	}
}

func TestOnRetryCallback(t *testing.T) {
	var sleeps []time.Duration
	var attempts []int
	_ = Do(func(attempt int) error {
		if attempt < 3 {
			return errors.New("fail")
		}
		return nil
	}, Options{
		Retries:    5,
		MinTimeout: time.Millisecond,
		Sleep:      recordSleep(&sleeps),
		OnRetry: func(err error, attempt int) {
			attempts = append(attempts, attempt)
		},
	})

	want := []int{1, 2}
	if len(attempts) != len(want) {
		t.Fatalf("want OnRetry attempts %v, got %v", want, attempts)
	}
	for i := range want {
		if attempts[i] != want[i] {
			t.Fatalf("attempt[%d] = %d, want %d", i, attempts[i], want[i])
		}
	}
}

func TestDefaults(t *testing.T) {
	// Factor and MinTimeout default to 2 and 1s.
	o := Options{}
	if got := o.Backoff(1); got != time.Second {
		t.Fatalf("default first backoff = %v, want 1s", got)
	}
	if got := o.Backoff(2); got != 2*time.Second {
		t.Fatalf("default second backoff = %v, want 2s", got)
	}
}

func TestZeroRetries(t *testing.T) {
	var sleeps []time.Duration
	calls := 0
	err := Do(func(attempt int) error {
		calls++
		return errors.New("fail")
	}, Options{Retries: 0, Sleep: recordSleep(&sleeps)})

	if err == nil {
		t.Fatal("want error")
	}
	if calls != 1 {
		t.Fatalf("want exactly 1 call, got %d", calls)
	}
	if len(sleeps) != 0 {
		t.Fatalf("want no sleeps, got %d", len(sleeps))
	}
}
