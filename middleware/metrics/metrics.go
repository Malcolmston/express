// Package metrics provides middleware that records basic request metrics for
// the express framework: the total number of requests and a breakdown by
// status class (2xx, 3xx, 4xx, 5xx). The response status is observed through a
// lightweight ResponseWriter wrapper, and all counters are safe for concurrent
// use.
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

func (w *statusWriter) WriteHeader(code int) {
	if !w.wrote {
		w.status = code
		w.wrote = true
	}
	w.ResponseWriter.WriteHeader(code)
}

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
