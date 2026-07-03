// Package slowlog provides middleware that logs a warning whenever a request
// takes longer than a configured threshold to complete, helping surface slow
// endpoints.
package slowlog

import (
	"log"
	"os"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the slow-request logger.
type Options struct {
	// Threshold is the duration above which a request is considered slow
	// (default 1s).
	Threshold time.Duration
	// Logger receives the warning line. When nil a logger writing to os.Stderr
	// is used.
	Logger *log.Logger
}

// New returns middleware that measures the time spent handling each request and
// logs a warning when it exceeds the threshold.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Threshold <= 0 {
		o.Threshold = time.Second
	}
	logger := o.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		start := time.Now()
		next()
		elapsed := time.Since(start)
		if elapsed > o.Threshold {
			logger.Printf("slowlog: WARNING %s %s took %s (threshold %s)",
				req.Method(), req.Path(), elapsed.Round(time.Microsecond), o.Threshold)
		}
	}
}
