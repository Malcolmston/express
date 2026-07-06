// Package express is a fast, minimalist web framework for Go, modeled after the
// Node.js Express framework. It is a stdlib-only port: everything is built on
// net/http and the standard library, with no third-party dependencies, so it
// drops cleanly into any Go project and interoperates with the wider net/http
// ecosystem. The goal is to give Go programmers the ergonomics that made
// Express popular — declarative routing, a chainable request/response API, and a
// composable middleware stack — while remaining idiomatic Go underneath.
//
// The core model mirrors Express's four building blocks: the Application, the
// Router, the Request, and the Response. express.New returns an *Application,
// which embeds a *Router; because of that embedding every routing method lives
// directly on the app — app.Use, app.Get, app.Post, app.All, and so on all work
// on the returned value, exactly as app.get / app.use do in JavaScript. An
// *Application also satisfies http.Handler through ServeHTTP, so it can be
// handed to http.ListenAndServe, wrapped by other net/http middleware, or
// driven directly in tests with net/http/httptest. The convenience Listen and
// ListenWithServer helpers exist for the common serving cases.
//
// A handler has the type express.Handler, defined as
// func(req *Request, res *Response, next Next). Every handler receives the
// wrapped request, the wrapped response, and a Next function used to pass
// control down the stack. Calling next() continues to the next matching layer;
// calling next(err) diverts to error-handling middleware, which is any function
// with the ErrorHandler shape func(err error, req *Request, res *Response, next Next).
// Request exposes Express-style accessors — Params for captured route
// parameters, Query for query-string values, Get for headers, Body and the
// Parse helpers for payloads — while Response offers chainable writers such as
// Status, Set, Send, JSON, Redirect, and Cookie. This keeps handler code terse
// and reads much like its JavaScript counterpart.
//
// Routing follows Express's path syntax. Patterns support named parameters
// (":id"), optional parameters (":id?"), custom regular-expression constraints
// (":id(\\d+)"), and wildcards ("*"), all compiled into standard library
// regexp under the hood. Routers are first-class: a Router created with
// NewRouter can register its own middleware and routes and then be mounted on
// the app (or another router) with app.Use("/prefix", subRouter), enabling
// modular, prefix-scoped route trees. RouterOptions mirrors Express's
// case-sensitive, strict, and merge-params toggles, and app.Param registers
// parameter-preprocessing callbacks that run once per request before the
// matched route.
//
// View rendering is included as well. app.Engine registers a template engine
// for a file extension, res.Render resolves a view against the configured
// "views" directories and "view engine" setting and streams the rendered HTML,
// and a cache-aware html/template engine is wired up by default for ".html" and
// ".tmpl" files. Application settings (Set, GetSetting, Enable, Disable,
// Enabled, Disabled) and per-request Locals round out the Express feature set,
// giving handlers a place to stash configuration and request-scoped data.
//
// Beyond the core framework, this module ports a large collection of the npm
// utility libraries that typically accompany an Express application. The
// middleware subpackages (under middleware/) reimplement popular connect-style
// middleware — body parsing, CORS, compression, sessions, rate limiting, and
// more — as express.Handler values, while standalone utility packages port
// libraries such as accepts, cookie, content-type, content-disposition,
// bytes, ms, camelcase, and dozens of others as ordinary, dependency-free Go
// packages. Each utility package documents the npm package it corresponds to
// and the parity it targets, so an Express codebase can be translated to Go
// piece by piece. The aim throughout is Express.js parity: familiar names,
// familiar semantics, and behavior that matches the Node originals closely
// enough to port real applications with minimal surprise.
//
// A minimal application looks like this:
//
//	app := express.New()
//
//	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
//		res.Send("Hello World")
//	})
//
//	log.Fatal(app.Listen(":3000"))
package express

import (
	"net/http"
)

// Application is the top-level express app. It embeds a Router so every
// routing method (Get, Post, Use, ...) is available directly on the app, and
// adds server lifecycle helpers such as Listen.
type Application struct {
	*Router

	// settings holds application-level configuration set via Set/Get.
	settings map[string]any

	// locals are variables shared across the lifetime of the app and made
	// available to every request via res.Locals merged at request time.
	locals map[string]any

	// viewEngines maps a file extension (".html", ".tmpl", ...) to the
	// template engine used to render views for res.Render.
	viewEngines map[string]EngineFunc

	// viewCache memoises resolved view paths and (for the built-in engine)
	// compiled templates when the "view cache" setting is enabled.
	viewCache *viewCache
}

// New creates a new express Application.
func New() *Application {
	app := &Application{
		Router:      NewRouter(),
		settings:    make(map[string]any),
		locals:      make(map[string]any),
		viewEngines: make(map[string]EngineFunc),
		viewCache:   newViewCache(),
	}
	// Sensible defaults mirroring Express.
	app.settings["env"] = "development"
	app.settings["x-powered-by"] = true
	app.settings["views"] = "views"
	app.settings["view engine"] = "html"
	// Express enables the view cache by default outside development.
	app.settings["view cache"] = false
	// Register the built-in html/template engine (cache-aware).
	app.Engine(".html", app.htmlTemplateEngine)
	app.Engine(".tmpl", app.htmlTemplateEngine)
	return app
}

// Set assigns a setting name to a value, returning the app for chaining.
func (app *Application) Set(name string, value any) *Application {
	app.settings[name] = value
	return app
}

// GetSetting returns the value of a previously configured setting.
func (app *Application) GetSetting(name string) any {
	return app.settings[name]
}

// Enabled reports whether a boolean setting is turned on.
func (app *Application) Enabled(name string) bool {
	v, ok := app.settings[name].(bool)
	return ok && v
}

// Disabled reports whether a boolean setting is turned off.
func (app *Application) Disabled(name string) bool {
	return !app.Enabled(name)
}

// Enable turns a boolean setting on.
func (app *Application) Enable(name string) *Application { return app.Set(name, true) }

// Disable turns a boolean setting off.
func (app *Application) Disable(name string) *Application { return app.Set(name, false) }

// Locals returns the application-wide locals map.
func (app *Application) Locals() map[string]any { return app.locals }

// ServeHTTP makes Application satisfy http.Handler so it can be handed to any
// net/http server, httptest, or wrapped by other middleware.
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := newRequest(r, app)
	res := newResponse(w, req, app)
	app.Router.handle(req, res, func(err ...error) {
		// Final fall-through handler: unmatched route or unhandled error.
		if len(err) > 0 && err[0] != nil {
			res.finalError(err[0])
			return
		}
		if !res.written {
			res.Status(http.StatusNotFound).Send("Cannot " + req.Method() + " " + req.Path())
		}
	})
}

// Listen binds and listens for connections on the given address, e.g. ":3000".
// It blocks until the server exits and returns any error from ListenAndServe.
func (app *Application) Listen(addr string) error {
	server := &http.Server{Addr: addr, Handler: app}
	return server.ListenAndServe()
}

// ListenWithServer lets callers supply a preconfigured *http.Server (for TLS,
// custom timeouts, etc.). The app is installed as the server's handler.
func (app *Application) ListenWithServer(server *http.Server) error {
	server.Handler = app
	return server.ListenAndServe()
}
