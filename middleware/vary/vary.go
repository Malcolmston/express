// Package vary provides express middleware that appends fields to the response
// Vary header, signalling which request headers a cached response depends on.
// It is a port of the widely-used "vary" npm package (the same helper that
// backs Express's res.vary() and the CORS middleware's header bookkeeping):
// given one or more field names, it ensures each appears exactly once in the
// response's Vary header, appending without disturbing entries already set by
// earlier middleware.
//
// The Vary header tells caches and CDNs which request headers were used to
// select the representation being returned, so a response negotiated on
// Accept-Encoding or Origin is not incorrectly served to a client that sent
// different values. Use this middleware whenever a response varies by request
// header — content negotiation, gzip/br compression, per-Origin CORS
// responses, or language selection — and you want the correct Vary field
// declared without hand-managing the header string. Declaring Vary is a
// correctness concern for shared caches: omitting it can serve the wrong body,
// while over-declaring "*" defeats caching entirely.
//
// Mount it with app.Use, typically before the handlers that produce the varied
// content so the field is present regardless of which branch runs. For each
// non-empty field in Options.Fields the middleware reads the current Vary
// header with res.GetHeader("Vary") and, only if that field is not already
// present, appends it via res.Append("Vary", f). It then always calls next();
// it never short-circuits and never writes a body. Because it appends rather
// than overwrites, it composes cleanly with other middleware (such as CORS)
// that also touch Vary.
//
// De-duplication is case-insensitive and comma-token aware: the internal
// hasField helper splits the existing header on commas, trims surrounding
// spaces and tabs from each token, and compares with an ASCII case-fold, so a
// pre-existing "accept-encoding" suppresses a configured "Accept-Encoding".
// Empty field strings in Options.Fields are skipped, and calling New with no
// Options (or an Options whose Fields is nil) yields a harmless pass-through
// that adds nothing. Note that each appended field becomes its own header line
// via res.Append rather than being folded into a single comma-joined value.
//
// Compared with the Node "vary" package this port covers the common append
// case but is narrower: it does not special-case or collapse a wildcard "*"
// (the npm version replaces the whole header with "*" when that field is
// added), it does not parse or re-emit a single combined header value, and it
// exposes only the middleware constructor rather than a standalone append
// function. If you need "*" semantics or a one-shot append helper, set the
// header directly with res.Vary or res.Append.
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
