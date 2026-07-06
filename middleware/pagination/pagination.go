// Package pagination provides middleware that parses the common "page" and
// "limit" query-string parameters into a normalized, bounds-checked Pagination
// value stored on the request. It ports the ubiquitous limit/offset paging
// convention used by Node helpers such as express-paginate and the manual
// req.query.page / req.query.limit parsing found in most Express APIs, turning
// that repetitive boilerplate into a single reusable handler.
//
// Use it on any list endpoint that returns results in pages. Instead of each
// route re-reading and re-validating the query parameters, the middleware does
// the parsing once and exposes a ready-to-use Pagination struct carrying the
// requested Page, the effective Limit, and a precomputed Offset suitable for
// passing straight to a database LIMIT/OFFSET clause or slice bounds.
//
// Mechanically the middleware runs before your route handlers, typically via
// app.Use. For each request it reads req.Query("page") and req.Query("limit"),
// converts them to integers (falling back to defaults on missing or malformed
// values), clamps them to sensible bounds, computes Offset as (Page-1)*Limit,
// and stores the resulting Pagination on the request with req.Set under a
// private context key before calling next. It never writes to the response or
// short-circuits the chain; handlers retrieve the value with the From helper.
//
// The clamping rules define the important semantics. Page is coerced to a
// minimum of 1, so negative, zero, or non-numeric page values become the first
// page. Limit falls back to Options.DefaultLimit (20 when unset) whenever the
// parameter is absent, non-numeric, or less than 1, and is capped at
// Options.MaxLimit (100 when unset) to protect the backend from unbounded
// result requests. Because a zero-value Options is filled in by applyDefaults,
// calling New() with no arguments yields the 20/100 defaults; From returns a
// zero Pagination if the middleware never ran, so guard accordingly if a route
// might be reached without it.
//
// Parity with the Node originals is conceptual rather than API-exact. Like
// express-paginate this package centralizes page/limit parsing and enforces a
// maximum limit, but it exposes the result as a typed struct fetched via From
// rather than by mutating req.query or attaching req.skip/req.offset fields, and
// it uses 1-based pages with a derived Offset instead of driver-specific
// helpers. The defaults and clamping behavior are chosen to be safe and
// predictable rather than to match any single Node library byte for byte.
package pagination

import (
	"strconv"

	"github.com/malcolmston/express"
)

const contextKey = "pagination"

// Pagination holds the parsed and normalized paging parameters for a request.
type Pagination struct {
	// Page is the 1-based page number.
	Page int
	// Limit is the number of items per page.
	Limit int
	// Offset is the zero-based item offset, equal to (Page-1)*Limit.
	Offset int
}

// Options configures the pagination middleware.
type Options struct {
	// DefaultLimit is used when no valid "limit" query parameter is provided
	// (default 20).
	DefaultLimit int
	// MaxLimit caps the "limit" query parameter (default 100).
	MaxLimit int
}

func (o *Options) applyDefaults() {
	if o.DefaultLimit <= 0 {
		o.DefaultLimit = 20
	}
	if o.MaxLimit <= 0 {
		o.MaxLimit = 100
	}
}

// New returns middleware that reads ?page and ?limit, clamps them to sensible
// bounds, and stores the resulting Pagination on the request. Retrieve it with
// From.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		page := parseInt(req.Query("page"), 1)
		if page < 1 {
			page = 1
		}
		limit := parseInt(req.Query("limit"), o.DefaultLimit)
		if limit < 1 {
			limit = o.DefaultLimit
		}
		if limit > o.MaxLimit {
			limit = o.MaxLimit
		}
		req.Set(contextKey, Pagination{
			Page:   page,
			Limit:  limit,
			Offset: (page - 1) * limit,
		})
		next()
	}
}

// From returns the Pagination stored by the middleware, or a zero value if the
// middleware did not run.
func From(req *express.Request) Pagination {
	if v, ok := req.Value(contextKey); ok {
		if p, ok := v.(Pagination); ok {
			return p
		}
	}
	return Pagination{}
}

func parseInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
