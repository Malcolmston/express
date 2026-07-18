package lang

import "math"

// This file extends the lang package with the numeric predicates and coercions
// from lodash's "Lang" category that operate on concrete float64/int values:
// IsInteger, IsFinite, IsNaN and IsSafeInteger test a float, while
// ToSafeInteger, ToLength and ToInt round or clamp a value into a usable
// integer range. They complement the reflection-based predicates already in the
// package and are deterministic, depending only on the math package.

// maxSafeInteger is 2^53-1, the largest exactly-representable integer in a
// float64, matching JavaScript's Number.MAX_SAFE_INTEGER.
const maxSafeInteger = 9007199254740991

// maxArrayLength mirrors lodash's clamp bound for ToLength (2^32-1).
const maxArrayLength = 4294967295

// IsFinite reports whether f is a finite number (neither NaN nor an infinity),
// mirroring lodash.isFinite for concrete float64 values.
func IsFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// IsNaN reports whether f is an IEEE-754 not-a-number value, mirroring
// lodash.isNaN for concrete float64 values.
func IsNaN(f float64) bool {
	return math.IsNaN(f)
}

// IsInteger reports whether f is finite and has no fractional part, mirroring
// lodash.isInteger.
func IsInteger(f float64) bool {
	return IsFinite(f) && math.Trunc(f) == f
}

// IsSafeInteger reports whether f is an integer within the safe range
// [-(2^53-1), 2^53-1], mirroring lodash.isSafeInteger.
func IsSafeInteger(f float64) bool {
	return IsInteger(f) && math.Abs(f) <= maxSafeInteger
}

// ToSafeInteger converts f to an integer clamped into the safe-integer range,
// mirroring lodash.toSafeInteger. NaN becomes 0.
func ToSafeInteger(f float64) float64 {
	if math.IsNaN(f) {
		return 0
	}
	t := math.Trunc(f)
	if t > maxSafeInteger {
		return maxSafeInteger
	}
	if t < -maxSafeInteger {
		return -maxSafeInteger
	}
	return t
}

// ToLength clamps f to a valid array length in [0, 2^32-1] after truncating any
// fractional part, mirroring lodash.toLength.
func ToLength(f float64) int {
	if math.IsNaN(f) || f <= 0 {
		return 0
	}
	t := math.Trunc(f)
	if t > maxArrayLength {
		return maxArrayLength
	}
	return int(t)
}
