package ulid_test

// Upstream-parity tests for the Go port of the npm "ulid" package (v3.0.2).
//
// Every input -> expected-output vector below is copied verbatim from the
// original library's own vitest specs, not invented here:
//
//   https://raw.githubusercontent.com/ulid/javascript/master/test/node/ulid.spec.ts
//   https://raw.githubusercontent.com/ulid/javascript/master/test/node/uuid.spec.ts
//   https://raw.githubusercontent.com/ulid/javascript/master/test/node/crockford.spec.ts
//
// Supporting constants (TIME_MAX, MAX_ULID, MIN_ULID) come from:
//   https://raw.githubusercontent.com/ulid/javascript/master/source/constants.ts
//
// The upstream uuid spec's ulidToUUID / uuidToULID vectors are exact
// ULID-string <-> 16-raw-byte pairings (a ULID's UUID form is just the hex of
// its 16 bytes), so they double as encode/decode parity vectors for this port.

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/malcolmston/express/ulid"
)

// TestParityDecodeTime mirrors ulid.spec.ts "decodeTime":
//
//	expect(decodeTime("01ARYZ6S41TSV4RRFFQ69G5FAV")).to.equal(1469918176385);
func TestParityDecodeTime(t *testing.T) {
	got, err := ulid.Timestamp("01ARYZ6S41TSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("Timestamp error: %v", err)
	}
	if got != 1469918176385 {
		t.Fatalf("Timestamp = %d, want 1469918176385", got)
	}
}

// TestParityEncodeTime mirrors ulid.spec.ts "encodeTime":
//
//	expect(encodeTime(1469918176385)).to.equal("01ARYZ6S41");
//
// The port has no standalone encodeTime, but the first TIME_LEN (10) chars of a
// ULID are exactly encodeTime of its timestamp, so we assert on that prefix.
func TestParityEncodeTime(t *testing.T) {
	id, err := ulid.NewWithEntropy(1469918176385, make([]byte, 10))
	if err != nil {
		t.Fatalf("NewWithEntropy error: %v", err)
	}
	if id[:10] != "01ARYZ6S41" {
		t.Fatalf("time part = %q, want 01ARYZ6S41", id[:10])
	}
}

// TestParityDecodeBytes uses the ulidToUUID vectors from uuid.spec.ts. The UUID
// is the hex of the ULID's 16 raw bytes, so Decode must reproduce those bytes.
//
//	ulidToUUID("01ARYZ6S41TSV4RRFFQ69G5FAV") == "01563DF3-6481-D676-4C61-EFB99302BD5B"
//	ulidToUUID("01JQ4T23H220KM7X0B3V1109DQ") == "0195C9A1-0E22-1027-43F4-0B1EC21025B7"
func TestParityDecodeBytes(t *testing.T) {
	cases := []struct {
		id      string
		hexBits string
	}{
		{"01ARYZ6S41TSV4RRFFQ69G5FAV", "01563DF36481D6764C61EFB99302BD5B"},
		{"01JQ4T23H220KM7X0B3V1109DQ", "0195C9A10E22102743F40B1EC21025B7"},
	}
	for _, c := range cases {
		want, err := hex.DecodeString(c.hexBits)
		if err != nil {
			t.Fatalf("bad hex fixture: %v", err)
		}
		b, err := ulid.Decode(c.id)
		if err != nil {
			t.Fatalf("Decode(%q) error: %v", c.id, err)
		}
		if hex.EncodeToString(b[:]) != hex.EncodeToString(want) {
			t.Fatalf("Decode(%q) = %X, want %s", c.id, b, c.hexBits)
		}
	}
}

