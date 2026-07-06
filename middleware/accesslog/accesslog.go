// Package accesslog provides middleware that writes an Apache "combined" style
// access log line for every request once it completes. It is the Go analogue
// of Node request loggers such as morgan (specifically its "combined" format),
// exposed as a drop-in express.Handler that emits one line per request to a
// configurable writer.
//
// Use this middleware to obtain conventional web-server access logs from an
// express application without pulling in an external logging framework.
// Because it emits the widely understood Apache combined format, its output
// can be consumed directly by log analyzers, shipped to aggregation pipelines,
// or tailed during development. Mount it once at the top of the chain to log
// the entire application, or attach it to a specific router to log only a
// subtree.
//
// Operationally the middleware must run first so it can observe the final
// outcome of the request. On entry it records a start timestamp and wraps
// res.Writer in an internal recorder that transparently proxies WriteHeader
// and Write while capturing the status code and the number of body bytes
// written. It then calls next() to let the rest of the chain run to
// completion, and only afterward formats and writes the log line. The line
// includes the client host, the timestamp, the request line (method, URI, and
// protocol), the captured status and byte count, and the Referer and
// User-Agent request headers.
//
// The single option is Options.Writer, the destination for log lines, which
// defaults to os.Stdout when nil. The recorder defaults the status to 200 so
// handlers that write a body without an explicit WriteHeader are logged
// correctly, and it records only the first WriteHeader call to mirror the real
// response. The client host is derived from RemoteAddr with the port stripped
// via net.SplitHostPort, falling back to the raw address or "-" when empty;
// likewise absent Referer and User-Agent headers are rendered as "-". Writes
// to the log destination are best-effort and their errors are ignored, so
// logging never disrupts the request, and because the line is emitted after
// next() returns the status and byte counts always reflect what was actually
// sent.
//
// Compared with morgan, this port implements only the fixed combined format
// rather than pluggable tokens or named presets, does not compute or log
// response time, and does not support skip predicates or immediate logging.
// It captures status and byte counts by wrapping the writer exactly as morgan
// does, so the recorded values match what the client received.
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
