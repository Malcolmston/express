// Package requestcounter provides middleware that counts the total number of
// requests handled. The count is exposed through an accessor function returned
// alongside the middleware and is safe for concurrent use.
package requestcounter

import (
	"sync/atomic"

	"github.com/malcolmston/express"
)

// New returns request-counting middleware together with an accessor that
// reports the number of requests observed so far.
func New() (express.Handler, func() int64) {
	var count int64
	handler := func(req *express.Request, res *express.Response, next express.Next) {
		atomic.AddInt64(&count, 1)
		next()
	}
	accessor := func() int64 { return atomic.LoadInt64(&count) }
	return handler, accessor
}
