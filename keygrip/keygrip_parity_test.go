package keygrip_test

// Upstream parity tests for the npm library "crypto-utils/keygrip".
//
// Test vectors are taken verbatim from the original project's test suite and
// package metadata:
//
//	https://raw.githubusercontent.com/crypto-utils/keygrip/master/test/keygrip.js
//	https://raw.githubusercontent.com/crypto-utils/keygrip/master/index.js
//	https://raw.githubusercontent.com/crypto-utils/keygrip/master/package.json
//
// The upstream library defaults to HMAC-SHA1 with base64 encoding, but lets the
// caller select an algorithm/encoding. This Go port intentionally fixes the
// scheme to HMAC-SHA256 with unpadded base64url output. The vectors below are
// therefore drawn from upstream's explicit `'sha256'` test cases, which use the
// same data, keys, and base64url digest form as this port. Upstream's default
// SHA1 vectors ("34Sr3OIsheUYWKL5_w--zJsdSNk", "_jl9qXYgk5AgBiKFOPYK073FMEQ")
// and the "hex" encoding vector ("df84abdce22c85e51858a2f9ff0fbecc9b1d48d9")
// are NOT ported because the port does not expose configurable algorithm or
// encoding; those are recorded as intentional gaps, not defects.

import (
	"testing"

	"github.com/malcolmston/express/keygrip"
)

// upstream sha256 vector:
//
//	new Keygrip(['SEKRIT1'], 'sha256').sign('Keyboard Cat has a hat.')
//	  === 'pu97aPRZRLKi3-eANtIlTG_CwSc39mAcIZ1c6FxsGCk'
const (
	parityData          = "Keyboard Cat has a hat."
	paritySEKRIT1Digest = "pu97aPRZRLKi3-eANtIlTG_CwSc39mAcIZ1c6FxsGCk"
	// Derived by applying the same HMAC-SHA256 base64url scheme to upstream's
	// exact key "SEKRIT2" and data; used for the rotation vectors below.
	paritySEKRIT2Digest = "lglsBnB7isG_dUwlD56WbA_PXLZZdA_MpHsPXhznfoc"
	// upstream "should fail when key not present" bogus digest.
	parityBogusDigest = "xmM8HQl2eBtPP9nmZ7BK_wpqoxQ"
)

// From upstream ".sign(data) with algorithm sha256".
func TestParitySignSHA256(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT1"})
	if got := kg.Sign(parityData); got != paritySEKRIT1Digest {
		t.Fatalf("Sign = %q, want %q", got, paritySEKRIT1Digest)
	}
}

// From upstream ".index(data) with algorithm sha256" -> 0.
func TestParityIndexSHA256(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT1"})
	if got := kg.Index(parityData, paritySEKRIT1Digest); got != 0 {
		t.Fatalf("Index = %d, want 0", got)
	}
}

// From upstream ".verify(data) with algorithm sha256" -> true.
func TestParityVerifySHA256(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT1"})
	if !kg.Verify(parityData, paritySEKRIT1Digest) {
		t.Fatal("Verify = false, want true")
	}
}

// From upstream '"keys" argument when empty array / when undefined' -> throws
// "Keys must be provided". The port panics for an empty/nil key list.
func TestParityConstructorEmptyThrows(t *testing.T) {
	for _, keys := range [][]string{nil, {}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("New(%v) did not panic; want panic", keys)
				}
			}()
			keygrip.New(keys)
		}()
	}
}

// From upstream ".index(data) should return -1 when no key matches".
func TestParityIndexNoMatch(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT2", "SEKRIT1"})
	if got := kg.Index(parityData, parityBogusDigest); got != -1 {
		t.Fatalf("Index = %d, want -1", got)
	}
}

// From upstream ".verify should fail with bogus data" and
// "should fail when key not present".
func TestParityVerifyBogus(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT2", "SEKRIT1"})
	if kg.Verify(parityData, "bogus data") {
		t.Fatal("Verify(bogus) = true, want false")
	}
	if kg.Verify(parityData, parityBogusDigest) {
		t.Fatal("Verify(absent key) = true, want false")
	}
}

// Mirrors upstream ".sign should sign with first secret": a two-key keygrip
// signs with the first (front) key. Upstream asserts this with SHA1; here the
// SHA256 digest of the front key SEKRIT2 is expected.
func TestParitySignFirstKey(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT2", "SEKRIT1"})
	if got := kg.Sign(parityData); got != paritySEKRIT2Digest {
		t.Fatalf("Sign = %q, want %q", got, paritySEKRIT2Digest)
	}
}

// Mirrors upstream ".index should return key index that signed data" for a
// rotated key list: the front key resolves to index 0, the trailing (rotated
// out) key to index 1.
func TestParityIndexRotation(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT2", "SEKRIT1"})
	if got := kg.Index(parityData, paritySEKRIT2Digest); got != 0 {
		t.Fatalf("Index(front key) = %d, want 0", got)
	}
	if got := kg.Index(parityData, paritySEKRIT1Digest); got != 1 {
		t.Fatalf("Index(rotated key) = %d, want 1", got)
	}
}

// Mirrors upstream ".verify should validate against any key" for a rotated list.
func TestParityVerifyAnyKey(t *testing.T) {
	kg := keygrip.New([]string{"SEKRIT2", "SEKRIT1"})
	if !kg.Verify(parityData, paritySEKRIT2Digest) {
		t.Fatal("Verify(front key) = false, want true")
	}
	if !kg.Verify(parityData, paritySEKRIT1Digest) {
		t.Fatal("Verify(rotated key) = false, want true")
	}
}
