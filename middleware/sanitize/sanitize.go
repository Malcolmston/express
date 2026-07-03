// Package sanitize provides middleware that strips HTML tags from all
// query-string values on the incoming request, mitigating trivial reflected
// XSS vectors that flow through query parameters.
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
