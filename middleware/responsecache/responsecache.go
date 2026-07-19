// Package responsecache provides simple in-memory caching of successful GET
// responses. It is the express framework's Go analogue of Node caching
// middleware such as apicache, express-cache-response, and the memory store
// behind route-level res.set('Cache-Control') helpers: a drop-in
// express.Handler that memoizes a handler's output and replays it to later
// callers without re-running the handler. The cache lives entirely in process
// memory, is keyed by request URI, and expires entries after a configurable
// time-to-live.
//
// Reach for this middleware to shed load from expensive but idempotent GET
// endpoints — a rendered page, an aggregation query, a third-party API proxy —
// whose responses can tolerate being a few seconds stale. It is best suited to
// read-only, anonymous, or otherwise cache-friendly routes; it is not a CDN, a
// shared cache, or a substitute for HTTP validators, and it does not honor
// client Cache-Control or per-user variation. Because it is a single unbounded
// map, use it for a bounded set of hot URIs rather than as a general store.
//
// Operationally the middleware belongs near the front of the chain, ahead of the
// handler whose output you want to cache. Only GET requests are considered; any
// other method calls next() immediately and is never cached. The cache key is
// req.Raw.URL.RequestURI(), so the path and query string together identify an
// entry (headers, cookies, and Vary are ignored). On a hit the stored status,
// body, and Content-Type are written, an "X-Cache: HIT" header is set, and the
// request short-circuits without calling next(). On a miss it sets "X-Cache:
// MISS", swaps res.Writer for an internal recorder that tees the response to the
// client while buffering it, and calls next() so the real handler runs.
//
// After next() returns, the recorded response is stored only when its status is
// exactly 200 OK; any other status (a 404, a 500, a redirect) is passed through
// to the client but left uncached, so errors are never memoized. Stored entries
// expire lazily: a lookup that finds an entry past its expiry deletes it and
// treats the request as a miss, so nothing is served stale beyond the TTL. The
// map is guarded by a sync.Mutex, so the middleware is safe for concurrent use;
// the buffered body is copied before being stored so the cached bytes are not
// aliased to the recorder. The sole option is Options.TTL, the entry lifetime,
// which defaults to 60 seconds when zero or negative.
//
// Compared with the richer Node originals this port is deliberately small. It
// caches only 200 GET responses, offers no manual invalidation, cache groups,
// max-entry limits, or LRU eviction, does not vary on request headers or emit
// Cache-Control/ETag/Last-Modified, and never spans processes. The one
// observable header it adds is X-Cache (HIT or MISS), which mirrors the debug
// header apicache exposes and is handy for confirming cache behavior in tests.
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

// WriteHeader implements http.ResponseWriter; it records the first status code
// written before delegating to the wrapped ResponseWriter.
func (r *recorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

// Write implements http.ResponseWriter; it buffers a copy of p while also
// writing it through to the wrapped ResponseWriter.
func (r *recorder) Write(p []byte) (int, error) {
	r.buf.Write(p)
	return r.ResponseWriter.Write(p)
}
