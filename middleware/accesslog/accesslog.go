// Package accesslog provides middleware that writes an Apache "combined"
// style access log line for every request once it completes.
package accesslog

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the access log middleware.
type Options struct {
	// Writer is where log lines are written (default os.Stdout).
	Writer io.Writer
}

// New returns middleware that writes an Apache combined-log-format line after
// each request, observing the final status code and number of bytes written.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Writer == nil {
		o.Writer = os.Stdout
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		start := time.Now()
		rec := &recorder{ResponseWriter: res.Writer, status: http.StatusOK}
		res.Writer = rec

		next()

		host := clientHost(req.Raw.RemoteAddr)
		line := fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
			host,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			req.Method(),
			req.Raw.URL.RequestURI(),
			req.Raw.Proto,
			rec.status,
			rec.bytes,
			valueOrDash(req.Get("Referer")),
			valueOrDash(req.Get("User-Agent")),
		)
		_, _ = io.WriteString(o.Writer, line)
	}
}

func clientHost(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	if remoteAddr == "" {
		return "-"
	}
	return remoteAddr
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// recorder observes the status code and number of bytes written.
type recorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *recorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.bytes += n
	return n, err
}
