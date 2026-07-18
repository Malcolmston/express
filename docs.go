package express

import (
	"sort"
	"strings"
)

// RouteInfo describes a single HTTP route discovered by introspecting an
// application's router tree. It is the raw material from which the OpenAPI,
// Postman and other specifications are generated.
type RouteInfo struct {
	// Method is the upper-case HTTP method, e.g. "GET" or "POST".
	Method string
	// Path is the Express-style path pattern as registered, e.g.
	// "/users/:id". Use [RouteInfo.OpenAPIPath] for the templated form.
	Path string
	// Params are the path parameter names in order of appearance.
	Params []string
}

// OpenAPIPath returns Path with Express-style parameters (":id") rewritten to
// OpenAPI/URI-template form ("{id}") and the wildcard "*" rewritten to
// "{wildcard}", so the value is a valid OpenAPI path key.
func (ri RouteInfo) OpenAPIPath() string {
	return expressPathToTemplate(ri.Path)
}

// Routes returns every HTTP route registered on the application, including
// those contributed by mounted sub-routers (with their mount prefixes
// applied). Middleware layers (which match every method) are omitted; only
// method-bound routes are reported. The result is sorted by path then method
// and de-duplicated, so a route with several handlers appears once.
func (app *Application) Routes() []RouteInfo {
	var out []RouteInfo
	seen := make(map[string]bool)
	app.Router.collectRoutes("", &out, seen)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

// Routes returns every HTTP route registered on the router (and its mounted
// sub-routers), sorted and de-duplicated. It is the router-level counterpart of
// [Application.Routes].
func (r *Router) Routes() []RouteInfo {
	var out []RouteInfo
	seen := make(map[string]bool)
	r.collectRoutes("", &out, seen)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

// collectRoutes walks the router stack, following mounted sub-routers, and
// appends each unique method+path route to out.
func (r *Router) collectRoutes(prefix string, out *[]RouteInfo, seen map[string]bool) {
	for _, l := range r.stack {
		if l.mounted != nil {
			l.mounted.collectRoutes(docsJoinPath(prefix, l.pattern.rawPath), out, seen)
			continue
		}
		if l.method == "" {
			continue // middleware, not a route
		}
		full := docsJoinPath(prefix, l.pattern.rawPath)
		key := l.method + " " + full
		if seen[key] {
			continue
		}
		seen[key] = true
		*out = append(*out, RouteInfo{
			Method: l.method,
			Path:   full,
			Params: docsParamNames(full),
		})
	}
}

// docsJoinPath joins a mount prefix with a sub-path, collapsing duplicate
// slashes and treating "" and "/" as the root.
func docsJoinPath(prefix, sub string) string {
	clean := func(s string) string {
		if s == "" {
			return "/"
		}
		return s
	}
	prefix, sub = clean(prefix), clean(sub)
	if prefix == "/" {
		return sub
	}
	p := strings.TrimRight(prefix, "/")
	if sub == "/" {
		if p == "" {
			return "/"
		}
		return p
	}
	if !strings.HasPrefix(sub, "/") {
		sub = "/" + sub
	}
	return p + sub
}

// docsParamNames extracts the ordered path-parameter names from an Express-style
// path. ":name" and ":name?" yield "name"; a bare "*" yields "wildcard".
func docsParamNames(path string) []string {
	var names []string
	i := 0
	for i < len(path) {
		switch path[i] {
		case ':':
			j := i + 1
			for j < len(path) && isNameChar(path[j]) {
				j++
			}
			if j > i+1 {
				names = append(names, path[i+1:j])
			}
			i = j
		case '*':
			names = append(names, "wildcard")
			i++
		default:
			i++
		}
	}
	return names
}

// expressPathToTemplate rewrites an Express path to OpenAPI/URI-template form:
// ":name" and ":name?" become "{name}", and "*" becomes "{wildcard}".
func expressPathToTemplate(path string) string {
	var b strings.Builder
	i := 0
	for i < len(path) {
		switch path[i] {
		case ':':
			j := i + 1
			for j < len(path) && isNameChar(path[j]) {
				j++
			}
			b.WriteByte('{')
			b.WriteString(path[i+1 : j])
			b.WriteByte('}')
			i = j
			if i < len(path) && path[i] == '?' {
				i++ // drop the optional marker
			}
		case '*':
			b.WriteString("{wildcard}")
			i++
		default:
			b.WriteByte(path[i])
			i++
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Registry and public configuration API
// ---------------------------------------------------------------------------

// docsRegistry stores the documentation metadata attached to an application.
type docsRegistry struct {
	describes map[string]RouteDoc   // key: "METHOD /path"
	channels  map[string]ChannelDoc // event/socket channels for AsyncAPI
	opts      DocsOptions
}

func (app *Application) docsReg() *docsRegistry {
	if app.docs == nil {
		app.docs = &docsRegistry{
			describes: make(map[string]RouteDoc),
			channels:  make(map[string]ChannelDoc),
		}
	}
	return app.docs
}

// RouteDoc carries optional, human-authored metadata for a single route that
// enriches the generated OpenAPI operation beyond what introspection alone can
// infer (which is just the method, path and path parameters).
type RouteDoc struct {
	// Summary is a short one-line description of the operation.
	Summary string
	// Description is a longer explanation (CommonMark is allowed by OpenAPI).
	Description string
	// Tags group related operations in the UI.
	Tags []string
	// OperationID is a unique, machine-friendly identifier for the operation.
	OperationID string
	// Deprecated marks the operation as deprecated.
	Deprecated bool
	// Parameters describes query, header, cookie (and extra path) parameters.
	// Path parameters discovered from the route are added automatically and
	// need not be repeated here.
	Parameters []ParamDoc
	// RequestBody documents the request payload, if any.
	RequestBody *BodyDoc
	// Responses maps a status code (e.g. "200") or "default" to a response.
	// When empty, a generic 200 response is generated.
	Responses map[string]ResponseDoc
}

// ParamDoc documents a single operation parameter.
type ParamDoc struct {
	// Name is the parameter name.
	Name string
	// In is the parameter location: "query", "header", "path" or "cookie".
	In string
	// Description explains the parameter.
	Description string
	// Required marks the parameter as mandatory. Path parameters are always
	// required regardless of this field.
	Required bool
	// Schema is a JSON-Schema object for the parameter's value. When nil a
	// permissive {"type":"string"} schema is used.
	Schema map[string]any
}

// BodyDoc documents a request body.
type BodyDoc struct {
	// Description explains the body.
	Description string
	// Required marks the body as mandatory.
	Required bool
	// ContentType is the media type; it defaults to "application/json".
	ContentType string
	// Schema is a JSON-Schema object describing the payload.
	Schema map[string]any
	// Example is an optional example value serialised alongside the schema.
	Example any
}

// ResponseDoc documents a single response.
type ResponseDoc struct {
	// Description explains the response; OpenAPI requires it to be non-empty,
	// so a default is substituted when blank.
	Description string
	// ContentType is the media type of the body; defaults to
	// "application/json" when a Schema or Example is present.
	ContentType string
	// Schema is a JSON-Schema object describing the response body.
	Schema map[string]any
	// Example is an optional example value.
	Example any
}

// Describe attaches documentation metadata to the route identified by method
// and path (the same path string used to register it, e.g. "/users/:id"). It
// may be called before or after the route is registered and returns the app for
// chaining. Method is matched case-insensitively.
func (app *Application) Describe(method, path string, doc RouteDoc) *Application {
	reg := app.docsReg()
	reg.describes[strings.ToUpper(method)+" "+path] = doc
	return app
}

// Channel registers an event/message channel (for example a Socket.IO or
// WebSocket topic) that is included in the generated AsyncAPI document. It
// returns the app for chaining.
func (app *Application) Channel(name string, doc ChannelDoc) *Application {
	reg := app.docsReg()
	reg.channels[name] = doc
	return app
}

// DocsOptions configures the documentation endpoints mounted by
// [Application.Docs]. The zero value is valid; every field has a sensible
// default. Set a *Path field to "-" to disable that individual endpoint
// (leaving it "" keeps the default).
type DocsOptions struct {
	// Title, Version and Description populate the spec's info block. Title
	// defaults to "API", Version to the app's "version" setting or "1.0.0".
	Title       string
	Version     string
	Description string
	// Servers are base URLs advertised in the OpenAPI servers block.
	Servers []string

	// UIPath serves the Swagger UI HTML page (default "/docs").
	UIPath string
	// RedocPath serves the ReDoc HTML page (default "/redoc").
	RedocPath string
	// OpenAPIPath serves the OpenAPI 3.1 document as JSON (default
	// "/openapi.json").
	OpenAPIPath string
	// OpenAPIYAMLPath serves the OpenAPI document as YAML (default
	// "/openapi.yaml").
	OpenAPIYAMLPath string
	// AsyncAPIPath serves the AsyncAPI 2.6 document as JSON (default
	// "/asyncapi.json"). Only meaningful when channels are registered.
	AsyncAPIPath string
	// PostmanPath serves a Postman v2.1 collection (default
	// "/postman.json").
	PostmanPath string

	// AssetBaseURL is the base URL for the Swagger UI / ReDoc browser assets.
	// It defaults to the jsDelivr CDN. Point it at a self-hosted copy for
	// offline use.
	AssetBaseURL string

	// Enrich, when set, is called for every operation as the OpenAPI document
	// is built, allowing programmatic customisation beyond Describe.
	Enrich func(route RouteInfo, op *Operation)
}

func (o *DocsOptions) withDefaults(app *Application) {
	if o.Title == "" {
		o.Title = "API"
	}
	if o.Version == "" {
		if v, ok := app.settings["version"].(string); ok && v != "" {
			o.Version = v
		} else {
			o.Version = "1.0.0"
		}
	}
	setDefault(&o.UIPath, "/docs")
	setDefault(&o.RedocPath, "/redoc")
	setDefault(&o.OpenAPIPath, "/openapi.json")
	setDefault(&o.OpenAPIYAMLPath, "/openapi.yaml")
	setDefault(&o.AsyncAPIPath, "/asyncapi.json")
	setDefault(&o.PostmanPath, "/postman.json")
	if o.AssetBaseURL == "" {
		o.AssetBaseURL = "https://cdn.jsdelivr.net/npm"
	}
}

func setDefault(p *string, def string) {
	if *p == "" {
		*p = def
	}
}

// docsEnabled reports whether an endpoint path is active. An empty path means
// "use the default" (already substituted by withDefaults); the sentinel "-"
// disables the endpoint.
func docsEnabled(path string) bool {
	return path != "" && path != "-"
}

// Docs enables automatic API documentation for the application. It generates an
// OpenAPI 3.1 specification from the registered routes (enriched by any
// [Application.Describe] calls), an AsyncAPI document from any
// [Application.Channel] calls, and a Postman collection, and mounts endpoints
// that serve them together with interactive Swagger UI and ReDoc pages.
//
// The specifications are (re)built on each request, so routes registered after
// Docs still appear. Pass a DocsOptions to customise titles, servers and
// endpoint paths; the zero value uses the documented defaults:
//
//	app.Docs()                                  // defaults
//	app.Docs(express.DocsOptions{Title: "..."}) // customised
//
// Docs returns the app for chaining.
func (app *Application) Docs(opts ...DocsOptions) *Application {
	reg := app.docsReg()
	if len(opts) > 0 {
		reg.opts = opts[0]
	}
	reg.opts.withDefaults(app)
	o := &reg.opts

	// OpenAPI JSON.
	if docsEnabled(o.OpenAPIPath) {
		app.Get(o.OpenAPIPath, func(req *Request, res *Response, next Next) {
			res.Set("Access-Control-Allow-Origin", "*")
			res.JSON(app.OpenAPI())
		})
	}
	// OpenAPI YAML.
	if docsEnabled(o.OpenAPIYAMLPath) {
		app.Get(o.OpenAPIYAMLPath, func(req *Request, res *Response, next Next) {
			res.Set("Access-Control-Allow-Origin", "*")
			res.Type("application/yaml").Send(app.OpenAPIYAML())
		})
	}
	// AsyncAPI JSON.
	if docsEnabled(o.AsyncAPIPath) {
		app.Get(o.AsyncAPIPath, func(req *Request, res *Response, next Next) {
			res.Set("Access-Control-Allow-Origin", "*")
			res.JSON(app.AsyncAPI())
		})
	}
	// Postman collection.
	if docsEnabled(o.PostmanPath) {
		app.Get(o.PostmanPath, func(req *Request, res *Response, next Next) {
			res.Set("Access-Control-Allow-Origin", "*")
			res.JSON(app.PostmanCollection())
		})
	}
	// Swagger UI.
	if docsEnabled(o.UIPath) {
		specURL := o.OpenAPIPath
		app.Get(o.UIPath, func(req *Request, res *Response, next Next) {
			res.Type("text/html").Send(swaggerUIHTML(o.Title, specURL, o.AssetBaseURL))
		})
	}
	// ReDoc.
	if docsEnabled(o.RedocPath) {
		specURL := o.OpenAPIPath
		app.Get(o.RedocPath, func(req *Request, res *Response, next Next) {
			res.Type("text/html").Send(redocHTML(o.Title, specURL, o.AssetBaseURL))
		})
	}
	return app
}
