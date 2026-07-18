package express

import (
	"strings"
)

// OpenAPIDoc is the root of a generated OpenAPI 3.1 document. It marshals to the
// canonical OpenAPI JSON shape and can be served directly with res.JSON.
type OpenAPIDoc struct {
	OpenAPI string                           `json:"openapi"`
	Info    OpenAPIInfo                      `json:"info"`
	Servers []OpenAPIServer                  `json:"servers,omitempty"`
	Tags    []OpenAPITag                     `json:"tags,omitempty"`
	Paths   map[string]map[string]*Operation `json:"paths"`
}

// OpenAPIInfo is the info block of an OpenAPI document.
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// OpenAPIServer is a single entry of the OpenAPI servers block.
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPITag documents a tag used to group operations.
type OpenAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Operation is a single OpenAPI operation (one method on one path). It is
// exposed so the DocsOptions.Enrich hook can customise generated operations.
type Operation struct {
	Tags        []string                   `json:"tags,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Deprecated  bool                       `json:"deprecated,omitempty"`
	Parameters  []OpenAPIParam             `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
}

// OpenAPIParam is a single operation parameter.
type OpenAPIParam struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

// OpenAPIRequestBody documents an operation's request body.
type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}

// OpenAPIMediaType is a schema/example pair keyed by media type in content maps.
type OpenAPIMediaType struct {
	Schema  map[string]any `json:"schema,omitempty"`
	Example any            `json:"example,omitempty"`
}

// OpenAPIResponse documents a single response.
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

// OpenAPI builds an OpenAPI 3.1 document describing every route registered on
// the application. Routes contribute their method, templated path and path
// parameters automatically; any metadata attached with [Application.Describe]
// enriches the matching operation, and a DocsOptions.Enrich hook (if set) gets
// the final say. The returned value can be served directly with res.JSON.
func (app *Application) OpenAPI() *OpenAPIDoc {
	reg := app.docsReg()
	opts := reg.opts
	if opts.Title == "" { // Docs() may not have run yet; apply defaults.
		opts.withDefaults(app)
	}

	doc := &OpenAPIDoc{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:       opts.Title,
			Version:     opts.Version,
			Description: opts.Description,
		},
		Paths: map[string]map[string]*Operation{},
	}
	for _, s := range opts.Servers {
		doc.Servers = append(doc.Servers, OpenAPIServer{URL: s})
	}

	tagSet := map[string]bool{}
	for _, ri := range app.Routes() {
		key := ri.Method + " " + ri.Path
		rd := reg.describes[key]

		op := &Operation{
			Summary:     rd.Summary,
			Description: rd.Description,
			Tags:        rd.Tags,
			OperationID: rd.OperationID,
			Deprecated:  rd.Deprecated,
			Responses:   map[string]OpenAPIResponse{},
		}
		if op.OperationID == "" {
			op.OperationID = defaultOperationID(ri.Method, ri.Path)
		}
		for _, t := range rd.Tags {
			tagSet[t] = true
		}

		// Path parameters (auto), honouring any user override of the same name.
		overrides := map[string]ParamDoc{}
		for _, p := range rd.Parameters {
			if p.In == "path" {
				overrides[p.Name] = p
			}
		}
		for _, name := range ri.Params {
			if ov, ok := overrides[name]; ok {
				op.Parameters = append(op.Parameters, toOpenAPIParam(ov, true))
				continue
			}
			op.Parameters = append(op.Parameters, OpenAPIParam{
				Name:     name,
				In:       "path",
				Required: true,
				Schema:   map[string]any{"type": "string"},
			})
		}
		// Non-path parameters (query, header, cookie) from the description.
		for _, p := range rd.Parameters {
			if p.In == "path" {
				continue
			}
			op.Parameters = append(op.Parameters, toOpenAPIParam(p, p.Required))
		}

		// Request body.
		if rd.RequestBody != nil {
			ct := rd.RequestBody.ContentType
			if ct == "" {
				ct = "application/json"
			}
			op.RequestBody = &OpenAPIRequestBody{
				Description: rd.RequestBody.Description,
				Required:    rd.RequestBody.Required,
				Content: map[string]OpenAPIMediaType{
					ct: {Schema: rd.RequestBody.Schema, Example: rd.RequestBody.Example},
				},
			}
		}

		// Responses.
		if len(rd.Responses) == 0 {
			op.Responses["200"] = OpenAPIResponse{Description: "Successful response"}
		} else {
			for status, r := range rd.Responses {
				desc := r.Description
				if desc == "" {
					desc = "Response"
				}
				resp := OpenAPIResponse{Description: desc}
				if r.Schema != nil || r.Example != nil {
					ct := r.ContentType
					if ct == "" {
						ct = "application/json"
					}
					resp.Content = map[string]OpenAPIMediaType{
						ct: {Schema: r.Schema, Example: r.Example},
					}
				}
				op.Responses[status] = resp
			}
		}

		if opts.Enrich != nil {
			opts.Enrich(ri, op)
		}

		tmpl := ri.OpenAPIPath()
		if doc.Paths[tmpl] == nil {
			doc.Paths[tmpl] = map[string]*Operation{}
		}
		doc.Paths[tmpl][strings.ToLower(ri.Method)] = op
	}

	for name := range tagSet {
		doc.Tags = append(doc.Tags, OpenAPITag{Name: name})
	}
	sortTags(doc.Tags)
	return doc
}

// OpenAPIYAML returns the OpenAPI document rendered as YAML.
func (app *Application) OpenAPIYAML() string {
	return docsToYAML(app.OpenAPI())
}

func toOpenAPIParam(p ParamDoc, required bool) OpenAPIParam {
	schema := p.Schema
	if schema == nil {
		schema = map[string]any{"type": "string"}
	}
	in := p.In
	if in == "" {
		in = "query"
	}
	return OpenAPIParam{
		Name:        p.Name,
		In:          in,
		Description: p.Description,
		Required:    required,
		Schema:      schema,
	}
}

// defaultOperationID derives a camelCase identifier from a method and path, e.g.
// ("GET", "/users/:id") -> "getUsersById".
func defaultOperationID(method, path string) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(method))
	upperNext := false
	byParam := false
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch {
		case c == ':':
			// ":id" -> "ById"
			b.WriteString("By")
			upperNext = true
			byParam = true
		case c == '/' || c == '-' || c == '_' || c == '.':
			upperNext = true
		case c == '*':
			b.WriteString("Wildcard")
		case isNameChar(c):
			ch := c
			if upperNext {
				ch = upperByte(ch)
				upperNext = false
			}
			b.WriteByte(ch)
		}
		_ = byParam
	}
	return b.String()
}

func upperByte(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func sortTags(tags []OpenAPITag) {
	for i := 1; i < len(tags); i++ {
		for j := i; j > 0 && tags[j-1].Name > tags[j].Name; j-- {
			tags[j-1], tags[j] = tags[j], tags[j-1]
		}
	}
}
