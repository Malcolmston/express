package express

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// EngineFunc renders a template file at path with the given data and returns the
// rendered output. Register one with app.Engine to support a template language.
type EngineFunc func(path string, data any) (string, error)

// viewCache memoises resolved view paths and compiled templates. Population is
// gated by the app's "view cache" setting; lookups are always safe to read.
type viewCache struct {
	mu       sync.RWMutex
	lookup   map[string]string             // view name -> resolved absolute path
	compiled map[string]*template.Template // path -> parsed template (built-in engine)
}

func newViewCache() *viewCache {
	return &viewCache{
		lookup:   make(map[string]string),
		compiled: make(map[string]*template.Template),
	}
}

// Engine registers a template engine for a file extension (e.g. ".html",
// ".pug"). The extension should include the leading dot.
func (app *Application) Engine(ext string, fn EngineFunc) *Application {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	app.viewEngines[ext] = fn
	return app
}

// viewCacheEnabled reports whether resolved views and compiled templates should
// be memoised, mirroring Express's "view cache" setting.
func (app *Application) viewCacheEnabled() bool {
	v, _ := app.settings["view cache"].(bool)
	return v
}

// viewRoots returns the configured "views" directories. The setting may be a
// single string or a []string; it defaults to "views".
func (app *Application) viewRoots() []string {
	switch v := app.settings["views"].(type) {
	case string:
		if v != "" {
			return []string{v}
		}
	case []string:
		if len(v) > 0 {
			return v
		}
	case []any:
		roots := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok && s != "" {
				roots = append(roots, s)
			}
		}
		if len(roots) > 0 {
			return roots
		}
	}
	return []string{"views"}
}

// defaultViewExt returns the extension implied by the "view engine" setting,
// including the leading dot (e.g. ".html").
func (app *Application) defaultViewExt() string {
	ve, _ := app.settings["view engine"].(string)
	if ve == "" {
		ve = "html"
	}
	return "." + strings.TrimPrefix(ve, ".")
}

// fileExists reports whether path names an existing regular (non-directory) file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// lookupView resolves a view name to an on-disk file, mirroring Express's View
// lookup: it applies the default extension, searches every "views" root, honours
// absolute paths, and falls back to an index file inside a matching directory.
// Results are cached when "view cache" is enabled.
func (app *Application) lookupView(name string) (string, error) {
	if app.viewCacheEnabled() {
		app.viewCache.mu.RLock()
		if p, ok := app.viewCache.lookup[name]; ok {
			app.viewCache.mu.RUnlock()
			return p, nil
		}
		app.viewCache.mu.RUnlock()
	}

	ext := filepath.Ext(name)
	if ext == "" {
		ext = app.defaultViewExt()
		name += ext
	}

	// Candidate paths, tried in order, for a given root directory.
	candidates := func(root string) []string {
		joined := filepath.Join(root, name)
		return []string{
			joined,
			// Express: a view that names a directory resolves to its index file.
			filepath.Join(strings.TrimSuffix(joined, ext), "index"+ext),
		}
	}

	var tried []string
	roots := app.viewRoots()
	// An absolute view path is used as-is (still honouring the index fallback).
	if filepath.IsAbs(name) {
		roots = []string{""}
	}
	for _, root := range roots {
		for _, c := range candidates(root) {
			if fileExists(c) {
				resolved := c
				if app.viewCacheEnabled() {
					app.viewCache.mu.Lock()
					app.viewCache.lookup[name] = resolved
					app.viewCache.mu.Unlock()
				}
				return resolved, nil
			}
			tried = append(tried, c)
		}
	}

	dirs := strings.Join(roots, ", ")
	if filepath.IsAbs(name) {
		dirs = filepath.Dir(name)
	}
	return "", fmt.Errorf("express: failed to lookup view %q in views directory %q (tried %s)",
		name, dirs, strings.Join(tried, ", "))
}

// Render resolves a view name against the app's "views" directories and
// "view engine" setting, renders it with the matching engine, and returns the
// output. It is the app-level equivalent of Express's app.render.
func (app *Application) Render(name string, data any) (string, error) {
	path, err := app.lookupView(name)
	if err != nil {
		return "", err
	}
	engine := app.viewEngines[filepath.Ext(path)]
	if engine == nil {
		return "", fmt.Errorf("express: no view engine registered for %q", filepath.Ext(path))
	}
	return engine(path, data)
}

// Render renders a view and sends it as HTML. The view name is resolved against
// the app's "views" directories and "view engine" setting. When data is a
// map[string]any it is layered over res.Locals (so per-render values win while
// app/request locals remain visible); any other value is passed through as-is.
// When data is omitted, res.Locals is used.
func (res *Response) Render(name string, data ...any) error {
	d := res.renderData(data)
	html, err := res.app.Render(name, d)
	if err != nil {
		res.finalError(err)
		return err
	}
	res.Type("html").Send(html)
	return nil
}

// renderData builds the value handed to the template: res.Locals by default, or
// res.Locals merged under a map argument (argument keys win).
func (res *Response) renderData(data []any) any {
	if len(data) == 0 || data[0] == nil {
		return res.Locals
	}
	m, ok := data[0].(map[string]any)
	if !ok {
		return data[0]
	}
	merged := make(map[string]any, len(res.Locals)+len(m))
	for k, v := range res.Locals {
		merged[k] = v
	}
	for k, v := range m {
		merged[k] = v
	}
	return merged
}

// htmlTemplateEngine is the built-in engine backed by html/template. It caches
// compiled templates per path when the "view cache" setting is enabled.
func (app *Application) htmlTemplateEngine(path string, data any) (string, error) {
	var tmpl *template.Template
	if app.viewCacheEnabled() {
		app.viewCache.mu.RLock()
		tmpl = app.viewCache.compiled[path]
		app.viewCache.mu.RUnlock()
	}
	if tmpl == nil {
		t, err := template.ParseFiles(path)
		if err != nil {
			return "", err
		}
		tmpl = t
		if app.viewCacheEnabled() {
			app.viewCache.mu.Lock()
			app.viewCache.compiled[path] = tmpl
			app.viewCache.mu.Unlock()
		}
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
