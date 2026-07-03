package express

import (
	"regexp"
	"strings"
)

// Handler is the express request handler signature. Every handler receives the
// request, the response, and a Next function used to pass control to the next
// matching handler in the stack.
type Handler func(req *Request, res *Response, next Next)

// ErrorHandler is a handler that additionally receives an error. When a handler
// calls next(err), express skips ordinary handlers and invokes the next
// registered ErrorHandler.
type ErrorHandler func(err error, req *Request, res *Response, next Next)

// Next advances to the next handler. Calling it with no arguments continues the
// stack; calling it with a non-nil error jumps to error-handling middleware.
type Next func(err ...error)

// layer is a single entry in a router stack: a compiled path matcher paired
// with either a normal handler or an error handler.
type layer struct {
	method  string // "" means the layer matches any method (middleware)
	pattern *pathPattern
	handler Handler
	errh    ErrorHandler
	// mounted, when non-nil, is a sub-router mounted at pattern's prefix.
	mounted *Router
}

// Router is an isolated instance of middleware and routes. Applications embed a
// root Router, and additional routers can be created and mounted to compose
// modular route handlers.
type Router struct {
	stack []*layer
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// method registers a handler chain for an HTTP method + path.
func (r *Router) method(m, path string, handlers []Handler) *Router {
	for _, h := range handlers {
		r.stack = append(r.stack, &layer{
			method:  m,
			pattern: compilePattern(path, false),
			handler: h,
		})
	}
	return r
}

// Get registers handlers for GET requests to path.
func (r *Router) Get(path string, handlers ...Handler) *Router {
	return r.method("GET", path, handlers)
}

// Post registers handlers for POST requests to path.
func (r *Router) Post(path string, handlers ...Handler) *Router {
	return r.method("POST", path, handlers)
}

// Put registers handlers for PUT requests to path.
func (r *Router) Put(path string, handlers ...Handler) *Router {
	return r.method("PUT", path, handlers)
}

// Delete registers handlers for DELETE requests to path.
func (r *Router) Delete(path string, handlers ...Handler) *Router {
	return r.method("DELETE", path, handlers)
}

// Patch registers handlers for PATCH requests to path.
func (r *Router) Patch(path string, handlers ...Handler) *Router {
	return r.method("PATCH", path, handlers)
}

// Head registers handlers for HEAD requests to path.
func (r *Router) Head(path string, handlers ...Handler) *Router {
	return r.method("HEAD", path, handlers)
}

// Options registers handlers for OPTIONS requests to path.
func (r *Router) Options(path string, handlers ...Handler) *Router {
	return r.method("OPTIONS", path, handlers)
}

// Query registers handlers for the HTTP QUERY method
// (https://www.ietf.org/archive/id/draft-ietf-httpbis-safe-method-w-body-latest.html),
// a safe, idempotent method that carries a request body — like GET, but with a
// payload used to describe the query. Express added app.query() for it; this is
// the Go equivalent.
func (r *Router) Query(path string, handlers ...Handler) *Router {
	return r.method(MethodQuery, path, handlers)
}

// MethodQuery is the QUERY HTTP method token. It is not yet part of the
// standard library's net/http method constants, so it is defined here.
const MethodQuery = "QUERY"

// All registers handlers that run for every HTTP method matching path.
func (r *Router) All(path string, handlers ...Handler) *Router { return r.method("", path, handlers) }

// Use mounts middleware, optionally scoped to a path prefix. It accepts an
// optional leading string path followed by any mix of Handler, ErrorHandler,
// and *Router values:
//
//	app.Use(logger)                  // app-wide middleware
//	app.Use("/api", apiRouter)       // mount a sub-router at /api
//	app.Use(errorHandler)            // error-handling middleware
func (r *Router) Use(args ...any) *Router {
	path := "/"
	rest := args
	if len(args) > 0 {
		if p, ok := args[0].(string); ok {
			path = p
			rest = args[1:]
		}
	}
	for _, a := range rest {
		switch h := a.(type) {
		case Handler:
			r.stack = append(r.stack, &layer{pattern: compilePattern(path, true), handler: h})
		case func(*Request, *Response, Next):
			r.stack = append(r.stack, &layer{pattern: compilePattern(path, true), handler: Handler(h)})
		case ErrorHandler:
			r.stack = append(r.stack, &layer{pattern: compilePattern(path, true), errh: h})
		case func(error, *Request, *Response, Next):
			r.stack = append(r.stack, &layer{pattern: compilePattern(path, true), errh: ErrorHandler(h)})
		case *Router:
			r.stack = append(r.stack, &layer{pattern: compilePattern(path, true), mounted: h})
		default:
			panic("express: Use received an unsupported handler type")
		}
	}
	return r
}

// handle walks the router stack for a request, invoking matching layers in
// order. The done callback is invoked when the stack is exhausted (or an error
// escapes it) so parent routers / the app can finish the response.
func (r *Router) handle(req *Request, res *Response, done Next) {
	basePath := req.path // path relative to this router's mount point
	idx := 0
	var carriedErr error

	var next Next
	next = func(errs ...error) {
		if len(errs) > 0 && errs[0] != nil {
			carriedErr = errs[0]
		}
		for idx < len(r.stack) {
			l := r.stack[idx]
			idx++

			// Method must match for route layers (middleware has method "").
			if l.method != "" && !strings.EqualFold(l.method, req.Method()) {
				continue
			}

			params, ok := l.pattern.match(basePath)
			if !ok {
				continue
			}

			if l.mounted != nil {
				// Mount a sub-router: strip the matched prefix and delegate.
				sub := req.withMountPath(l.pattern.prefixLen(basePath), params)
				l.mounted.handle(sub, res, next)
				return
			}

			// Merge captured params into the request.
			req.mergeParams(params)

			if carriedErr != nil {
				// In error mode, only error handlers run.
				if l.errh == nil {
					continue
				}
				e := carriedErr
				carriedErr = nil
				l.errh(e, req, res, next)
				return
			}

			// Normal mode: skip error handlers.
			if l.handler == nil {
				continue
			}
			l.handler(req, res, next)
			return
		}
		// Stack exhausted.
		if carriedErr != nil {
			done(carriedErr)
			return
		}
		done()
	}

	next()
}

// ---- path pattern matching (a small path-to-regexp) -------------------------

// pathPattern compiles an Express-style path such as "/users/:id" or "/files/*"
// into a regular expression plus an ordered list of parameter names.
type pathPattern struct {
	raw     string
	re      *regexp.Regexp
	keys    []string
	prefix  bool // true for Use() layers: match the path or any sub-path
	rawPath string
}

var paramSegment = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)`)

func compilePattern(path string, prefix bool) *pathPattern {
	if path == "" {
		path = "/"
	}
	p := &pathPattern{raw: path, prefix: prefix, rawPath: path}

	// Escape regex metacharacters, then re-introduce our own tokens.
	var b strings.Builder
	b.WriteString("^")

	segments := path
	// Build regex by scanning for :params and * wildcards.
	i := 0
	for i < len(segments) {
		c := segments[i]
		switch {
		case c == ':':
			m := paramSegment.FindStringSubmatch(segments[i:])
			if m != nil {
				p.keys = append(p.keys, m[1])
				b.WriteString(`([^/]+)`)
				i += len(m[0])
				continue
			}
			b.WriteString(regexp.QuoteMeta(string(c)))
			i++
		case c == '*':
			p.keys = append(p.keys, "*")
			b.WriteString(`(.*)`)
			i++
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
			i++
		}
	}

	if prefix {
		// Prefix match: allow an exact match or a match followed by "/...".
		// Normalize a trailing slash so "/api" matches "/api" and "/api/x".
		suffix := b.String()
		suffix = strings.TrimSuffix(suffix, "/")
		p.re = regexp.MustCompile(suffix + `(?:/.*)?$`)
	} else {
		b.WriteString("/?$") // tolerate an optional trailing slash
		p.re = regexp.MustCompile(b.String())
	}
	return p
}

// match returns the captured params if path matches, plus whether it matched.
func (p *pathPattern) match(path string) (map[string]string, bool) {
	m := p.re.FindStringSubmatch(path)
	if m == nil {
		return nil, false
	}
	params := make(map[string]string, len(p.keys))
	for i, k := range p.keys {
		if i+1 < len(m) {
			params[k] = m[i+1]
		}
	}
	return params, true
}

// prefixLen returns the length of the matched prefix for a mounted router, used
// to compute the residual path handed to the sub-router.
func (p *pathPattern) prefixLen(path string) int {
	// The static prefix is everything up to the first dynamic token.
	raw := p.rawPath
	if idx := strings.IndexAny(raw, ":*"); idx >= 0 {
		raw = raw[:idx]
	}
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" || raw == "/" {
		return 0
	}
	if strings.HasPrefix(path, raw) {
		return len(raw)
	}
	return 0
}
