// Package timeout provides middleware that bounds the time a downstream
// handler chain may take before the client is answered. It is the express
// framework's Go analogue of Connect's connect-timeout middleware and the
// classic express-timeout wrappers: a drop-in express.Handler that runs the
// rest of the chain under a deadline and, if that deadline elapses first,
// answers the client with 503 Service Unavailable instead of leaving the
// request hanging.
//
// Reach for this middleware when a slow or stuck handler must not be allowed
// to tie up a client indefinitely: routes that call flaky upstream services,
// database queries without their own timeout, report generation, or any place
// where a bounded worst-case latency is more valuable than always waiting for
// the "real" answer. Placing it near the front of the chain -- before the
// handlers you want to bound -- makes every handler downstream of it subject
// to the same Duration budget.
//
// Operationally, on each request the middleware derives a context.WithTimeout
// from req.Raw.Context() using Options.Duration and installs it back onto the
// request, so a cooperative downstream handler that watches req.Raw.Context()
// can observe cancellation. It then swaps res.Writer for a handlerWriter that
// buffers headers in a private http.Header map and runs next() in a separate
// goroutine, while the middleware itself selects on the handler's completion
// versus the context deadline. If the handler finishes first, its buffered
// headers and body are flushed to the real writer and the request completes
// normally; if the deadline fires first, the middleware writes the 503.
//
// Because two goroutines could otherwise both try to commit the response, a
// respGuard built on sync.Once elects exactly one winner: the handler claims
// id 1 on its first Write or WriteHeader, the timeout path claims id 2, and
// only the winner is allowed to touch the underlying writer. A handler that
// loses the race still runs to completion, but its writes are silently
// discarded (Write reports success without copying bytes) so a late response
// can never corrupt or duplicate the 503 already sent. On timeout the
// middleware also defaults the Content-Type to text/plain; charset=utf-8 when
// the handler had not already set one.
//
// Two option fields tune the behavior: Duration (the deadline, defaulting to
// 30 seconds when <= 0) and Message (the 503 body, defaulting to "Service
// Unavailable"). Note the important semantic difference from a synchronous
// port: this middleware does not abort the handler goroutine -- Go cannot
// forcibly kill a goroutine -- so a handler ignoring context cancellation
// keeps running in the background until it returns, merely with its output
// suppressed. Callers must therefore ensure long-running work honors context
// cancellation to actually reclaim resources; the timeout guarantees a timely
// response to the client, not immediate termination of the work behind it.
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

// Header implements http.ResponseWriter; it returns the handler's private
// header map rather than the underlying ResponseWriter's headers.
func (w *handlerWriter) Header() http.Header { return w.hdr }

// WriteHeader implements http.ResponseWriter; it claims the response guard and,
// if the handler wins the race with the timeout, copies the buffered headers to
// the underlying ResponseWriter and writes the status code. Otherwise it is a
// no-op.
func (w *handlerWriter) WriteHeader(code int) {
	if w.g.claim(1) {
		copyHeader(w.orig.Header(), w.hdr)
		w.orig.WriteHeader(code)
	}
}

// Write implements http.ResponseWriter; it writes b to the underlying
// ResponseWriter when the handler wins the response guard, and otherwise
// discards b (reporting it as fully written) so a timed-out handler cannot race
// the timeout writer.
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
