// Package servertiming provides express middleware that collects timing metrics
// during request handling and emits them in a Server-Timing response header,
// which browsers surface in their developer tools' network panel. It lets a Go
// service report how long individual server-side phases took — a database query,
// a template render, an upstream call — directly to the browser inspecting the
// response, in the same standardized format that Node middleware such as
// server-timing produces.
//
// Reach for it when you want per-request server observability that requires no
// external tooling: because the numbers ride along on the response the client
// already receives, a developer can open dev tools and see the breakdown next to
// the request, and synthetic monitors or tests can parse the header to assert on
// backend latency. Each metric is a named segment with a duration in
// milliseconds and an optional human-readable description, so a single header
// can carry several labelled measurements at once.
//
// Mechanically New installs a fresh Metrics collector on the request under a
// private context key and registers an OnBeforeWrite hook that, just before the
// response is committed, renders the accumulated entries and appends them as the
// Server-Timing header when at least one metric was recorded. Handlers reach the
// collector with From, which resolves the per-request Metrics value and records
// measurements through Add or AddWithDesc. Recording is guarded by a mutex, so a
// handler may add entries from multiple goroutines that fan out during a single
// request without a data race.
//
// The rendered header follows the Server-Timing grammar: each entry is emitted
// as name;dur=<milliseconds> with two decimal places, and when a description is
// present a ;desc="..." segment is inserted between them. Any double quotes in a
// description are replaced with single quotes so the value cannot break out of
// its quoted form. Durations are supplied as time.Duration values and divided
// down to fractional milliseconds, so sub-millisecond timings still appear with
// precision rather than rounding to zero. When no metrics were recorded the hook
// writes no header at all, keeping responses clean for routes that opt out.
//
// From is deliberately null-safe: if it is called on a request that never passed
// through this middleware it returns a detached, throwaway Metrics rather than
// nil, so handler code can time work unconditionally without guarding against a
// missing collector — the measurements simply go nowhere. Relative to the Node
// original this port keeps the core collect-then-emit model and the standard
// header format while exposing a compact Go surface: a HeaderName constant, the
// Metrics type with Add and AddWithDesc, and the From accessor, in place of
// JavaScript's chainable request-decorating API.
package servertiming

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/malcolmston/express"
)

// HeaderName is the response header used to report metrics.
const HeaderName = "Server-Timing"

// contextKey is the key under which the Metrics collector is stored on req.
const contextKey = "serverTiming"

// entry is a single Server-Timing metric.
type entry struct {
	name string
	dur  time.Duration
	desc string
}

// Metrics collects Server-Timing entries for a single request. It is safe for
// concurrent use.
type Metrics struct {
	mu      sync.Mutex
	entries []entry
}

// Add records a metric with the given name and duration.
func (m *Metrics) Add(name string, dur time.Duration) {
	m.AddWithDesc(name, "", dur)
}

// AddWithDesc records a metric with a human-readable description.
func (m *Metrics) AddWithDesc(name, desc string, dur time.Duration) {
	m.mu.Lock()
	m.entries = append(m.entries, entry{name: name, dur: dur, desc: desc})
	m.mu.Unlock()
}

// header renders the collected metrics as a Server-Timing header value, or ""
// if there are none.
func (m *Metrics) header() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m.entries))
	for _, e := range m.entries {
		var b strings.Builder
		b.WriteString(e.name)
		if e.desc != "" {
			b.WriteString(`;desc="`)
			b.WriteString(strings.ReplaceAll(e.desc, `"`, `'`))
			b.WriteString(`"`)
		}
		ms := float64(e.dur) / float64(time.Millisecond)
		b.WriteString(";dur=")
		b.WriteString(strconv.FormatFloat(ms, 'f', 2, 64))
		parts = append(parts, b.String())
	}
	return strings.Join(parts, ", ")
}

// From returns the Metrics collector associated with the request, allowing
// handlers to add timing entries. It never returns nil; if no collector is
// installed (middleware not mounted) a detached one is returned so callers
// need not nil-check.
func From(req *express.Request) *Metrics {
	if v, ok := req.Value(contextKey); ok {
		if m, ok := v.(*Metrics); ok {
			return m
		}
	}
	return &Metrics{}
}

// New returns middleware that installs a Metrics collector on the request and
// emits the accumulated Server-Timing header just before the response is
// written.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		m := &Metrics{}
		req.Set(contextKey, m)
		res.OnBeforeWrite(func() {
			if h := m.header(); h != "" {
				res.Append(HeaderName, h)
			}
		})
		next()
	}
}
