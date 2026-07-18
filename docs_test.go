package express

import (
	"encoding/json"
	"strings"
	"testing"
)

// sampleApp builds an app with a representative mix of routes, a mounted
// sub-router, descriptions and a channel.
func sampleApp() *Application {
	app := New()
	noop := func(req *Request, res *Response, next Next) {}

	app.Get("/health", noop)
	app.Get("/users", noop)
	app.Post("/users", noop)
	app.Get("/users/:id", noop)
	app.Delete("/users/:id", noop)

	api := NewRouter()
	api.Get("/items", noop)
	api.Get("/items/:itemId", noop)
	app.Use("/api", api)

	app.Describe("GET", "/users/:id", RouteDoc{
		Summary:     "Fetch a user",
		Description: "Returns a single user by id.",
		Tags:        []string{"users"},
		Responses: map[string]ResponseDoc{
			"200": {Description: "The user", Schema: map[string]any{"type": "object"}},
			"404": {Description: "Not found"},
		},
	})
	app.Describe("POST", "/users", RouteDoc{
		Summary: "Create a user",
		Tags:    []string{"users"},
		RequestBody: &BodyDoc{
			Required: true,
			Schema:   map[string]any{"type": "object", "required": []any{"name"}},
		},
		Parameters: []ParamDoc{
			{Name: "dryRun", In: "query", Description: "Validate only", Schema: map[string]any{"type": "boolean"}},
		},
	})

	app.Channel("chat.message", ChannelDoc{
		Description: "Chat messages",
		Subscribe:   &MessageDoc{Name: "messageReceived", Payload: map[string]any{"type": "object"}},
		Publish:     &MessageDoc{Name: "sendMessage", Payload: map[string]any{"type": "object"}, Example: map[string]any{"text": "hi"}},
	})
	return app
}

func TestRoutesIntrospection(t *testing.T) {
	app := sampleApp()
	routes := app.Routes()

	want := map[string]bool{
		"GET /health":            true,
		"GET /users":             true,
		"POST /users":            true,
		"GET /users/:id":         true,
		"DELETE /users/:id":      true,
		"GET /api/items":         true,
		"GET /api/items/:itemId": true,
	}
	got := map[string]bool{}
	for _, r := range routes {
		got[r.Method+" "+r.Path] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("Routes() missing %q; got %v", k, got)
		}
	}
	if len(routes) != len(want) {
		t.Errorf("Routes() returned %d routes, want %d: %v", len(routes), len(want), got)
	}

	// Params extracted for a mounted, parameterized route.
	for _, r := range routes {
		if r.Path == "/api/items/:itemId" {
			if len(r.Params) != 1 || r.Params[0] != "itemId" {
				t.Errorf("params for /api/items/:itemId = %v, want [itemId]", r.Params)
			}
		}
	}
}

func TestRoutesDedup(t *testing.T) {
	app := New()
	noop := func(req *Request, res *Response, next Next) {}
	// Two handlers on the same method+path must collapse to one route.
	app.Get("/x", noop, noop)
	app.Use("/x", noop) // middleware, must not appear
	routes := app.Routes()
	if len(routes) != 1 || routes[0].Method != "GET" || routes[0].Path != "/x" {
		t.Fatalf("dedup failed: %v", routes)
	}
}

