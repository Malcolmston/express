// Package querynormalize provides middleware that normalizes the request query
// string: it lower-cases parameter keys, trims surrounding whitespace from
// values, and rebuilds the raw query in a deterministic (key-sorted) order.
//
// It plays the role commonly handled in Node by a query canonicalization step
// layered on top of a parser such as qs — the sort of normalization done before
// caching, signing, or comparing URLs (comparable in spirit to the
// "normalize-url" package's query-key sorting and trimming). By collapsing
// case, whitespace, and ordering differences, semantically equivalent requests
// such as "?B=2&A=1" and "?a=1&b=2" produce an identical canonical query
// string, which improves cache-key stability and makes downstream query
// handling predictable.
//
// Use it when handlers or caches rely on a stable, case-insensitive view of the
// query string, or when you want to defend against clients that vary key case,
// pad values with spaces, or reorder parameters. Register it early with app.Use
// so the rewrite happens before any middleware or handler reads req.Query,
// req.QueryValues, or req.Raw.URL.RawQuery; everything downstream then observes
// the normalized form.
//
// On each request the middleware reads req.Raw.URL.RawQuery and, when it is
// non-empty, parses it with url.ParseQuery. Each key is lower-cased with
// strings.ToLower and each value is trimmed with strings.TrimSpace, preserving
// every value (including duplicates) under its normalized key. The result is
// re-encoded with url.Values.Encode, which sorts keys alphabetically, and the
// encoded string is written back to req.Raw.URL.RawQuery in place. The
// middleware writes no response headers and never short-circuits; it always
// calls next() to continue the chain.
//
// Important semantics and edge cases: requests with no query string are left
// untouched and pass straight through. A query string that fails url.ParseQuery
// is intentionally left as-is rather than dropped, so malformed input is not
// silently discarded. Only keys are lower-cased — values keep their original
// case and are merely whitespace-trimmed — and re-encoding may change
// percent-encoding to Go's canonical form. Because keys are sorted and case is
// folded, output is fully deterministic regardless of input ordering. This
// middleware takes no options.
//
// Compared with the Node originals it draws from, parity is partial and
// deliberate: it performs key lower-casing, value trimming, and key sorting, but
// does not attempt the broader normalization (default-value handling, array
// bracket syntax, or unicode canonicalization) that a full qs plus
// normalize-url pipeline might apply.
package querynormalize

import (
	"net/url"
	"strings"

	"github.com/malcolmston/express"
)

// New returns middleware that rewrites req.Raw.URL.RawQuery in normalized form.
// Requests without a query string are left untouched.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		raw := req.Raw.URL.RawQuery
		if raw == "" {
			next()
			return
		}
		values, err := url.ParseQuery(raw)
		if err != nil {
			// Leave a malformed query string as-is rather than dropping data.
			next()
			return
		}

		normalized := make(url.Values, len(values))
		for key, vals := range values {
			lk := strings.ToLower(key)
			for _, v := range vals {
				normalized[lk] = append(normalized[lk], strings.TrimSpace(v))
			}
		}

		// url.Values.Encode sorts keys, producing deterministic output.
		req.Raw.URL.RawQuery = normalized.Encode()
		next()
	}
}
