package express

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CORSOptions configures the [CORS] middleware. The zero value is valid and
// permits any origin with the common HTTP methods.
type CORSOptions struct {
	// AllowOrigins is the list of allowed origins. An entry of "*" (or an empty
	// list) allows any origin. Otherwise the request's Origin is echoed back
	// only when it appears in the list.
	AllowOrigins []string
	// AllowMethods is the list of methods advertised on preflight. Defaults to
	// GET, HEAD, PUT, PATCH, POST and DELETE.
	AllowMethods []string
	// AllowHeaders is advertised on preflight; when empty the requested headers
	// are reflected.
	AllowHeaders []string
	// ExposeHeaders lists response headers browsers may read.
	ExposeHeaders []string
	// AllowCredentials sets Access-Control-Allow-Credentials.
	AllowCredentials bool
	// MaxAge is the preflight cache lifetime in seconds.
	MaxAge int
}

// CORS returns middleware that adds Cross-Origin Resource Sharing headers and
// answers preflight (OPTIONS) requests with 204, mirroring the popular Express
// cors middleware.
func CORS(opts ...CORSOptions) Handler {
	o := CORSOptions{}
	if len(opts) > 0 {
		o = opts[0]
	}
	methods := "GET,HEAD,PUT,PATCH,POST,DELETE"
	if len(o.AllowMethods) > 0 {
		methods = strings.Join(o.AllowMethods, ",")
	}
	return func(req *Request, res *Response, next Next) {
		origin := req.Get("Origin")
		if allow := corsResolveOrigin(o.AllowOrigins, origin); allow != "" {
			res.Set("Access-Control-Allow-Origin", allow)
			if allow != "*" {
				res.Vary("Origin")
			}
		}
		if o.AllowCredentials {
			res.Set("Access-Control-Allow-Credentials", "true")
		}
		if len(o.ExposeHeaders) > 0 {
			res.Set("Access-Control-Expose-Headers", strings.Join(o.ExposeHeaders, ","))
		}
		if req.Method() == "OPTIONS" {
			res.Set("Access-Control-Allow-Methods", methods)
			if len(o.AllowHeaders) > 0 {
				res.Set("Access-Control-Allow-Headers", strings.Join(o.AllowHeaders, ","))
			} else if reqH := req.Get("Access-Control-Request-Headers"); reqH != "" {
				res.Set("Access-Control-Allow-Headers", reqH)
			}
			if o.MaxAge > 0 {
				res.Set("Access-Control-Max-Age", strconv.Itoa(o.MaxAge))
			}
			res.SendStatus(http.StatusNoContent)
			return
		}
		next()
	}
}

func corsResolveOrigin(allowed []string, origin string) string {
	if len(allowed) == 0 {
		return "*"
	}
	for _, a := range allowed {
		if a == "*" {
			return "*"
		}
		if a == origin {
			return origin
		}
	}
	return ""
}

// SecurityOptions configures the [SecurityHeaders] middleware. The zero value
// applies sensible, helmet-like defaults.
type SecurityOptions struct {
	// ContentSecurityPolicy sets Content-Security-Policy. Empty leaves it unset;
	// use "-" to explicitly omit it.
	ContentSecurityPolicy string
	// FrameOptions sets X-Frame-Options; defaults to "SAMEORIGIN". Use "-" to
	// omit.
	FrameOptions string
	// HSTSMaxAge, when > 0, sets Strict-Transport-Security with that max-age.
	HSTSMaxAge int
	// ReferrerPolicy sets Referrer-Policy; defaults to "no-referrer". Use "-" to
	// omit.
	ReferrerPolicy string
	// DisableNoSniff omits the X-Content-Type-Options: nosniff header (set by
	// default).
	DisableNoSniff bool
}

