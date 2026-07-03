// Package responsetime provides express middleware that measures how long a
// request takes to process and reports it in an X-Response-Time header.
package responsetime

import (
	"strconv"
	"time"

	"github.com/malcolmston/express"
)

// HeaderName is the response header carrying the measured duration.
const HeaderName = "X-Response-Time"

// now is overridable in tests.
var now = time.Now

// New returns middleware that records the time spent handling a request and
// sets X-Response-Time (in milliseconds, e.g. "12.34ms"). The header is
// written via OnBeforeWrite so the timing reflects work up to the moment the
// response headers are committed.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		start := now()
		res.OnBeforeWrite(func() {
			elapsed := now().Sub(start)
			ms := float64(elapsed) / float64(time.Millisecond)
			res.Set(HeaderName, strconv.FormatFloat(ms, 'f', 2, 64)+"ms")
		})
		next()
	}
}
