package etag

import (
	"testing"
	"time"
)

// Upstream parity vectors extracted verbatim from the original jshttp/etag
// package tests and source:
//
//	https://raw.githubusercontent.com/jshttp/etag/master/test/test.js
//	https://raw.githubusercontent.com/jshttp/etag/master/index.js
//
// The Node etag(entity, {weak}) function maps onto this port as:
//   - string / Buffer content -> Generate(data, weak)
//   - fs.Stats                -> GenerateStat(size, mtime, weak)
//
// Note: upstream also asserts on a 5KB string/buffer whose bytes come from a
// JS seedrandom('etag test') PRNG. That PRNG is not reproducible with the Go
// standard library, so those two vectors are intentionally omitted here.

// TestParityStringVectors covers the string content vectors from test.js.
func TestParityStringVectors(t *testing.T) {
	cases := []struct {
		name string
		in   string
		weak bool
		want string
	}{
		// test.js:22
		{"beep boop", "beep boop", false, `"9-fINXV39R1PCo05OqGqr7KIY9lCE"`},
		// test.js:26 (multibyte UTF-8, 3 bytes)
		{"multibyte strong", "论", false, `"3-QkSKq8sXBjHL2tFAZknA2n6LYzM"`},
		// test.js:27
		{"multibyte weak", "论", true, `W/"3-QkSKq8sXBjHL2tFAZknA2n6LYzM"`},
		// test.js:31 / 70
		{"empty", "", false, `"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`},
		// test.js:88
		{"empty weak", "", true, `W/"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`},
		// test.js:71 / 89
		{"beep boop weak", "beep boop", true, `W/"9-fINXV39R1PCo05OqGqr7KIY9lCE"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate([]byte(tc.in), tc.weak)
			if got != tc.want {
				t.Fatalf("Generate(%q, %v) = %s, want %s", tc.in, tc.weak, got, tc.want)
			}
		})
	}
}

// TestParityBufferVectors covers the Buffer content vectors from test.js.
func TestParityBufferVectors(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		weak bool
		want string
	}{
		// test.js:37 / 77
		{"bytes 1,2,3 strong", []byte{1, 2, 3}, false, `"3-cDeAcZjCKn0rCAc3HXY3eahP388"`},
		// test.js:95
		{"bytes 1,2,3 weak", []byte{1, 2, 3}, true, `W/"3-cDeAcZjCKn0rCAc3HXY3eahP388"`},
		// test.js:41 / 76
		{"empty buffer strong", []byte{}, false, `"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`},
		// test.js:94
		{"empty buffer weak", []byte{}, true, `W/"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate(tc.in, tc.weak)
			if got != tc.want {
				t.Fatalf("Generate(%v, %v) = %s, want %s", tc.in, tc.weak, got, tc.want)
			}
		})
	}
}

// TestParityStatVector covers the fs.Stats vector from test.js:55-63.
//
// Upstream fakeStat: { mtime: new Date('2014-09-01T14:52:07Z'), ino: 0,
// size: 3027 } and stats default to weak, producing 'W/"bd3-14831b399d8"'.
// stattag = '"' + size.toString(16) + '-' + mtime.getTime().toString(16) + '"'.
func TestParityStatVector(t *testing.T) {
	mtime, err := time.Parse(time.RFC3339, "2014-09-01T14:52:07Z")
	if err != nil {
		t.Fatal(err)
	}

	// Strong form of the same tag body.
	if got, want := GenerateStat(3027, mtime, false), `"bd3-14831b399d8"`; got != want {
		t.Fatalf("GenerateStat strong = %s, want %s", got, want)
	}
	// Weak form, matching upstream's stats-default-weak behaviour.
	if got, want := GenerateStat(3027, mtime, true), `W/"bd3-14831b399d8"`; got != want {
		t.Fatalf("GenerateStat weak = %s, want %s", got, want)
	}
}
