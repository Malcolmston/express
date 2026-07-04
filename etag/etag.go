// Package etag creates HTTP entity tags, a port of the npm "etag" package.
//
// A strong tag is formatted as "<hex-len>-<base64-sha1>" where the base64 of
// the SHA-1 digest is truncated to 27 characters. A stat-based tag is formatted
// as "<hex-size>-<hex-mtime-ms>". Weak tags are prefixed with "W/".
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
