// Package cuid generates collision-resistant unique identifiers, a Go port of
// the npm "cuid" package (in the style of its cuid2 successor). Ids are
// URL-safe, hard to guess, and safe to generate across many hosts and
// processes without any central coordination or shared state.
//
// Reach for a cuid when you need a unique identifier for a database row, an
// event, an uploaded object, or a distributed message, and you want something
// friendlier than a raw UUID: the output is a short lowercase alphanumeric
// string that always begins with a letter, so it is safe to use in URLs, as
// an HTML element id, or as a filename without escaping. Unlike a
// sequential integer key, a cuid reveals little about how many ids have been
// issued and does not require a round trip to a central allocator, which
// makes it convenient for client-side generation and for sharded systems.
//
// Each id is derived by hashing several independent sources of entropy
// together with SHA-512 and encoding the digest in base-36. The inputs are:
// the current wall-clock time in nanoseconds (base-36 encoded), a
// process-global atomic counter that is seeded from crypto/rand at startup
// and incremented on every call, a block of fresh random characters drawn
// per call, and a per-process fingerprint built from the PID, hostname, and
// additional random salt. Mixing a monotonic counter and a timestamp with
// per-call and per-process randomness is what gives the scheme its
// collision resistance: even two calls in the same nanosecond on the same
// host receive distinct counter values and distinct random blocks. The first
// character of the id is an independently chosen random letter (never a
// digit), and the first character of the hash body is dropped because it can
// be biased toward low values.
//
// A few semantics and edge cases are worth noting. New produces ids of
// DefaultLength (24) characters. NewLength lets you choose a length, but the
// value is clamped to the range 2..32 rather than rejected: a request for a
// length below 2 is treated as 2 and one above 32 as 32, so NewLength never
// returns an out-of-range id. When the hash body is shorter than the
// requested length it is extended with additional random characters and then
// truncated to exactly n. IsCuid performs a syntactic check only: it accepts
// strings of length 2..32 that start with a lowercase letter and otherwise
// contain only lowercase letters and digits; it does not and cannot verify
// that a string was actually produced by this package. Although the functions
// return an error for interface symmetry, generation does not fail in
// practice because crypto/rand failures fall back to deterministic defaults.
//
// Parity with the JavaScript cuid family is at the level of guarantees rather
// than byte-for-byte output. Ids from this package share the important
// properties of the original: they are horizontally scalable, monotonically
// influenced by time, start with a letter, use a URL-safe lowercase
// alphanumeric alphabet, and resist collisions and guessing. Because the
// fingerprint incorporates this process's PID, hostname, and random salt, and
// because the counter is seeded randomly, the exact strings are not
// reproducible across runs or portable from a Node process; the intent is a
// compatible identifier scheme, not identical values.
package cuid

import (
	"crypto/rand"
	"crypto/sha512"
	"math/big"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	// DefaultLength is the default cuid2 id length.
	DefaultLength = 24
	bigLength     = 32
	letters       = "abcdefghijklmnopqrstuvwxyz"
)

var counter uint64

var fingerprint string

// slugFingerprint is a stable two-character base-36 condensation of the
// process fingerprint, used to give slugs a per-process flavour the same way
// the upstream cuid.slug() mixes in fingerprint().slice(0,1)+slice(-1).
var slugFingerprint string

func init() {
	// Seed the counter with a random starting value.
	var seed [8]byte
	if _, err := rand.Read(seed[:]); err == nil {
		var n uint64
		for _, b := range seed {
			n = n<<8 | uint64(b)
		}
		atomic.StoreUint64(&counter, n)
	}
	fingerprint = buildFingerprint()
	sum := sha512.Sum512([]byte(fingerprint))
	sf := base36Encode(sum[:])
	for len(sf) < 2 {
		sf += "0"
	}
	slugFingerprint = sf[:2]
}

func buildFingerprint() string {
	pid := strconv.Itoa(os.Getpid())
	host, _ := os.Hostname()
	var salt [16]byte
	rand.Read(salt[:])
	return pid + host + randomString(16) + string(salt[:])
}

// randomLetter returns a random lowercase starting letter.
func randomLetter() byte {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
	if err != nil {
		return 'k'
	}
	return letters[n.Int64()]
}

// randomString returns n random base36 characters from crypto/rand.
func randomString(n int) string {
	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, n)
	buf := make([]byte, n)
	rand.Read(buf)
	for i := 0; i < n; i++ {
		out[i] = base36[int(buf[i])%len(base36)]
	}
	return string(out)
}

// base36 encodes a byte slice as a lowercase base-36 string.
func base36Encode(b []byte) string {
	n := new(big.Int).SetBytes(b)
	if n.Sign() == 0 {
		return "0"
	}
	return n.Text(36)
}

// New returns a cuid2 with the default length.
func New() (string, error) {
	return NewLength(DefaultLength)
}

// NewLength returns a cuid2 of the given length (must be between 2 and 32).
func NewLength(n int) (string, error) {
	if n < 2 {
		n = 2
	}
	if n > bigLength {
		n = bigLength
	}

	firstLetter := randomLetter()

	t := strconv.FormatInt(time.Now().UnixNano(), 36)
	c := strconv.FormatUint(atomic.AddUint64(&counter, 1), 36)
	entropy := randomString(bigLength)

	h := sha512.New()
	h.Write([]byte(t))
	h.Write([]byte(c))
	h.Write([]byte(entropy))
	h.Write([]byte(fingerprint))
	hashed := base36Encode(h.Sum(nil))

	// Drop the first hash char (can be biased low) and take the rest.
	if len(hashed) > 1 {
		hashed = hashed[1:]
	}

	// Compose: leading letter + hash body, trimmed to length n.
	body := string(firstLetter) + hashed
	for len(body) < n {
		body += randomString(bigLength)
	}
	return body[:n], nil
}

// Slug returns a short, slug-style identifier, a port of the upstream
// cuid.slug() helper. Like the original it concatenates the tail of a base-36
// timestamp, a base-36 counter (at most four characters), a two-character
// per-process fingerprint, and two random characters. The result is always
// URL-safe and between 7 and 10 characters long, so IsSlug reports it valid.
func Slug() (string, error) {
	date := strconv.FormatInt(time.Now().UnixMilli(), 36)
	if len(date) > 2 {
		date = date[len(date)-2:]
	}

	c := strconv.FormatUint(atomic.AddUint64(&counter, 1), 36)
	if len(c) > 4 {
		c = c[len(c)-4:]
	}

	r := randomString(4)
	random := r[len(r)-2:]

	return date + c + slugFingerprint + random, nil
}

// IsSlug reports whether s has the length of a cuid slug. Mirroring the
// upstream cuid.isSlug(), it is a length check only: any string of length 7
// through 10 (inclusive) is accepted; everything else is rejected.
func IsSlug(s string) bool {
	n := len(s)
	return n >= 7 && n <= 10
}

// IsCuid reports whether s is a syntactically valid cuid2.
func IsCuid(s string) bool {
	if len(s) < 2 || len(s) > bigLength {
		return false
	}
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			return false
		}
	}
	return true
}
