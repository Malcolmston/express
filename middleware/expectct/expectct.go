// Package expectct provides middleware that sets the Expect-CT response header,
// allowing sites to opt in to reporting and/or enforcement of Certificate
// Transparency requirements.
package expectct

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the expectct middleware. The zero value is usable and
// yields a report-only header with max-age=0.
type Options struct {
	// MaxAge is the number of seconds the browser should cache the policy.
	MaxAge int

	// Enforce adds the "enforce" directive.
	Enforce bool

	// ReportURI, when non-empty, adds a report-uri directive.
	ReportURI string
}

// New returns middleware that sets the Expect-CT header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	maxAge := o.MaxAge
	if maxAge < 0 {
		maxAge = 0
	}

	parts := []string{"max-age=" + strconv.Itoa(maxAge)}
	if o.Enforce {
		parts = append(parts, "enforce")
	}
	if o.ReportURI != "" {
		parts = append(parts, `report-uri="`+o.ReportURI+`"`)
	}
	value := strings.Join(parts, ", ")

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Expect-CT", value)
		next()
	}
}
