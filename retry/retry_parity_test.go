package retry

import (
	"testing"
	"time"
)

// Upstream parity tests for the npm "retry" package (tim-kos/node-retry).
//
// Vectors are taken verbatim from the upstream test suite and the backoff
// formula in the original library:
//
//   test/integration/test-timeouts.js:
//     https://raw.githubusercontent.com/tim-kos/node-retry/master/test/integration/test-timeouts.js
//   lib/retry.js (createTimeout / timeouts):
//     https://raw.githubusercontent.com/tim-kos/node-retry/master/lib/retry.js
//
// Upstream createTimeout(attempt, opts) computes, with randomize=false:
//     round(max(minTimeout, 1) * factor^attempt)   clamped to maxTimeout
// where `attempt` is the 0-based index into the timeouts() array.
//
// The Go port's Options.Backoff(n) is 1-based:
//     min(maxTimeout, minTimeout * factor^(n-1))
// so upstream createTimeout(k) == Go Backoff(k+1).

// TestParityDefaultTimeouts mirrors testDefaultValues() plus the full default
// exponential schedule. Upstream defaults: retries=10, factor=2,
// minTimeout=1000ms, maxTimeout=Infinity. timeouts()[k] == 1000 * 2^k.
func TestParityDefaultTimeouts(t *testing.T) {
	o := Options{Factor: 2, MinTimeout: 1000 * time.Millisecond}
	// createTimeout(k) for k = 0..9, i.e. the full 10-element default array.
	want := []time.Duration{
		1000 * time.Millisecond,   // timeouts[0]
		2000 * time.Millisecond,   // timeouts[1]
		4000 * time.Millisecond,   // timeouts[2]
		8000 * time.Millisecond,   // timeouts[3]
		16000 * time.Millisecond,  // timeouts[4]
		32000 * time.Millisecond,  // timeouts[5]
		64000 * time.Millisecond,  // timeouts[6]
		128000 * time.Millisecond, // timeouts[7]
		256000 * time.Millisecond, // timeouts[8]
		512000 * time.Millisecond, // timeouts[9]
	}
	for k, w := range want {
		if got := o.Backoff(k + 1); got != w {
			t.Fatalf("Backoff(%d) = %v, want createTimeout(%d) = %v", k+1, got, k, w)
		}
	}
}

// TestParityFactorLessThanOne mirrors testTimeoutsAreIncrementalForFactorsLessThanOne():
//
//	retry.timeouts({retries: 3, factor: 0.5}) deepEquals [250, 500, 1000].
//
// The upstream array is sorted ascending; the raw per-attempt createTimeout
// values are createTimeout(0)=1000, createTimeout(1)=500, createTimeout(2)=250.
// The Go Backoff (unsorted, per-attempt) must produce those same raw values.
func TestParityFactorLessThanOne(t *testing.T) {
	o := Options{Factor: 0.5, MinTimeout: 1000 * time.Millisecond}
	want := []time.Duration{
		1000 * time.Millisecond, // createTimeout(0) -> Backoff(1)
		500 * time.Millisecond,  // createTimeout(1) -> Backoff(2)
		250 * time.Millisecond,  // createTimeout(2) -> Backoff(3)
	}
	for k, w := range want {
		if got := o.Backoff(k + 1); got != w {
			t.Fatalf("Backoff(%d) = %v, want createTimeout(%d) = %v", k+1, got, k, w)
		}
	}
}

// TestParityMaxTimeoutClamp mirrors the upstream Math.min(timeout, maxTimeout)
// clamp in createTimeout. With factor=3, minTimeout=1000ms, maxTimeout=10000ms
// the raw schedule 1000, 3000, 9000, 27000, 81000 is clamped to
// 1000, 3000, 9000, 10000, 10000.
func TestParityMaxTimeoutClamp(t *testing.T) {
	o := Options{Factor: 3, MinTimeout: 1000 * time.Millisecond, MaxTimeout: 10000 * time.Millisecond}
	want := []time.Duration{
		1000 * time.Millisecond,  // 1000 * 3^0
		3000 * time.Millisecond,  // 1000 * 3^1
		9000 * time.Millisecond,  // 1000 * 3^2
		10000 * time.Millisecond, // 27000 clamped
		10000 * time.Millisecond, // 81000 clamped
	}
	for k, w := range want {
		if got := o.Backoff(k + 1); got != w {
			t.Fatalf("Backoff(%d) = %v, want clamped createTimeout(%d) = %v", k+1, got, k, w)
		}
	}
}