// SecurityHeaders returns middleware that sets a baseline of security-related
// response headers (X-Content-Type-Options, X-Frame-Options, Referrer-Policy
// and optionally Strict-Transport-Security and Content-Security-Policy),
// analogous to helmet for Express.
func SecurityHeaders(opts ...SecurityOptions) Handler {
	o := SecurityOptions{}
	if len(opts) > 0 {
		o = opts[0]
	}
	frame := o.FrameOptions
	if frame == "" {
		frame = "SAMEORIGIN"
	}
	ref := o.ReferrerPolicy
	if ref == "" {
		ref = "no-referrer"
	}
	return func(req *Request, res *Response, next Next) {
		if !o.DisableNoSniff {
			res.Set("X-Content-Type-Options", "nosniff")
		}
		if frame != "-" {
			res.Set("X-Frame-Options", frame)
		}
		if ref != "-" {
			res.Set("Referrer-Policy", ref)
		}
		if o.HSTSMaxAge > 0 {
			res.Set("Strict-Transport-Security", "max-age="+strconv.Itoa(o.HSTSMaxAge)+"; includeSubDomains")
		}
		if o.ContentSecurityPolicy != "" && o.ContentSecurityPolicy != "-" {
			res.Set("Content-Security-Policy", o.ContentSecurityPolicy)
		}
		next()
	}
}

// RequestID returns middleware that ensures every request carries an
// X-Request-Id: it reuses an inbound one or generates a random 128-bit id, and
// echoes it on both the request and the response.
func RequestID() Handler {
	return func(req *Request, res *Response, next Next) {
		id := req.Get("X-Request-Id")
		if id == "" {
			id = expressRandomID()
		}
		req.Raw.Header.Set("X-Request-Id", id)
		res.Set("X-Request-Id", id)
		next()
	}
}

func expressRandomID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "0000000000000000"
	}
	return hex.EncodeToString(b)
}

// ResponseTime returns middleware that measures how long the downstream
// handlers take and sets the elapsed milliseconds on the X-Response-Time header
// just before the response is written.
func ResponseTime() Handler {
	return func(req *Request, res *Response, next Next) {
		start := time.Now()
		res.OnBeforeWrite(func() {
			ms := float64(time.Since(start).Microseconds()) / 1000.0
			res.Set("X-Response-Time", strconv.FormatFloat(ms, 'f', 3, 64)+"ms")
		})
		next()
	}
}

// NoCache returns middleware that sets response headers instructing clients and
// proxies not to cache the response.
func NoCache() Handler {
	return func(req *Request, res *Response, next Next) {
		res.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		res.Set("Pragma", "no-cache")
		res.Set("Expires", "0")
		res.Set("Surrogate-Control", "no-store")
		next()
	}
}

// MethodOverride returns middleware that lets a POST request masquerade as
// another method via the X-HTTP-Method-Override header or a `_method` query
// parameter — useful for HTML forms that cannot issue PUT/DELETE directly.
func MethodOverride() Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Method() == "POST" {
			m := req.Get("X-HTTP-Method-Override")
			if m == "" {
				m = req.Query("_method")
			}
			if m != "" {
				req.Raw.Method = strings.ToUpper(m)
			}
		}
		next()
	}
}

// BasicAuthOptions configures the [BasicAuth] middleware.
type BasicAuthOptions struct {
	// Users maps usernames to passwords. Used when Validate is nil.
	Users map[string]string
	// Validate, when set, authenticates a (username, password) pair and takes
	// precedence over Users.
	Validate func(user, pass string) bool
	// Realm is the authentication realm sent in the challenge; defaults to
	// "Restricted".
	Realm string
}

// BasicAuth returns middleware that enforces HTTP Basic authentication. It
// compares credentials in constant time when using the Users map, and responds
// with 401 and a WWW-Authenticate challenge when authentication fails.
func BasicAuth(opts BasicAuthOptions) Handler {
	realm := opts.Realm
	if realm == "" {
		realm = "Restricted"
	}
	return func(req *Request, res *Response, next Next) {
		user, pass, ok := expressParseBasicAuth(req.Get("Authorization"))
		authed := false
		if ok {
			if opts.Validate != nil {
				authed = opts.Validate(user, pass)
			} else if want, exists := opts.Users[user]; exists {
				authed = subtle.ConstantTimeCompare([]byte(pass), []byte(want)) == 1
			}
		}
		if !authed {
			res.Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			res.Status(http.StatusUnauthorized).Send("Unauthorized")
			return
		}
		next()
	}
}

