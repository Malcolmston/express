// Package cuid generates collision-resistant unique identifiers, a Go port of
// the npm "cuid" package. Ids are monotonic, URL-safe, and safe to generate
// across many hosts without coordination.
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
