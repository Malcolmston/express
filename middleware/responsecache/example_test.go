package responsecache_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/responsecache"
)

// ExampleNew mounts the response cache in front of an expensive GET handler and
// issues the same request twice through httptest. A counter inside the handler
// records how many times it actually executes, and the X-Cache response header
// reports whether each request was served fresh or from cache. The first
// request misses the cache (X-Cache: MISS), runs the handler, and its 200 body
// is stored for the configured TTL; the second request hits the cache
// (X-Cache: HIT) and replays the stored body without invoking the handler
// again. The example prints the header, status, and body for both calls plus
// the final handler-invocation count to show the short-circuit in action.
func ExampleNew() {
	app := express.New()
	app.Use(responsecache.New(responsecache.Options{TTL: time.Minute}))

	calls := 0
	app.Get("/report", func(req *express.Request, res *express.Response, next express.Next) {
		calls++
		res.Type("text").Send("expensive result")
	})

	do := func() {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/report", nil))
		fmt.Printf("%s %d %s\n", w.Header().Get("X-Cache"), w.Code, w.Body.String())
	}

	do() // miss: handler runs
	do() // hit: served from cache

	fmt.Printf("handler ran %d time(s)\n", calls)

	// Output:
	// MISS 200 expensive result
	// HIT 200 expensive result
	// handler ran 1 time(s)
}
