package express

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// EngineFunc renders a template file at path with the given data and returns the
// rendered output. Register one with app.Engine to support a template language.
type EngineFunc func(path string, data any) (string, error)

// Engine registers a template engine for a file extension (e.g. ".html",
// ".pug"). The extension should include the leading dot.
func (app *Application) Engine(ext string, fn EngineFunc) *Application {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	app.viewEngines[ext] = fn
	return app
}

// renderView resolves a view name to a file under the "views" directory and
// renders it with the engine registered for its extension. When the name has no
// extension, the "view engine" setting supplies the default.
func (app *Application) renderView(name string, data any) (string, error) {
	ext := filepath.Ext(name)
	if ext == "" {
		ve, _ := app.settings["view engine"].(string)
		if ve == "" {
			ve = "html"
		}
		ext = "." + strings.TrimPrefix(ve, ".")
		name += ext
	}
	dir, _ := app.settings["views"].(string)
	if dir == "" {
		dir = "views"
	}
	engine := app.viewEngines[ext]
	if engine == nil {
		return "", fmt.Errorf("express: no view engine registered for %q", ext)
	}
	return engine(filepath.Join(dir, name), data)
}

// Render renders a view and sends it as HTML. The view name is resolved against
// the app's "views" directory and "view engine" setting. Optional data is
// passed to the template; when omitted, res.Locals is used.
func (res *Response) Render(name string, data ...any) error {
	var d any = res.Locals
	if len(data) > 0 {
		d = data[0]
	}
	html, err := res.app.renderView(name, d)
	if err != nil {
		res.finalError(err)
		return err
	}
	res.Type("html").Send(html)
	return nil
}

// htmlTemplateEngine is the built-in engine backed by html/template.
func htmlTemplateEngine(path string, data any) (string, error) {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
