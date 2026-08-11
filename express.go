package express

import (
	"html"
	"net/http"
	"strings"
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

	// docs holds API-documentation metadata (route descriptions, event
	// channels and options) registered via Describe/Channel/Docs. It is
	// created lazily so apps that never use the docs feature pay nothing.
	docs *docsRegistry
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
	// Forwarding headers (X-Forwarded-For / -Proto / -Host) are attacker
	// controlled unless a trusted proxy terminates the connection, so, exactly
	// as in Express, they are ignored until the operator opts in with
	// app.Enable("trust proxy"). See Request.IP and Request.Protocol.
	app.settings["trust proxy"] = false
	// Register the built-in html/template engine (cache-aware).
	app.Engine(".html", app.htmlTemplateEngine)
	app.Engine(".tmpl", app.htmlTemplateEngine)
	return app
}

// Set assigns a setting name to a value, returning the app for chaining.
//
// The two routing settings Express reads lazily when it builds its router —
// "strict routing" and "case sensitive routing" — are wired straight through to
// the application's root router, so routes registered after the setting honour
// it (Express itself only reads them once, at first route registration).
func (app *Application) Set(name string, value any) *Application {
	app.settings[name] = value
	if b, ok := value.(bool); ok {
		switch name {
		case "strict routing":
			app.Router.opts.Strict = b
		case "case sensitive routing":
			app.Router.opts.CaseSensitive = b
		}
	}
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
	req.errHandlers = app.Router.hasErrorHandler()
	res := newResponse(w, req, app)
	app.Router.handle(req, res, func(err ...error) {
		// Final fall-through handler: unmatched route or unhandled error.
		if len(err) > 0 && err[0] != nil {
			res.finalError(err[0])
			return
		}
		if !res.written {
			res.defaultNotFound(req.Method(), req.Path())
		}
	})
}

// defaultNotFound renders the framework's own unmatched-route response, matching
// Express's finalhandler: an HTML error page with a Content-Security-Policy that
// forbids loading any subresource and X-Content-Type-Options: nosniff, so the
// echoed method/path can never be interpreted as active content.
func (res *Response) defaultNotFound(method, path string) {
	msg := "Cannot " + method + " " + path
	body := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n<title>Error</title>\n</head>\n<body>\n<pre>" +
		html.EscapeString(msg) + "</pre>\n</body>\n</html>\n"
	res.Set("Content-Security-Policy", "default-src 'none'")
	res.Set("X-Content-Type-Options", "nosniff")
	res.Status(http.StatusNotFound).Type("html")
	res.writeHeaderOnce()
	if !strings.EqualFold(method, "HEAD") {
		_, _ = res.Writer.Write([]byte(body))
	}
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
