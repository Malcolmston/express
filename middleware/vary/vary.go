// Package vary provides express middleware that appends fields to the response
// Vary header, signalling which request headers a cached response depends on.
package vary

import "github.com/malcolmston/express"

// Options configures the vary middleware.
type Options struct {
	// Fields is the list of header names to append to the Vary header.
	Fields []string
}

// New returns middleware that appends each configured field to the response's
// Vary header. Duplicate entries already present on the header are skipped so
// the header stays tidy.
func New(opts ...Options) express.Handler {
	var fields []string
	if len(opts) > 0 {
		fields = opts[0].Fields
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		for _, f := range fields {
			if f == "" {
				continue
			}
			if !hasField(res.GetHeader("Vary"), f) {
				res.Append("Vary", f)
			}
		}
		next()
	}
}

// hasField reports whether the comma-separated Vary header already contains
// field (case-insensitive).
func hasField(header, field string) bool {
	if header == "" {
		return false
	}
	start := 0
	for i := 0; i <= len(header); i++ {
		if i == len(header) || header[i] == ',' {
			token := trimSpace(header[start:i])
			if equalFold(token, field) {
				return true
			}
			start = i + 1
		}
	}
	return false
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