func expressParseBasicAuth(header string) (user, pass string, ok bool) {
	const prefix = "Basic "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", "", false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[len(prefix):]))
	if err != nil {
		return "", "", false
	}
	i := strings.IndexByte(string(raw), ':')
	if i < 0 {
		return "", "", false
	}
	return string(raw[:i]), string(raw[i+1:]), true
}

// BodyLimit returns middleware that rejects requests whose body exceeds
// maxBytes with 413 Payload Too Large, and caps reads of the body so a lying
// Content-Length cannot exhaust memory.
func BodyLimit(maxBytes int64) Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Raw.ContentLength > maxBytes {
			res.Status(http.StatusRequestEntityTooLarge).Send("Payload Too Large")
			return
		}
		req.Raw.Body = http.MaxBytesReader(res.Writer, req.Raw.Body, maxBytes)
		next()
	}
}

// RealIP returns middleware that rewrites the request's remote address from the
// X-Real-IP or the first X-Forwarded-For entry, so downstream code and req.IP
// observe the client address when running behind a trusted proxy.
func RealIP() Handler {
	return func(req *Request, res *Response, next Next) {
		if ip := strings.TrimSpace(req.Get("X-Real-IP")); ip != "" {
			req.Raw.RemoteAddr = ip + ":0"
		} else if fwd := req.Get("X-Forwarded-For"); fwd != "" {
			if first := strings.TrimSpace(strings.Split(fwd, ",")[0]); first != "" {
				req.Raw.RemoteAddr = first + ":0"
			}
		}
		next()
	}
}

// Favicon returns middleware that serves the icon at path for GET/HEAD requests
// to /favicon.ico (with a one-day cache), short-circuiting the rest of the
// stack. Other requests pass through.
func Favicon(path string) Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Path() != "/favicon.ico" {
			next()
			return
		}
		if req.Method() != "GET" && req.Method() != "HEAD" {
			res.Status(http.StatusMethodNotAllowed).Set("Allow", "GET, HEAD").End()
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			res.SendStatus(http.StatusNotFound)
			return
		}
		res.Set("Content-Type", "image/x-icon").Set("Cache-Control", "public, max-age=86400")
		res.Send(data)
	}
}

// RateLimitOptions configures the [RateLimit] middleware.
type RateLimitOptions struct {
	// Max is the maximum number of requests per key per window. Defaults to 60.
	Max int
	// Window is the length of the fixed rate-limiting window. Defaults to one
	// minute.
	Window time.Duration
	// KeyFunc derives the rate-limit bucket key from a request; defaults to the
	// client IP.
	KeyFunc func(*Request) string
	// Message is the 429 response body; defaults to "Too Many Requests".
	Message string
}

