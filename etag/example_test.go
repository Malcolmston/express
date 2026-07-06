package etag_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/etag"
)

// ExampleGenerate shows the common case of deriving a strong entity tag from a
// response body held in memory. The body's bytes are hashed with SHA-1 and the
// digest is base64-encoded and truncated, while the byte length is emitted as
// lowercase hexadecimal. The returned string already includes the surrounding
// double quotes, so it can be written straight into an ETag header. Passing
// false selects a strong validator, which asserts byte-for-byte equality. The
// hyphen joins the hex length ("b" for eleven bytes) and the truncated hash.
func ExampleGenerate() {
	tag := etag.Generate([]byte("Hello World"), false)
	fmt.Println(tag)
	// Output: "b-Ck1VqNd45QIvq3AZd8XYQLvEhtA"
}

// ExampleGenerate_weak demonstrates the weak variant together with empty input.
// Empty content hashes to the well-known SHA-1 of the empty string, giving the
// stable tag body "0-2jmj7l5rSw0yVb/vlWAYkK/YBwk". Because weak is true, the
// tag is prefixed with the "W/" marker that signals a semantic rather than a
// byte-for-byte match. Weak tags are useful when two representations should be
// treated as equivalent even if their bytes differ slightly. The double quotes
// remain part of the value, sitting inside the "W/" prefix.
func ExampleGenerate_weak() {
	tag := etag.Generate([]byte(""), true)
	fmt.Println(tag)
	// Output: W/"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"
}

// ExampleGenerateStat builds a tag from a resource's size and modification time
// instead of hashing its bytes, which is what static file servers do to avoid
// reading large payloads. The size and the modification time in milliseconds
// since the Unix epoch are both formatted as lowercase hexadecimal and joined
// with a hyphen. Here 1024 bytes becomes "400" and a modification time of one
// second past the epoch (1000 ms) becomes "3e8". Passing false again selects a
// strong validator with no "W/" prefix.
func ExampleGenerateStat() {
	modtime := time.Unix(1, 0).UTC()
	tag := etag.GenerateStat(1024, modtime, false)
	fmt.Println(tag)
	// Output: "400-3e8"
}
