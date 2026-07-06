// Package sanitize provides middleware that strips HTML tags from every
// query-string value on the incoming request, mitigating trivial reflected
// cross-site-scripting (XSS) vectors that flow through query parameters. It is
// the express framework's Go analogue of the input-cleaning middlewares Node
// developers reach for — such as express-sanitizer, sanitize-html wired into a
// middleware, or the deprecated express-validator sanitizers — reduced to a
// single, dependency-free tag-stripping pass over the request's query.
//
// Reach for this middleware when query parameters are reflected back into HTML
// responses, log lines, or templates and you want a cheap, always-on backstop
// against markup injection. It is best understood as defence in depth rather
// than a primary control: it removes the angle-bracketed tags that carry most
// naive <script> and event-handler payloads before your handlers ever see the
// value, so a parameter echoed straight into a page cannot smuggle a live
// element. It should sit alongside, not instead of, proper contextual output
// encoding at render time.
//
// Operationally the middleware belongs early in the chain, ahead of any handler
// that reads the query, so downstream code observes only sanitized values. On
// each request it parses req.Raw.URL.Query(), runs StripTags over every value
// of every key, and — only if something actually changed — re-encodes the
// result back into req.Raw.URL.RawQuery. It then calls next() exactly once and
// never writes to the response or short-circuits; a request is always allowed
// to proceed. Because the rewrite happens on the underlying *http.Request URL,
// both express accessors such as req.Query and any raw net/http inspection of
// the URL see the cleaned data.
//
// The stripping rule is the exported StripTags helper, backed by the regular
// expression <[^>]*>: it deletes any run beginning with "<" and ending at the
// next ">", tags and all. This is intentionally blunt — it treats the text
// between tags as-is and does not decode entities, so "<b>hi</b>" becomes "hi"
// while a lone "a < b" comparison or an already-encoded "&lt;script&gt;" is
// left untouched. Only the query string is processed; request bodies, headers,
// route parameters, and cookies are not, and the middleware takes no Options.
//
// Compared with the full-featured Node sanitizers it stands in for, this port
// is deliberately minimal and must not be mistaken for an HTML sanitizer. It
// applies no allow-list of safe tags or attributes, does not neutralise
// dangerous attributes, javascript: URLs, or malformed/unclosed tags, and can
// be defeated by inputs that never form a complete "<...>" span. Treat it as a
// convenience filter for query parameters and always pair it with real output
// encoding — for example html/template — wherever untrusted data is rendered.
package sanitize

import (
	"regexp"

	"github.com/malcolmston/express"
)

// tagPattern matches an HTML/XML tag, e.g. <script> or </div>.
var tagPattern = regexp.MustCompile(`<[^>]*>`)

// StripTags removes anything that looks like an HTML tag from s.
func StripTags(s string) string {
	return tagPattern.ReplaceAllString(s, "")
}

// New returns middleware that removes HTML tags from every query-string value
// in place, rewriting the request's RawQuery so downstream handlers see the
// sanitized values.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		q := req.Raw.URL.Query()
		changed := false
		for _, values := range q {
			for i, v := range values {
				if nv := StripTags(v); nv != v {
					values[i] = nv
					changed = true
				}
			}
		}
		if changed {
			req.Raw.URL.RawQuery = q.Encode()
		}
		next()
	}
}
