// Package readiness provides a readiness-probe endpoint for the express
// framework. It exposes a single well-known path (defaulting to "/readyz")
// that answers 200 "ready" when the application reports itself ready to serve
// traffic and 503 "not ready" otherwise; requests to any other path are left
// untouched and continue down the chain. It ports the readiness-gating concept
// that Node process managers such as terminus and lightship attach to an HTTP
// server, where an orchestrator (Kubernetes, a load balancer, a service mesh)
// polls the endpoint and only routes live traffic to instances that report
// themselves ready.
//
// Use it to separate "the process is up" from "the process should receive
// requests." A server may be running yet still be warming a cache, replaying a
// migration, draining connections before shutdown, or waiting on a downstream
// dependency; during those windows the readiness probe should fail so the
// orchestrator holds traffic back without killing the pod (that is the job of a
// separate liveness probe). Because the Ready callback is evaluated on every
// probe, the same endpoint reflects the current state and flips back to 200 the
// moment the application recovers, which the tests exercise directly.
//
// Mechanically the middleware inspects req.Path() first. When the path does not
// match the configured Path it calls next() immediately and reads or writes
// nothing else, so it is safe to register globally via app.Use at the top of
// the chain ahead of routers and other middleware. When the path matches it
// short-circuits the chain: it invokes the Ready function and writes the
// response itself with res.Status(...).Send(...), never calling next(). It does
// not consult any request headers and emits a plain-text body, so it is
// deliberately dependency-free and cheap to call at a high probe frequency.
//
// The behavior is governed by two options with sensible defaults. Path defaults
// to "/readyz" when left empty. Ready defaults to a function that always
// returns true when nil, so a zero-value Options yields an endpoint that is
// always ready — useful as a liveness-style stub. A ready probe returns HTTP
// 200 with the body "ready"; a failing probe returns HTTP 503 Service
// Unavailable with the body "not ready", the conventional signal that tells an
// orchestrator to stop sending traffic without treating the instance as dead.
// There is no allowance for HTTP method: any method hitting the path is probed,
// mirroring the permissive behavior of typical probe endpoints.
//
// Parity with the Node originals is behavioral rather than structural. Libraries
// like terminus and lightship expose richer machinery — multiple named checks,
// graceful-shutdown hooks, signal handling, and JSON health documents — whereas
// this package distills the piece that matters for request gating: a single
// boolean predicate mapped onto 200/503 at a configurable path. Callers that
// need composite checks can aggregate them inside their own Ready function,
// keeping this middleware a thin, allocation-free adapter over the express
// handler signature.
package readiness

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the readiness middleware.
type Options struct {
	// Ready reports whether the application is ready to serve traffic. When nil
	// the application is always considered ready.
	Ready func() bool
	// Path is the endpoint served. Defaults to "/readyz".
	Path string
}

// New returns readiness middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Path == "" {
		o.Path = "/readyz"
	}
	ready := o.Ready
	if ready == nil {
		ready = func() bool { return true }
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Path() != o.Path {
			next()
			return
		}
		if ready() {
			res.Status(http.StatusOK).Send("ready")
			return
		}
		res.Status(http.StatusServiceUnavailable).Send("not ready")
	}
}