// RateLimit returns middleware implementing a fixed-window rate limiter keyed by
// client IP (or a custom key). It sets X-RateLimit-Limit/Remaining/Reset headers
// and answers 429 with a Retry-After header once the limit is exceeded.
func RateLimit(opts RateLimitOptions) Handler {
	if opts.Max <= 0 {
		opts.Max = 60
	}
	if opts.Window <= 0 {
		opts.Window = time.Minute
	}
	msg := opts.Message
	if msg == "" {
		msg = "Too Many Requests"
	}
	keyFunc := opts.KeyFunc
	if keyFunc == nil {
		keyFunc = func(r *Request) string { return r.IP() }
	}
	type entry struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*entry{}
	return func(req *Request, res *Response, next Next) {
		key := keyFunc(req)
		now := time.Now()
		mu.Lock()
		e := buckets[key]
		if e == nil || now.After(e.reset) {
			e = &entry{reset: now.Add(opts.Window)}
			buckets[key] = e
		}
		e.count++
		count := e.count
		reset := e.reset
		mu.Unlock()

		remaining := opts.Max - count
		if remaining < 0 {
			remaining = 0
		}
		res.Set("X-RateLimit-Limit", strconv.Itoa(opts.Max))
		res.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		res.Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
		if count > opts.Max {
			res.Set("Retry-After", strconv.Itoa(int(time.Until(reset).Seconds())+1))
			res.Status(http.StatusTooManyRequests).Send(msg)
			return
		}
		next()
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) { return w.gz.Write(b) }

func (w *gzipResponseWriter) Flush() {
	_ = w.gz.Flush()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Compress returns middleware that gzip-encodes the response body when the
// client advertises gzip support via Accept-Encoding. It sets Content-Encoding,
// varies on Accept-Encoding and removes any premature Content-Length.
func Compress() Handler {
	return func(req *Request, res *Response, next Next) {
		if !strings.Contains(req.Get("Accept-Encoding"), "gzip") {
			next()
			return
		}
		gz := gzip.NewWriter(res.Writer)
		defer gz.Close()
		res.Set("Content-Encoding", "gzip")
		res.Vary("Accept-Encoding")
		res.RemoveHeader("Content-Length")
		res.Writer = &gzipResponseWriter{ResponseWriter: res.Writer, gz: gz}
		next()
	}
}

// Timeout returns middleware that attaches a deadline of d to the request
// context, so downstream handlers that honour req.Raw.Context() abort when the
// deadline elapses.
func Timeout(d time.Duration) Handler {
	return func(req *Request, res *Response, next Next) {
		ctx, cancel := context.WithTimeout(req.Raw.Context(), d)
		defer cancel()
		req.Raw = req.Raw.WithContext(ctx)
		next()
	}
}

// SetHeaders returns middleware that sets a fixed set of response headers on
// every request.
func SetHeaders(headers map[string]string) Handler {
	return func(req *Request, res *Response, next Next) {
		for k, v := range headers {
			res.Set(k, v)
		}
		next()
	}
}

// HealthCheck returns middleware that answers GET requests to path with 200 and
// the given body (or "OK" when empty), short-circuiting the stack — a ready-made
// liveness/readiness endpoint.
func HealthCheck(path, body string) Handler {
	if body == "" {
		body = "OK"
	}
	return func(req *Request, res *Response, next Next) {
		if req.Path() == path && (req.Method() == "GET" || req.Method() == "HEAD") {
			res.Status(http.StatusOK).Send(body)
			return
		}
		next()
	}
}

// RedirectToHTTPS returns middleware that 301-redirects insecure requests to
// their https:// equivalent. Requests already served over TLS pass through.
func RedirectToHTTPS() Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Secure() {
			next()
			return
		}
		res.Redirect(http.StatusMovedPermanently, "https://"+req.Raw.Host+req.OriginalURL())
	}
}

// Compose combines several handlers into a single Handler that runs them in
// order, each able to short-circuit by not calling next. It is handy for
// bundling a fixed middleware stack into one value.
func Compose(handlers ...Handler) Handler {
	return func(req *Request, res *Response, next Next) {
		var run func(i int)
		run = func(i int) {
			if i >= len(handlers) {
				next()
				return
			}
			handlers[i](req, res, func(err ...error) {
				if len(err) > 0 && err[0] != nil {
					next(err[0])
					return
				}
				run(i + 1)
			})
		}
		run(0)
	}
}

// When returns middleware that runs h only when pred reports true for the
// request; otherwise it passes straight through.
func When(pred func(*Request) bool, h Handler) Handler {
	return func(req *Request, res *Response, next Next) {
		if pred(req) {
			h(req, res, next)
			return
		}
		next()
	}
}
