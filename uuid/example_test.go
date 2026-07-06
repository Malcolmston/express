package uuid_test

import (
	"fmt"

	"github.com/malcolmston/express/uuid"
)

// ExampleV4 generates a random version 4 UUID. The value is random, so the
// example asserts the two invariants that always hold rather than printing the
// identifier: the canonical string is 36 characters long, and it passes
// Validate. V4 reads 16 bytes from crypto/rand and stamps in the version nibble
// and RFC 4122 variant bits before formatting. With 122 random bits, collisions
// are astronomically unlikely and require no coordination between generators.
// This is the identifier to reach for when you need an opaque, unguessable key.
func ExampleV4() {
	id, err := uuid.V4()
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id), uuid.Validate(id))
	// Output: 36 true
}

// ExampleV5 generates a deterministic, name-based UUID. It hashes the namespace
// UUID together with the name using SHA-1 and stamps in the version and variant
// bits, so the same namespace and name always yield the same identifier. Here
// the DNS namespace combined with "example.com" produces a fixed, well-known
// value that is byte-for-byte identical to what Node's uuid.v5 returns. This
// makes V5 ideal for stable ids derived from URLs or domain names. Unlike V4 it
// is reproducible rather than random.
func ExampleV5() {
	id, err := uuid.V5(uuid.NamespaceDNS, "example.com")
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	// Output: cfbff0d1-9375-5685-968c-48ce8b15ae17
}

// ExampleParse converts a canonical UUID string into its raw 16 bytes and back
// again with Format, confirming the round trip. Parse is strict about the
// hyphenated 36-character layout and rejects wrong lengths, misplaced dashes, or
// non-hex digits. Format performs the inverse, rendering the bytes as the
// canonical lowercase dashed string. Here the value survives the round trip
// unchanged. Validate is a thin wrapper that reports whether Parse would succeed.
func ExampleParse() {
	b, err := uuid.Parse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	if err != nil {
		panic(err)
	}
	fmt.Println(uuid.Format(b))
	// Output: 6ba7b810-9dad-11d1-80b4-00c04fd430c8
}
