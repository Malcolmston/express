package dnsprefetch_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/dnsprefetch"
)

// Example mounts the dnsprefetch middleware with prefetching enabled and shows
// the resulting response header. Passing Options{Allow: true} makes the
// middleware emit X-DNS-Prefetch-Control: on, opting into browser DNS
// prefetching for lower perceived latency; the default New() would emit off. The
// header is written on every response before the route handler runs, and the
// middleware always calls next, so the handler still produces its normal body.
// The example drives the app with httptest and prints the header value, which is
// fully deterministic. Switch Allow to false, or drop the option entirely, to
// prioritize privacy instead.
func Example() {
	app := express.New()
	app.Use(dnsprefetch.New(dnsprefetch.Options{Allow: true}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-DNS-Prefetch-Control"))
	// Output: on
}
