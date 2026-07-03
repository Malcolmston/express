// Package responsecache provides simple in-memory caching of successful GET
// responses. Cached entries are keyed by request URI and replayed to clients
// until their time-to-live expires. Only 200 OK GET responses are cached.
package responsecache

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the response cache.
type Options struct {
	// TTL is how long a cached response remains valid (default 60s).
	TTL time.Duration
}

type entry struct {
	status      int
	body        []byte
	contentType string
	expires     time.Time
}

// New returns middleware that caches successful GET responses in memory for the
// configured TTL. On a cache hit the stored status, body, and Content-Type are
// replayed and the request short-circuits; on a miss the response is captured
// and stored. It is safe for concurrent use.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.TTL <= 0 {
		o.TTL = 60 * time.Second
	}

	var (
		mu    sync.Mutex
		cache = make(map[string]entry)
	)

	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Method() != http.MethodGet {
			next()
			return
		}
		key := req.Raw.URL.RequestURI()

		mu.Lock()
		e, ok := cache[key]
		if ok && time.Now().After(e.expires) {
			delete(cache, key)
			ok = false
		}
		mu.Unlock()

		if ok {
			if e.contentType != "" {
				res.Set("Content-Type", e.contentType)
			}
			res.Set("X-Cache", "HIT")
			res.Status(e.status).Send(e.body)
			return
		}

		rec := &recorder{ResponseWriter: res.Writer, status: http.StatusOK}
		res.Writer = rec
		res.Set("X-Cache", "MISS")
		next()

		if rec.status == http.StatusOK {
			mu.Lock()
			cache[key] = entry{
				status:      rec.status,
				body:        append([]byte(nil), rec.buf.Bytes()...),
				contentType: rec.Header().Get("Content-Type"),
				expires:     time.Now().Add(o.TTL),
			}
			mu.Unlock()
		}
	}
}

// recorder is an http.ResponseWriter that tees writes to the underlying writer
// while buffering the body and recording the status code.
type recorder struct {
	http.ResponseWriter
	status      int
	buf         bytes.Buffer
	wroteHeader bool
}

func (r *recorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	r.buf.Write(p)
	return r.ResponseWriter.Write(p)
}
