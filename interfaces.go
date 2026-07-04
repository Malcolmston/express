package express

import "net/http"

// RouteRegistrar is the route-registration surface shared by *Router and
// *Application. An Application embeds *Router, so both types expose exactly the
// same set of routing methods; capturing them in one interface documents that
// shared contract and lets helpers accept "anything routes can be registered
// on" — whether that is the top-level app or a sub-router — without depending
// on the concrete type.
//
// Every method returns *Router (the receiver, or the embedded router for an
// Application) so registrations remain chainable.
type RouteRegistrar interface {
	// Use mounts middleware, an error handler, or a sub-router, optionally
	// scoped to a leading path prefix.
	Use(args ...any) *Router
	// All registers handlers that run for every HTTP method matching path.
	All(path string, handlers ...Handler) *Router
	// Get registers handlers for GET requests to path.
	Get(path string, handlers ...Handler) *Router
	// Post registers handlers for POST requests to path.
	Post(path string, handlers ...Handler) *Router
	// Put registers handlers for PUT requests to path.
	Put(path string, handlers ...Handler) *Router
	// Delete registers handlers for DELETE requests to path.
	Delete(path string, handlers ...Handler) *Router
	// Patch registers handlers for PATCH requests to path.
	Patch(path string, handlers ...Handler) *Router
	// Head registers handlers for HEAD requests to path.
	Head(path string, handlers ...Handler) *Router
	// Options registers handlers for OPTIONS requests to path.
	Options(path string, handlers ...Handler) *Router
	// Query registers handlers for the QUERY method.
	Query(path string, handlers ...Handler) *Router
	// Param registers a route-parameter callback.
	Param(name string, fn ParamHandler) *Router
	// Route returns a Route for registering several methods on one path.
	Route(path string) *Route
}

// Compile-time assertions documenting the contracts the framework's core types
// already fulfill.
var (
	// Both the app and any router expose the same registration surface.
	_ RouteRegistrar = (*Router)(nil)
	_ RouteRegistrar = (*Application)(nil)

	// Application is a drop-in net/http handler, so it can be served by any
	// *http.Server, wrapped by other middleware, or driven with httptest.
	_ http.Handler = (*Application)(nil)
)
