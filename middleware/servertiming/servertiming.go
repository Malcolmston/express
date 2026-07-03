// Package servertiming provides express middleware that collects timing
// metrics during request handling and emits them in a Server-Timing response
// header, which browsers surface in their developer tools.
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
