// Package metrics provides middleware that records basic request metrics for
// the express framework: the total number of requests and a breakdown by
// status class (2xx, 3xx, 4xx, 5xx). It is a minimal, dependency-free analog of
// Node instrumentation middleware such as express-status-monitor, morgan's
// counting variants, or a prom-client HTTP collector, but it stays in the
// standard library and simply accumulates counters in memory rather than
// exporting a Prometheus endpoint.
//
// Use it when you want a cheap, always-on sense of request volume and error
// ratios — for health dashboards, a debug/admin route, or lightweight alerting on
// the 5xx count — without pulling in a full metrics stack. It answers "how many
// requests have we served and how many failed" and nothing more, which is often
// exactly what a small service needs.
//
// Register the handler early with app.Use so it wraps the whole chain. On each
// request it swaps res.Writer for a lightweight statusWriter, calls next() to run
// the rest of the chain, then restores the original writer and records the
// observed status. The statusWriter remembers the first status code written
// (falling back to 200 OK when a handler writes a body without an explicit
// WriteHeader), so the recorded class reflects what the client actually received.
// Because the middleware measures on the way back out, it must sit outside the
// handlers whose responses you want counted.
//
// The counters live on the *Metrics value returned alongside the handler, and
// they are guarded by a mutex so the middleware is safe under concurrent traffic.
// Read them with Snapshot, which returns a copy containing "total" plus one entry
// per status class ("2xx", "3xx", "4xx", "5xx"); status codes below 200 are
// counted in total but fall into no class bucket. Note the two-value constructor:
// New returns (express.Handler, *Metrics) — you must keep the *Metrics to observe
// anything. The counters are process-local and reset when the process restarts;
// this package does no persistence, sampling, or latency timing.
//
// Parity with the Node originals is deliberately partial. It captures the request
// count and status-class breakdown that those tools expose, but it does not track
// response-time histograms, per-route labels, in-flight gauges, or a scrape
// endpoint. If you need Prometheus-style export, layer it on top of Snapshot; this
// package's contract is just the accumulator and the writer-wrapping middleware.
package metrics

import (
	"net/http"
	"sync"

	"github.com/malcolmston/express"
)

// Metrics holds request counters observed by the middleware.
type Metrics struct {
	mu      sync.Mutex
	total   int64
	classes map[string]int64
}

func newMetrics() *Metrics {
	return &Metrics{classes: map[string]int64{"2xx": 0, "3xx": 0, "4xx": 0, "5xx": 0}}
}

// Snapshot returns a copy of the current counters. The returned map contains
// the key "total" plus one entry per status class.
func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]int64, len(m.classes)+1)
	out["total"] = m.total
	for k, v := range m.classes {
		out[k] = v
	}
	return out
}

func (m *Metrics) record(status int) {
	m.mu.Lock()
	m.total++
	switch {
	case status >= 500:
		m.classes["5xx"]++
	case status >= 400:
		m.classes["4xx"]++
	case status >= 300:
		m.classes["3xx"]++
	case status >= 200:
		m.classes["2xx"]++
	}
	m.mu.Unlock()
}

// statusWriter records the first status code written to the response.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

// WriteHeader implements http.ResponseWriter; it records the first status code
// written before delegating to the wrapped ResponseWriter.
func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write implements http.ResponseWriter; it records an implicit 200 status on
// the first write and delegates to the wrapped ResponseWriter.
func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.status = http.StatusOK
		w.wrote = true
	}
	return w.ResponseWriter.Write(b)
}

// New returns metrics middleware together with the *Metrics accumulator it
// writes to.
func New() (express.Handler, *Metrics) {
	m := newMetrics()
	handler := func(req *express.Request, res *express.Response, next express.Next) {
		sw := &statusWriter{ResponseWriter: res.Writer, status: http.StatusOK}
		orig := res.Writer
		res.Writer = sw
		next()
		res.Writer = orig
		m.record(sw.status)
	}
	return handler, m
}
