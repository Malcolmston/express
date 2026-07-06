// Package retry retries a function with exponential backoff. It is a
// stdlib-only Go port of the npm "retry" package (https://www.npmjs.com/package/retry,
// also published as node-retry) and the higher-level "p-retry" / "async-retry"
// wrappers built on top of it. Those JavaScript libraries repeatedly invoke an
// operation that may fail transiently — a flaky network request, a database
// call, a rate-limited API — pausing between attempts for a delay that grows on
// each failure, and they are the retry engine behind a large part of the Node
// ecosystem. This package brings the same retry-with-backoff contract to Go
// using only the standard library.
//
// The central entry point is Do, which calls fn and, on error, sleeps and calls
// it again until the operation succeeds, the retry budget is exhausted, or an
// AbortError signals that further attempts are pointless. fn receives the
// 1-based attempt number so it can log or vary behavior across attempts. When
// every attempt fails, the error from the final attempt is returned unchanged;
// when an AbortError is returned, its wrapped error (if any) is returned and no
// further attempts are made. This mirrors p-retry, where returning an
// AbortError short-circuits the retry loop.
//
// Backoff follows the node-retry formula: the delay before retry attempt n
// (1-based, where attempt 1 is the pause before the first retry) is
// MinTimeout * Factor^(n-1), capped at MaxTimeout. Factor defaults to 2,
// MinTimeout defaults to one second, and MaxTimeout defaults to effectively no
// cap. With the defaults this yields the classic doubling schedule 1s, 2s, 4s,
// 8s, and so on; a Factor of 10 with a MinTimeout of 10ms and MaxTimeout of
// 100ms yields 10ms, 100ms, 100ms, 100ms as the exponential growth saturates at
// the cap. The Retries option counts retries after the first attempt, so
// Retries of 3 means up to four total invocations, and Retries of 0 means fn is
// attempted exactly once.
//
// One intentional deviation from node-retry is that this port does not apply
// randomized jitter. node-retry exposes a "randomize" option that multiplies
// each delay by a random factor between 1 and 2 to spread out retries from many
// clients (the thundering-herd problem); here the backoff is fully
// deterministic, which keeps the schedule predictable and testable. Callers who
// need jitter can wrap Do or add it inside a custom Sleep. The Backoff method is
// exported so the exact schedule can be inspected or reused without running the
// retry loop.
//
// Timing is injectable: Options.Sleep replaces the default time.Sleep, and
// Options.OnRetry is invoked before each retry with the failing error and the
// attempt number that just failed. Together these make the backoff schedule
// observable and testable without waiting on real time, and let callers surface
// retry activity to logs or metrics. Do is safe to call concurrently as long as
// the supplied fn and callbacks are themselves safe; the package holds no shared
// state between calls.
package retry

import (
	"errors"
	"math"
	"time"
)

// AbortError wraps an error to signal that retrying must stop immediately. The
// wrapped error is returned to the caller of Do.
type AbortError struct {
	Err error
}

// Error implements the error interface.
func (a *AbortError) Error() string {
	if a.Err == nil {
		return "aborted"
	}
	return a.Err.Error()
}

// Unwrap returns the wrapped error so errors.Is / errors.As work through it.
func (a *AbortError) Unwrap() error { return a.Err }

// Abort creates an AbortError that stops retrying immediately when returned
// from the retried function.
func Abort(err error) *AbortError {
	return &AbortError{Err: err}
}

// Options configures Do.
type Options struct {
	// Retries is the maximum number of retries after the first attempt. A
	// value of 0 means the function is attempted exactly once. This mirrors
	// the "retries" option of node-retry / p-retry.
	Retries int
	// Factor is the exponential backoff factor. Defaults to 2 when zero.
	Factor float64
	// MinTimeout is the delay before the first retry. Defaults to 1s when
	// zero.
	MinTimeout time.Duration
	// MaxTimeout caps the delay between retries. Defaults to no cap
	// (math.MaxInt64) when zero.
	MaxTimeout time.Duration
	// OnRetry, if set, is called before each retry with the error that caused
	// the retry and the attempt number that just failed (1-based).
	OnRetry func(err error, attempt int)
	// Sleep, if set, is used instead of time.Sleep. It makes the backoff
	// injectable for testing.
	Sleep func(d time.Duration)
}

func (o Options) factor() float64 {
	if o.Factor <= 0 {
		return 2
	}
	return o.Factor
}

func (o Options) minTimeout() time.Duration {
	if o.MinTimeout <= 0 {
		return time.Second
	}
	return o.MinTimeout
}

func (o Options) maxTimeout() time.Duration {
	if o.MaxTimeout <= 0 {
		return time.Duration(math.MaxInt64)
	}
	return o.MaxTimeout
}

func (o Options) sleep(d time.Duration) {
	if o.Sleep != nil {
		o.Sleep(d)
		return
	}
	time.Sleep(d)
}

// Backoff computes the delay before the given retry attempt. attempt is
// 1-based, where attempt 1 is the delay before the first retry.
//
// The formula is min(MaxTimeout, MinTimeout * Factor^(attempt-1)).
func (o Options) Backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	min := float64(o.minTimeout())
	delay := min * math.Pow(o.factor(), float64(attempt-1))
	max := float64(o.maxTimeout())
	if delay > max {
		delay = max
	}
	if delay < 0 || math.IsInf(delay, 1) {
		return o.maxTimeout()
	}
	return time.Duration(delay)
}

// Do calls fn, retrying with exponential backoff on error until it succeeds,
// the retry budget is exhausted, or an AbortError is returned.
//
// fn receives the 1-based attempt number. The error returned by the final
// failed attempt is returned to the caller (unwrapped from AbortError).
func Do(fn func(attempt int) error, opts Options) error {
	if opts.Retries < 0 {
		opts.Retries = 0
	}
	var lastErr error
	for attempt := 1; ; attempt++ {
		err := fn(attempt)
		if err == nil {
			return nil
		}

		var abortErr *AbortError
		if errors.As(err, &abortErr) {
			if abortErr.Err != nil {
				return abortErr.Err
			}
			return abortErr
		}

		lastErr = err
		// attempt is 1-based; we have used `attempt` tries. We may retry while
		// the number of retries used (attempt) is <= Retries.
		if attempt > opts.Retries {
			return lastErr
		}

		if opts.OnRetry != nil {
			opts.OnRetry(err, attempt)
		}
		opts.sleep(opts.Backoff(attempt))
	}
}