func TestOpenAPIPathTemplating(t *testing.T) {
	cases := map[string]string{
		"/users/:id": "/users/{id}",
		"/a/:x/b/:y": "/a/{x}/b/{y}",
		"/files/*":   "/files/{wildcard}",
		"/opt/:id?":  "/opt/{id}",
		"/static":    "/static",
	}
	for in, want := range cases {
		if got := (RouteInfo{Path: in}).OpenAPIPath(); got != want {
			t.Errorf("OpenAPIPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOpenAPIDocument(t *testing.T) {
	app := sampleApp()
	app.Docs(DocsOptions{Title: "Sample", Version: "2.0.0", Servers: []string{"https://api.example.com"}})
	doc := app.OpenAPI()

	if doc.OpenAPI != "3.1.0" {
		t.Errorf("openapi version = %q", doc.OpenAPI)
	}
	if doc.Info.Title != "Sample" || doc.Info.Version != "2.0.0" {
		t.Errorf("info = %+v", doc.Info)
	}
	if len(doc.Servers) != 1 || doc.Servers[0].URL != "https://api.example.com" {
		t.Errorf("servers = %+v", doc.Servers)
	}

	// Path templated and methods present.
	pi, ok := doc.Paths["/users/{id}"]
	if !ok {
		t.Fatalf("missing /users/{id}; paths=%v", keysOf(doc.Paths))
	}
	get, ok := pi["get"]
	if !ok {
		t.Fatalf("missing GET on /users/{id}")
	}
	if get.Summary != "Fetch a user" {
		t.Errorf("summary = %q", get.Summary)
	}
	if len(get.Parameters) != 1 || get.Parameters[0].Name != "id" || get.Parameters[0].In != "path" || !get.Parameters[0].Required {
		t.Errorf("path param wrong: %+v", get.Parameters)
	}
	if _, ok := get.Responses["404"]; !ok {
		t.Errorf("missing 404 response: %v", get.Responses)
	}
	if get.OperationID == "" {
		t.Errorf("operationId not defaulted")
	}

	// POST /users has a request body and a query parameter.
	post := doc.Paths["/users"]["post"]
	if post.RequestBody == nil || !post.RequestBody.Required {
		t.Errorf("request body missing/optional: %+v", post.RequestBody)
	}
	if _, ok := post.RequestBody.Content["application/json"]; !ok {
		t.Errorf("request body content type: %+v", post.RequestBody.Content)
	}
	var hasQuery bool
	for _, p := range post.Parameters {
		if p.Name == "dryRun" && p.In == "query" {
			hasQuery = true
		}
	}
	if !hasQuery {
		t.Errorf("missing dryRun query param: %+v", post.Parameters)
	}

	// Routes without a description still get a default 200.
	health := doc.Paths["/health"]["get"]
	if _, ok := health.Responses["200"]; !ok {
		t.Errorf("default 200 missing for /health: %v", health.Responses)
	}

	// Tags aggregated.
	var hasUsersTag bool
	for _, tg := range doc.Tags {
		if tg.Name == "users" {
			hasUsersTag = true
		}
	}
	if !hasUsersTag {
		t.Errorf("users tag not aggregated: %v", doc.Tags)
	}

	// Whole doc must marshal to JSON.
	if _, err := json.Marshal(doc); err != nil {
		t.Fatalf("marshal openapi: %v", err)
	}
}

func TestDefaultOperationID(t *testing.T) {
	cases := map[[2]string]string{
		{"GET", "/users"}:      "getUsers",
		{"GET", "/users/:id"}:  "getUsersById",
		{"POST", "/api/items"}: "postApiItems",
		{"GET", "/a/:x/b"}:     "getAByXB",
	}
	for k, want := range cases {
		if got := defaultOperationID(k[0], k[1]); got != want {
			t.Errorf("defaultOperationID(%q,%q) = %q, want %q", k[0], k[1], got, want)
		}
	}
}

func TestServeEndpoints(t *testing.T) {
	app := sampleApp()
	app.Docs(DocsOptions{Title: "Served"})

	// OpenAPI JSON.
	res := do(app, "GET", "/openapi.json", "")
	if res.Code != 200 {
		t.Fatalf("openapi.json status %d", res.Code)
	}
	if ct := res.Header().Get("Content-Type"); !strings.Contains(ct, "json") {
		t.Errorf("openapi.json content-type = %q", ct)
	}
	var parsed map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("openapi.json not valid JSON: %v", err)
	}
	if parsed["openapi"] != "3.1.0" {
		t.Errorf("served openapi version = %v", parsed["openapi"])
	}

	// Swagger UI HTML.
	ui := do(app, "GET", "/docs", "")
	if ui.Code != 200 || !strings.Contains(ui.Body.String(), "swagger-ui") {
		t.Errorf("swagger UI not served: code=%d", ui.Code)
	}
	if ct := ui.Header().Get("Content-Type"); !strings.Contains(ct, "html") {
		t.Errorf("swagger UI content-type = %q", ct)
	}

	// ReDoc HTML.
	rd := do(app, "GET", "/redoc", "")
	if rd.Code != 200 || !strings.Contains(rd.Body.String(), "redoc") {
		t.Errorf("redoc not served: code=%d", rd.Code)
	}

	// YAML.
	y := do(app, "GET", "/openapi.yaml", "")
	if y.Code != 200 || !strings.Contains(y.Body.String(), "openapi:") {
		t.Errorf("openapi.yaml not served: %q", y.Body.String()[:min(80, y.Body.Len())])
	}

	// AsyncAPI.
	aa := do(app, "GET", "/asyncapi.json", "")
	if aa.Code != 200 {
		t.Fatalf("asyncapi status %d", aa.Code)
	}
	var async map[string]any
	if err := json.Unmarshal(aa.Body.Bytes(), &async); err != nil {
		t.Fatalf("asyncapi not JSON: %v", err)
	}
	if async["asyncapi"] != "2.6.0" {
		t.Errorf("asyncapi version = %v", async["asyncapi"])
	}

	// Postman.
	pm := do(app, "GET", "/postman.json", "")
	if pm.Code != 200 {
		t.Fatalf("postman status %d", pm.Code)
	}
	var col map[string]any
	if err := json.Unmarshal(pm.Body.Bytes(), &col); err != nil {
		t.Fatalf("postman not JSON: %v", err)
	}
	if _, ok := col["item"]; !ok {
		t.Errorf("postman missing item array")
	}
}

func TestAsyncAPIDocument(t *testing.T) {
	app := sampleApp()
	doc := app.AsyncAPI()
	if doc.AsyncAPI != "2.6.0" {
		t.Errorf("asyncapi version %q", doc.AsyncAPI)
	}
	ch, ok := doc.Channels["chat.message"]
	if !ok {
		t.Fatalf("missing channel; have %v", app.ChannelNames())
	}
	if ch.Subscribe == nil || ch.Subscribe.Message.Name != "messageReceived" {
		t.Errorf("subscribe message wrong: %+v", ch.Subscribe)
	}
	if ch.Publish == nil || ch.Publish.Message.ContentType != "application/json" {
		t.Errorf("publish message default content type wrong: %+v", ch.Publish)
	}
	if ch.Publish == nil || len(ch.Publish.Message.Examples) != 1 {
		t.Errorf("publish example not carried: %+v", ch.Publish)
	}
	if _, err := json.Marshal(doc); err != nil {
		t.Fatalf("marshal asyncapi: %v", err)
	}
}

func TestPostmanCollection(t *testing.T) {
	app := sampleApp()
	col := app.PostmanCollection()
	if len(col.Item) != len(app.Routes()) {
		t.Errorf("postman items = %d, routes = %d", len(col.Item), len(app.Routes()))
	}
	var found bool
	for _, it := range col.Item {
		if it.Request.Method == "GET" && it.Request.URL.Raw == "{{baseUrl}}/users/:id" {
			found = true
			if len(it.Request.URL.Path) != 2 || it.Request.URL.Path[1] != ":id" {
				t.Errorf("postman path segments wrong: %v", it.Request.URL.Path)
			}
		}
	}
	if !found {
		t.Errorf("postman missing GET /users/:id")
	}
}

func TestEnrichHook(t *testing.T) {
	app := sampleApp()
	app.Docs(DocsOptions{
		Enrich: func(route RouteInfo, op *Operation) {
			if route.Method == "GET" {
				op.Description = "enriched"
			}
		},
	})
	doc := app.OpenAPI()
	if doc.Paths["/health"]["get"].Description != "enriched" {
		t.Errorf("enrich hook not applied")
	}
}

func TestYAMLEncoder(t *testing.T) {
	in := map[string]any{
		"openapi": "3.1.0",
		"info":    map[string]any{"title": "T", "version": "1.0.0"},
		"list":    []any{"a", "b"},
		"num":     float64(200),
		"flag":    true,
		"empty":   map[string]any{},
	}
	y := docsToYAML(in)
	for _, want := range []string{"openapi: 3.1.0", "info:", "title: T", "num: 200", "flag: true", "empty: {}", "- a"} {
		if !strings.Contains(y, want) {
			t.Errorf("yaml missing %q in:\n%s", want, y)
		}
	}
}

func TestDisableEndpoint(t *testing.T) {
	app := sampleApp()
	app.Docs(DocsOptions{RedocPath: "-", UIPath: "/docs"})
	if rd := do(app, "GET", "/redoc", ""); rd.Code == 200 {
		t.Errorf("redoc should be disabled, got 200")
	}
	if ui := do(app, "GET", "/docs", ""); ui.Code != 200 {
		t.Errorf("docs should be enabled, got %d", ui.Code)
	}
}

func keysOf(m map[string]map[string]*Operation) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
