package timingsafe

// Upstream parity tests for Node.js crypto.timingSafeEqual(a, b).
//
// Vectors transcribed verbatim from the canonical Node.js reference test:
//
//	nodejs/node  test/sequential/test-crypto-timing-safe-equal.js
//	https://raw.githubusercontent.com/nodejs/node/main/test/sequential/test-crypto-timing-safe-equal.js
//
// The upstream semantics exercised there are:
//   - equal-length, equal-content buffers  -> true
//   - equal-length, differing-content      -> false
//   - timingSafeEqual compares the underlying raw bytes only, ignoring the
//     TypedArray element type (so TypedArray views over identical bytes are
//     equal, and floating-point 0 vs -0 or NaN bit-patterns are compared
//     bytewise, not by numeric/SameValue semantics)
//   - differing byte length -> Node throws ERR_CRYPTO_TIMING_SAFE_EQUAL_LENGTH
//     (RangeError); this Go port returns false instead (documented parity note).
//
// The Go port exposes Equal([]byte, []byte) bool and EqualString(string, string)
// bool, so the "throws on non-buffer" TypeError cases are compile-time in Go and
// have no runtime analogue; they are intentionally omitted.

import (
	"encoding/binary"
	"math"
	"testing"
)

// TestParityEqualStrings mirrors the two headline assertions in the Node test:
//
//	crypto.timingSafeEqual(Buffer.from('foo'), Buffer.from('foo')) === true
//	crypto.timingSafeEqual(Buffer.from('foo'), Buffer.from('bar')) === false
func TestParityEqualStrings(t *testing.T) {
	if got := Equal([]byte("foo"), []byte("foo")); got != true {
		t.Errorf("Equal(foo, foo) = %v, want true", got)
	}
	if got := EqualString("foo", "foo"); got != true {
		t.Errorf("EqualString(foo, foo) = %v, want true", got)
	}
	if got := Equal([]byte("foo"), []byte("bar")); got != false {
		t.Errorf("Equal(foo, bar) = %v, want false", got)
	}
	if got := EqualString("foo", "bar"); got != false {
		t.Errorf("EqualString(foo, bar) = %v, want false", got)
	}
}

// TestParityTypedArrayViewsBytewise mirrors the TypedArray block: a single
// 16-byte buffer viewed as Uint8/Uint16/Uint32 has identical underlying bytes,
// so every cross-comparison is true. In Go all views collapse to the same
// []byte, so the meaningful parity assertion is: identical byte content of a
// non-trivial length compares equal.
func TestParityTypedArrayViewsBytewise(t *testing.T) {
	buf := []byte{
		0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77,
		0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	other := make([]byte, len(buf))
	copy(other, buf)
	if got := Equal(buf, other); got != true {
		t.Errorf("Equal(identical 16 bytes) = %v, want true", got)
	}
}

// TestParityFloatBytewise mirrors the floating-point block, which asserts that
// timingSafeEqual compares raw bytes and has neither == nor Object.is
// semantics. The expected timingSafeEqual results are transcribed directly:
//
//	Float32Array([NaN]) vs Float32Array([NaN])            -> true  (same bytes)
//	Float64Array([0])   vs Float64Array([-0])             -> false (sign bit differs)
//	Float64Array(bits[0x7ff0..1,0xfff0..1]) vs [NaN,NaN]  -> false (bit patterns differ)
//
// Bytes are laid out little-endian to match the x86 TypedArray memory layout
// the upstream test observes.
func TestParityFloatBytewise(t *testing.T) {
	le := binary.LittleEndian

	// Float32Array([NaN]) — JS canonicalizes to the quiet NaN 0x7fc00000.
	nan32 := make([]byte, 4)
	le.PutUint32(nan32, math.Float32bits(float32(math.NaN())))
	nan32b := make([]byte, 4)
	le.PutUint32(nan32b, math.Float32bits(float32(math.NaN())))
	if got := Equal(nan32, nan32b); got != true {
		t.Errorf("Equal(Float32 NaN, Float32 NaN) = %v, want true", got)
	}

	// Float64Array([0]) vs Float64Array([-0]) — bytes differ only in the sign bit.
	zero := make([]byte, 8)
	le.PutUint64(zero, math.Float64bits(0))
	negZero := make([]byte, 8)
	le.PutUint64(negZero, math.Float64bits(math.Copysign(0, -1)))
	if got := Equal(zero, negZero); got != false {
		t.Errorf("Equal(+0.0, -0.0) = %v, want false", got)
	}

	// Signaling-NaN bit patterns vs the canonical quiet NaN 0x7ff8000000000000.
	sig := make([]byte, 16)
	le.PutUint64(sig[0:8], 0x7ff0000000000001)
	le.PutUint64(sig[8:16], 0xfff0000000000001)
	canonNaN := make([]byte, 16)
	le.PutUint64(canonNaN[0:8], math.Float64bits(math.NaN()))
	le.PutUint64(canonNaN[8:16], math.Float64bits(math.NaN()))
	if got := Equal(sig, canonNaN); got != false {
		t.Errorf("Equal(signaling NaN bits, canonical NaN bits) = %v, want false", got)
	}
}

// TestParityLengthMismatch mirrors the final throwing assertion:
//
//	crypto.timingSafeEqual(Buffer.from([1,2,3]), Buffer.from([1,2]))
//	  -> throws ERR_CRYPTO_TIMING_SAFE_EQUAL_LENGTH (RangeError)
//
// Documented parity divergence: this Go port returns false for a length
// mismatch rather than panicking, so the observable "not equal" outcome is
// preserved. See the package doc comment.
func TestParityLengthMismatch(t *testing.T) {
	if got := Equal([]byte{1, 2, 3}, []byte{1, 2}); got != false {
		t.Errorf("Equal([1 2 3], [1 2]) = %v, want false", got)
	}
	if got := Equal([]byte{1, 2}, []byte{1, 2, 3}); got != false {
		t.Errorf("Equal([1 2], [1 2 3]) = %v, want false", got)
	}
}
