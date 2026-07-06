// Package etag creates HTTP entity tags, a port of the npm "etag" package.
//
// An entity tag (ETag) is an opaque validator that a server sends in the ETag
// response header so that clients can make conditional requests with
// If-None-Match or If-Match. When the representation of a resource changes its
// ETag changes, allowing caches and browsers to revalidate cheaply and servers
// to answer unchanged requests with 304 Not Modified instead of resending the
// body.
//
// Use Generate when you have the full response body in memory and want a value
// derived from the content itself. Use GenerateStat when you would rather avoid
// hashing a large payload and can instead identify a file by its size and last
// modification time, which is what express and send do for static files. Both
// return a ready-to-use header value, including the surrounding double quotes.
//
// The content algorithm mirrors the Node original: it takes the SHA-1 digest of
// the bytes, base64-encodes it with standard encoding, and truncates the result
// to 27 characters (dropping the trailing base64 padding). The byte length is
// formatted as lowercase hexadecimal and the two parts are joined with a hyphen
// inside double quotes, producing "<hex-len>-<base64-sha1>". GenerateStat uses a
// cheaper scheme, formatting the size and the modification time in milliseconds
// since the Unix epoch as lowercase hex and joining them as
// "<hex-size>-<hex-mtime-ms>".
//
// A weak tag differs from a strong tag only by a leading "W/" marker. Strong
// tags assert byte-for-byte equality, while weak tags assert that two
// representations are semantically equivalent even if their bytes differ; when
// weak is true both functions simply prepend "W/" to the quoted tag. Empty
// content is fully supported: Generate of an empty slice yields the well-known
// SHA-1 of the empty string, "0-2jmj7l5rSw0yVb/vlWAYkK/YBwk", and GenerateStat
// of a zero size and the Unix epoch yields "0-0". Neither function can fail, so
// there is no error return.
//
// Compared to the Node package, this port keeps the exact tag format but exposes
// two explicit entry points instead of a single overloaded function. The Node
// etag(entity, options) accepts a string, Buffer, or fs.Stats and infers weak
// versus strong from the options and the entity type (stat-based tags default to
// weak there); here the caller chooses Generate versus GenerateStat and passes
// the weak flag directly. Streaming inputs and the automatic weak default for
// stats are intentionally omitted in favour of this smaller, explicit surface.
package etag

import (
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"
)

// Generate computes an entity tag for the given content. When weak is true the
// tag is prefixed with "W/".
func Generate(data []byte, weak bool) string {
	sum := sha1.Sum(data)
	hash := base64.StdEncoding.EncodeToString(sum[:])
	if len(hash) > 27 {
		hash = hash[:27]
	}
	tag := "\"" + strconv.FormatInt(int64(len(data)), 16) + "-" + hash + "\""
	if weak {
		return "W/" + tag
	}
	return tag
}

// GenerateStat computes a stat-based entity tag from a resource's size and
// modification time. The modification time is expressed in milliseconds since
// the Unix epoch. When weak is true the tag is prefixed with "W/".
func GenerateStat(size int64, modtime time.Time, weak bool) string {
	mtimeMs := modtime.UnixNano() / int64(time.Millisecond)
	tag := "\"" + strconv.FormatInt(size, 16) + "-" + strconv.FormatInt(mtimeMs, 16) + "\""
	if weak {
		return "W/" + tag
	}
	return tag
}