// TestParityEncodeBytes uses the uuidToULID vectors from uuid.spec.ts, reversed:
// the 16 raw bytes (hex of the UUID) must encode back to the given ULID string.
//
//	uuidToULID("0195C9A4-2E32-C014-5F4F-A7CEF5BE83D5") == "01JQ4T8BHJR0A5YKX7SVTVX0YN"
//	uuidToULID("0195C9A4-74CC-5149-BCC4-0A556A0CF19D") == "01JQ4T8X6CA54VSH0AANN0SWCX"
func TestParityEncodeBytes(t *testing.T) {
	cases := []struct {
		hexBits string
		want    string
	}{
		{"0195C9A42E32C0145F4FA7CEF5BE83D5", "01JQ4T8BHJR0A5YKX7SVTVX0YN"},
		{"0195C9A474CC5149BCC40A556A0CF19D", "01JQ4T8X6CA54VSH0AANN0SWCX"},
	}
	for _, c := range cases {
		raw, err := hex.DecodeString(c.hexBits)
		if err != nil {
			t.Fatalf("bad hex fixture: %v", err)
		}
		ms := uint64(raw[0])<<40 | uint64(raw[1])<<32 | uint64(raw[2])<<24 |
			uint64(raw[3])<<16 | uint64(raw[4])<<8 | uint64(raw[5])
		id, err := ulid.NewWithEntropy(ms, raw[6:])
		if err != nil {
			t.Fatalf("NewWithEntropy error: %v", err)
		}
		if id != c.want {
			t.Fatalf("encode(%s) = %q, want %q", c.hexBits, id, c.want)
		}
	}
}

// TestParityMaxTimestamp mirrors constants.ts TIME_MAX / MAX_ULID: encoding
// TIME_MAX (2^48-1 = 281474976710655) yields the "7ZZZZZZZZZ" time prefix, and
// MAX_ULID is "7ZZZZZZZZZZZZZZZZZZZZZZZZZ".
func TestParityMaxTimestamp(t *testing.T) {
	const timeMax = uint64(281474976710655)
	maxEntropy := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	id, err := ulid.NewWithEntropy(timeMax, maxEntropy)
	if err != nil {
		t.Fatalf("NewWithEntropy error: %v", err)
	}
	if id[:10] != "7ZZZZZZZZZ" {
		t.Fatalf("time part = %q, want 7ZZZZZZZZZ", id[:10])
	}
	// constants.ts MAX_ULID = "7" followed by 25 "Z" (26 chars total).
	maxULID := "7" + strings.Repeat("Z", 25)
	if id != maxULID {
		t.Fatalf("MAX_ULID = %q, want %q", id, maxULID)
	}
	// Round-trip the max timestamp back out.
	got, err := ulid.Timestamp(id)
	if err != nil {
		t.Fatalf("Timestamp error: %v", err)
	}
	if got != timeMax {
		t.Fatalf("Timestamp = %d, want %d", got, timeMax)
	}
}

// TestParityMinUlid mirrors constants.ts MIN_ULID: a zero timestamp with zero
// entropy is "00000000000000000000000000".
func TestParityMinUlid(t *testing.T) {
	id, err := ulid.NewWithEntropy(0, make([]byte, 10))
	if err != nil {
		t.Fatalf("NewWithEntropy error: %v", err)
	}
	if id != "00000000000000000000000000" {
		t.Fatalf("MIN_ULID = %q, want 00000000000000000000000000", id)
	}
}

// TestParityCrockfordAliases mirrors the i/l -> 1 and o -> 0 substitutions of
// crockford.spec.ts "fixULIDBase32". Upstream also strips hyphens; this port's
// Decode does not, so only the alias-letter behavior is asserted here (hyphen
// stripping is recorded as a known gap in the sync notes).
//
//	fixULIDBase32("oLARYZ6-S41TSV4RRF-FQ69G5FAV") == "01ARYZ6S41TSV4RRFFQ69G5FAV"
func TestParityCrockfordAliases(t *testing.T) {
	canonical, err := ulid.Decode("01ARYZ6S41TSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("Decode canonical error: %v", err)
	}
	// "oLARYZ6S41TSV4RRFFQ69G5FAV": leading o->0, L->1.
	aliased, err := ulid.Decode("oLARYZ6S41TSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("Decode aliased error: %v", err)
	}
	if aliased != canonical {
		t.Fatalf("alias decode = %X, want %X", aliased, canonical)
	}
}
