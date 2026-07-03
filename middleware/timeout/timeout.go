// Package timeout provides middleware that bounds the time a downstream
// handler chain may take. The next handler runs in a separate goroutine under
// a context deadline; if it does not complete before the deadline, the client
// receives a 503 Service Unavailable response.
//
// A sync.Once elects a single "winner" (either the handler or the timeout) that
// is permitted to write the response, so the two goroutines can never both
// commit headers. The handler goroutine writes through a wrapper with a private
// header map, ensuring it never touches the underlying writer once it has lost
// the race.
package timeout

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the timeout middleware.
type Options struct {
	// Duration is the maximum time the downstream chain may run. Values <= 0
	// default to 30 seconds.
	Duration time.Duration
	// Message is the body sent on timeout. Defaults to "Service Unavailable".
	Message string
}

// respGuard elects a single owner of the response.
type respGuard struct {
	once   sync.Once
	winner int // 1 = handler, 2 = timeout
}

func (g *respGuard) claim(id int) bool {
	g.once.Do(func() { g.winner = id })
	return g.winner == id
}

// handlerWriter is the writer the downstream chain writes through. It keeps a
// private header map so that a losing handler never races the timeout writer on
// the underlying ResponseWriter.
type handlerWriter struct {
	orig http.ResponseWriter
	hdr  http.Header
	g    *respGuard
}

func (w *handlerWriter) Header() http.Header { return w.hdr }

func (w *handlerWriter) WriteHeader(code int) {
	if w.g.claim(1) {
		copyHeader(w.orig.Header(), w.hdr)
		w.orig.WriteHeader(code)
	}
}

func (w *handlerWriter) Write(b []byte) (int, error) {
	if w.g.claim(1) {
		return w.orig.Write(b)
	}
	return len(b), nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		dst[k] = append([]string(nil), vv...)
	}
}

// New returns timeout middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Duration <= 0 {
		o.Duration = 30 * time.Second
	}
	if o.Message == "" {
		o.Message = "Service Unavailable"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		ctx, cancel := context.WithTimeout(req.Raw.Context(), o.Duration)
		defer cancel()
		req.Raw = req.Raw.WithContext(ctx)

		g := &respGuard{}
		orig := res.Writer
		res.Writer = &handlerWriter{orig: orig, hdr: make(http.Header), g: g}

		done := make(chan struct{})
		go func() {
			defer close(done)
			next()
		}()

		select {
		case <-done:
			return
		case <-ctx.Done():
			if g.claim(2) {
				h := orig.Header()
				if h.Get("Content-Type") == "" {
					h.Set("Content-Type", "text/plain; charset=utf-8")
				}
				orig.WriteHeader(http.StatusServiceUnavailable)
				_, _ = orig.Write([]byte(o.Message))
			}
			return
		}
	}
}
