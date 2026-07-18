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
