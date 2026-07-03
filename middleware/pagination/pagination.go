// Package pagination provides middleware that parses common "page" and "limit"
// query-string parameters into a normalized, bounds-checked Pagination value
// stored on the request.
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
