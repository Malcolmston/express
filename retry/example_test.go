package retry_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/malcolmston/express/retry"
)

// ExampleDo retries a failing operation until it succeeds. The supplied function
// receives the 1-based attempt number and fails on the first two attempts,
// succeeding on the third. To keep the example fast and deterministic the Sleep
// hook is replaced with a no-op so no real time passes between attempts. Do
// returns nil once the operation succeeds, and here the function was invoked
// exactly three times. Retries of 5 means up to six total invocations were
// permitted.
func ExampleDo() {
	count := 0
	err := retry.Do(func(attempt int) error {
		count++
		if attempt < 3 {
			return errors.New("transient failure")
		}
		return nil
	}, retry.Options{Retries: 5, Sleep: func(time.Duration) {}})
	fmt.Println(count, err)
	// Output: 3 <nil>
}

// ExampleOptions_Backoff inspects the backoff schedule without running the retry
// loop. With a Factor of 2 and a MinTimeout of one second the delays double on
// each attempt: 1s, 2s, then 4s. The MaxTimeout of 5 seconds caps the fourth
// delay, which would otherwise be 8 seconds, at 5 seconds. The attempt argument
// is 1-based, where attempt 1 is the pause before the first retry. This lets
// callers verify or reuse the exact schedule the retry loop will follow.
func ExampleOptions_Backoff() {
	o := retry.Options{Factor: 2, MinTimeout: time.Second, MaxTimeout: 5 * time.Second}
	fmt.Println(o.Backoff(1), o.Backoff(2), o.Backoff(3), o.Backoff(4))
	// Output: 1s 2s 4s 5s
}

// ExampleAbort stops retrying immediately by returning an AbortError, even
// though the retry budget has not been exhausted. The wrapped error is returned
// to the caller unchanged and no further attempts are made, which is the right
// behavior for a permanent failure such as an authentication error where
// retrying is pointless. Here the function is called only once. Abort mirrors
// p-retry's AbortError, which short-circuits the retry loop.
func ExampleAbort() {
	count := 0
	err := retry.Do(func(attempt int) error {
		count++
		return retry.Abort(errors.New("permanent failure"))
	}, retry.Options{Retries: 5, Sleep: func(time.Duration) {}})
	fmt.Println(count, err)
	// Output: 1 permanent failure
}
