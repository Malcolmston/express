package express

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//go:embed html/swagger-ui.html
var swaggerUITemplate string

//go:embed html/redoc.html
var redocTemplate string

// swaggerUIHTML returns a self-contained HTML page that renders Swagger UI for
// the OpenAPI document at specURL, loading the Swagger UI assets from assetBase.
func swaggerUIHTML(title, specURL, assetBase string) string {
	r := strings.NewReplacer(
		"__TITLE__", htmlEscape(title),
		"__ASSET_BASE__", strings.TrimRight(assetBase, "/"),
		"__SPEC_URL_JSON__", jsString(specURL),
	)
	return r.Replace(swaggerUITemplate)
}

// redocHTML returns a self-contained HTML page that renders ReDoc for the
// OpenAPI document at specURL, loading the ReDoc bundle from assetBase.
func redocHTML(title, specURL, assetBase string) string {
	r := strings.NewReplacer(
		"__TITLE__", htmlEscape(title),
		"__ASSET_BASE__", strings.TrimRight(assetBase, "/"),
		"__SPEC_URL_ATTR__", htmlAttr(specURL),
	)
	return r.Replace(redocTemplate)
}

// ---------------------------------------------------------------------------
// Postman collection export
// ---------------------------------------------------------------------------

// PostmanCollection is a minimal Postman Collection v2.1 document generated from
// the application's routes. It marshals to a file importable by Postman/Insomnia.
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanInfo is the info block of a Postman collection.
type PostmanInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

// PostmanItem is a single request entry in a Postman collection.
type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request"`
}

// PostmanRequest describes the HTTP request of a Postman item.
type PostmanRequest struct {
	Method string     `json:"method"`
	Header []any      `json:"header"`
	URL    PostmanURL `json:"url"`
}

// PostmanURL is the structured URL of a Postman request.
type PostmanURL struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

// PostmanCollection builds a Postman Collection v2.1 from the registered routes.
// Path parameters are rendered in Postman's ":param" style. The {{baseUrl}}
// variable stands in for the server so the collection is host-agnostic.
func (app *Application) PostmanCollection() *PostmanCollection {
	reg := app.docsReg()
	opts := reg.opts
	if opts.Title == "" {
		opts.withDefaults(app)
	}
	col := &PostmanCollection{
		Info: PostmanInfo{
			Name:   opts.Title,
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
	}
	for _, ri := range app.Routes() {
		segs := splitPath(ri.Path)
		name := ri.Method + " " + ri.Path
		if rd, ok := reg.describes[ri.Method+" "+ri.Path]; ok && rd.Summary != "" {
			name = rd.Summary
		}
		col.Item = append(col.Item, PostmanItem{
			Name: name,
			Request: PostmanRequest{
				Method: ri.Method,
				Header: []any{},
				URL: PostmanURL{
					Raw:  "{{baseUrl}}" + ri.Path,
					Host: []string{"{{baseUrl}}"},
					Path: segs,
				},
			},
		})
	}
	return col
}

func splitPath(path string) []string {
	var segs []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	return segs
}

// ---------------------------------------------------------------------------
// YAML encoding (for the OpenAPI YAML endpoint)
// ---------------------------------------------------------------------------

// docsToYAML renders any JSON-marshalable value as YAML. It round-trips through
// encoding/json so struct tags and omitempty are honoured, then emits a YAML
// document from the generic tree. Map keys are sorted for deterministic output.
func docsToYAML(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return ""
	}
	var sb strings.Builder
	encodeYAML(&sb, tree, 0, false)
	s := sb.String()
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

func encodeYAML(sb *strings.Builder, v any, indent int, inline bool) {
	pad := strings.Repeat("  ", indent)
	switch val := v.(type) {
	case map[string]any:
		if len(val) == 0 {
			sb.WriteString("{}\n")
			return
		}
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		first := true
		for _, k := range keys {
			if inline && first {
				// key printed on the parent's dash line
			} else {
				sb.WriteString(pad)
			}
			first = false
			sb.WriteString(yamlKey(k))
			sb.WriteByte(':')
			child := val[k]
			if isScalar(child) {
				sb.WriteByte(' ')
				sb.WriteString(yamlScalar(child))
				sb.WriteByte('\n')
			} else if isEmptyContainer(child) {
				sb.WriteByte(' ')
				encodeYAML(sb, child, indent+1, false)
			} else {
				sb.WriteByte('\n')
				encodeYAML(sb, child, indent+1, false)
			}
		}
	case []any:
		if len(val) == 0 {
			sb.WriteString("[]\n")
			return
		}
		for _, item := range val {
			sb.WriteString(pad)
			sb.WriteString("- ")
			if isScalar(item) {
				sb.WriteString(yamlScalar(item))
				sb.WriteByte('\n')
			} else if isEmptyContainer(item) {
				encodeYAML(sb, item, indent+1, false)
			} else {
				// Inline the first key of a map on the dash line.
				encodeYAML(sb, item, indent+1, true)
			}
		}
	default:
		sb.WriteString(yamlScalar(v))
		sb.WriteByte('\n')
	}
}

func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

func isEmptyContainer(v any) bool {
	switch c := v.(type) {
	case map[string]any:
		return len(c) == 0
	case []any:
		return len(c) == 0
	default:
		return false
	}
}

func yamlKey(k string) string {
	if k == "" || yamlNeedsQuote(k) {
		return strconv.Quote(k)
	}
	return k
}

func yamlScalar(v any) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'g', -1, 64)
	case string:
		if val == "" || yamlNeedsQuote(val) {
			return strconv.Quote(val)
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// yamlNeedsQuote reports whether a plain scalar would be ambiguous in YAML and
// therefore must be quoted.
func yamlNeedsQuote(s string) bool {
	if s == "" {
		return true
	}
	switch s {
	case "null", "Null", "NULL", "true", "True", "TRUE", "false", "False", "FALSE",
		"yes", "no", "on", "off", "~":
		return true
	}
	if strings.ContainsAny(s, ":#{}[],&*!|>'\"%@`\n\t") {
		return true
	}
	if strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
		return true
	}
	switch s[0] {
	case '-', '?', '&', '*', '!', '%', '@', '`', '>', '|':
		return true
	}
	// Looks numeric -> quote to keep it a string.
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func htmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func htmlAttr(s string) string {
	return strconv.Quote(strings.ReplaceAll(s, `"`, "&quot;"))
}

func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
