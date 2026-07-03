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
	stack          []*layer
	paramCallbacks map[string][]ParamHandler
}

// ParamHandler processes a captured route parameter before the route's handlers
// run. It receives the parameter's value and, like any handler, must call next
// to continue (or next(err) to fail).
type ParamHandler func(req *Request, res *Response, next Next, value string)

// Param registers a callback that runs when name is captured as a route
// parameter, once per request, before the matched route's handlers — the
// equivalent of Express's app.param(). Typical use is to load a record by id.
func (r *Router) Param(name string, fn ParamHandler) *Router {
	if r.paramCallbacks == nil {
		r.paramCallbacks = make(map[string][]ParamHandler)
	}
	r.paramCallbacks[name] = append(r.paramCallbacks[name], fn)
	return r
}

// Route returns a Route bound to path, allowing several HTTP methods to be
// registered for the same path in a chain — the equivalent of app.route(path):
//
//	app.Route("/users").
//		Get(list).
//		Post(create)
func (r *Router) Route(path string) *Route {
	return &Route{router: r, path: path}
}

// Route binds handlers for multiple HTTP methods to a single path.
type Route struct {
	router *Router
	path   string
}

// Get registers GET handlers on the route.
func (rt *Route) Get(h ...Handler) *Route { rt.router.Get(rt.path, h...); return rt }

// Post registers POST handlers on the route.
func (rt *Route) Post(h ...Handler) *Route { rt.router.Post(rt.path, h...); return rt }

// Put registers PUT handlers on the route.
func (rt *Route) Put(h ...Handler) *Route { rt.router.Put(rt.path, h...); return rt }

// Delete registers DELETE handlers on the route.
func (rt *Route) Delete(h ...Handler) *Route { rt.router.Delete(rt.path, h...); return rt }

// Patch registers PATCH handlers on the route.
func (rt *Route) Patch(h ...Handler) *Route { rt.router.Patch(rt.path, h...); return rt }

// All registers handlers for every method on the route.
func (rt *Route) All(h ...Handler) *Route { rt.router.All(rt.path, h...); return rt }

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

			// The match path is read live from the request so that middleware
			// which rewrites the path (via req.SetPath) re-routes subsequent
			// layers. It is relative to this router's mount point.
			basePath := req.path

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
			r.invokeWithParams(req, res, l, next)
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

// invokeWithParams runs any pending app.Param callbacks for the matched layer's
// parameters, then invokes the layer's handler. Each parameter's callbacks run
// at most once per request.
func (r *Router) invokeWithParams(req *Request, res *Response, l *layer, next Next) {
	if len(r.paramCallbacks) == 0 {
		l.handler(req, res, next)
		return
	}

	type job struct {
		value string
		fn    ParamHandler
	}
	var jobs []job
	for _, key := range l.pattern.keys {
		if req.paramDone[key] {
			continue
		}
		if fns := r.paramCallbacks[key]; len(fns) > 0 {
			for _, fn := range fns {
				jobs = append(jobs, job{value: req.params[key], fn: fn})
			}
			req.paramDone[key] = true
		}
	}
	if len(jobs) == 0 {
		l.handler(req, res, next)
		return
	}

	idx := 0
	var step Next
	step = func(errs ...error) {
		if len(errs) > 0 && errs[0] != nil {
			next(errs[0])
			return
		}
		if idx < len(jobs) {
			j := jobs[idx]
			idx++
			j.fn(req, res, step, j.value)
			return
		}
		l.handler(req, res, next)
	}
	step()
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

func isNameChar(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// compilePattern compiles an Express-style path into a regular expression plus
// the ordered parameter names. Supported syntax:
//
//	:name          a required parameter ([^/]+)
//	:name?         an optional parameter (the preceding "/" is optional too)
//	:name(\d+)     a parameter constrained by a custom regular expression
//	*              a wildcard captured as the "*" parameter
//
// A custom regex should not contain capturing groups (use (?:...)) — the
// parameter itself is the capturing group.
func compilePattern(path string, prefix bool) *pathPattern {
	if path == "" {
		path = "/"
	}
	p := &pathPattern{raw: path, prefix: prefix, rawPath: path}

	var frags []string
	i := 0
	for i < len(path) {
		c := path[i]
		switch c {
		case ':':
			// Parameter name.
			j := i + 1
			for j < len(path) && isNameChar(path[j]) {
				j++
			}
			name := path[i+1 : j]
			if name == "" {
				frags = append(frags, regexp.QuoteMeta(":"))
				i++
				continue
			}
			// Optional custom regex constraint in balanced parentheses.
			pattern := `[^/]+`
			if j < len(path) && path[j] == '(' {
				depth, k := 1, j+1
				for k < len(path) && depth > 0 {
					switch path[k] {
					case '(':
						depth++
					case ')':
						depth--
					}
					k++
				}
				pattern = path[j+1 : k-1]
				j = k
			}
			// Optional marker.
			optional := false
			if j < len(path) && path[j] == '?' {
				optional = true
				j++
			}
			group := "(" + pattern + ")"
			if optional && len(frags) > 0 && frags[len(frags)-1] == "/" {
				// Make the leading slash optional along with the parameter.
				frags[len(frags)-1] = "(?:/" + group + ")?"
			} else if optional {
				frags = append(frags, group+"?")
			} else {
				frags = append(frags, group)
			}
			p.keys = append(p.keys, name)
			i = j
		case '*':
			p.keys = append(p.keys, "*")
			frags = append(frags, "(.*)")
			i++
		case '/':
			frags = append(frags, "/")
			i++
		default:
			frags = append(frags, regexp.QuoteMeta(string(c)))
			i++
		}
	}

	body := strings.Join(frags, "")
	if prefix {
		body = strings.TrimSuffix(body, "/")
		p.re = regexp.MustCompile("^" + body + `(?:/.*)?$`)
	} else {
		p.re = regexp.MustCompile("^" + body + "/?$")
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
