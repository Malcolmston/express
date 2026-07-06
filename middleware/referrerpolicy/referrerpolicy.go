// Package referrerpolicy provides middleware that sets the Referrer-Policy
// response header, controlling how much referrer information the browser
// attaches to outgoing requests originated from your pages. It is the express
// port of Helmet's referrerPolicy middleware (helmet.referrerPolicy()), which
// in turn wraps the single standardized header that browsers honor to govern
// the Referer request header on navigations, resource loads, and prefetches.
//
// Use it whenever you want to limit referrer leakage: the default Referer
// header can disclose the full URL of the page a user came from, including
// path and query string, to third-party origins and to downgraded (HTTPS to
// HTTP) destinations. Setting an explicit policy such as "no-referrer" or
// "strict-origin-when-cross-origin" is a common privacy and security baseline,
// which is why it appears in Helmet and most hardened header presets. It is
// cheap enough to enable globally with app.Use.
//
// Mechanically the middleware is trivial and body-neutral: it computes a single
// header value once at construction time and, on every request, calls
// res.Set("Referrer-Policy", value) and then immediately invokes next() to
// continue the chain. It reads no request state, writes exactly one header,
// and never short-circuits or terminates the response. Because it only mutates
// response headers, its position in the chain matters only insofar as it must
// run before the headers are flushed; registering it early is the usual choice
// so the header is present on both normal and error responses.
//
// The semantics are driven by Options.Policy. When no options are supplied, or
// Policy is empty, the middleware emits "no-referrer" — the most restrictive
// policy, which strips the Referer header entirely. When Policy contains one or
// more tokens they are joined with ", " and used verbatim, allowing a fallback
// list (for example {"no-referrer", "strict-origin-when-cross-origin"}) where
// browsers use the last token they understand. The package does not validate
// tokens against the specification, so any string is emitted as given; an
// invalid token results in browsers ignoring the header and falling back to
// their built-in default policy. The last write wins if a later handler
// overwrites the header.
//
// Parity with the Node original is exact for the behavior that matters: like
// helmet.referrerPolicy() this package sets a single Referrer-Policy header from
// a configurable policy value (a string or list) and defaults to "no-referrer".
// It does not reproduce Helmet's broader bundle of headers (see the helmet
// middleware for that), and because Go's net/http canonicalizes header casing
// there is no configuration surface on which to diverge.
package referrerpolicy

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the referrerpolicy middleware. The zero value is usable
// and yields Referrer-Policy: no-referrer.
type Options struct {
	// Policy is one or more Referrer-Policy tokens. When empty, "no-referrer"
	// is used. Multiple tokens are joined with ", ".
	Policy []string
}

// New returns middleware that sets the Referrer-Policy header.
func New(opts ...Options) express.Handler {
	value := "no-referrer"
	if len(opts) > 0 && len(opts[0].Policy) > 0 {
		value = strings.Join(opts[0].Policy, ", ")
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Referrer-Policy", value)
		next()
	}
}
