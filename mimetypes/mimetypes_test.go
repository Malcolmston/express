package mimetypes

import "testing"

func TestLookup(t *testing.T) {
	cases := map[string]string{
		"file.html":      "text/html",
		"index.htm":      "text/html",
		"style.css":      "text/css",
		"app.js":         "application/javascript",
		"mod.mjs":        "application/javascript",
		"data.json":      "application/json",
		"doc.xml":        "application/xml",
		"notes.txt":      "text/plain",
		"rows.csv":       "text/csv",
		"readme.md":      "text/markdown",
		"a.markdown":     "text/markdown",
		"img.png":        "image/png",
		"photo.jpg":      "image/jpeg",
		"photo.jpeg":     "image/jpeg",
		"anim.gif":       "image/gif",
		"vec.svg":        "image/svg+xml",
		"pic.webp":       "image/webp",
		"fav.ico":        "image/x-icon",
		"doc.pdf":        "application/pdf",
		"a.zip":          "application/zip",
		"a.gz":           "application/gzip",
		"a.tar":          "application/x-tar",
		"v.mp4":          "video/mp4",
		"s.mp3":          "audio/mpeg",
		"s.wav":          "audio/wav",
		"s.ogg":          "audio/ogg",
		"v.webm":         "video/webm",
		"f.woff":         "font/woff",
		"f.woff2":        "font/woff2",
		"f.ttf":          "font/ttf",
		"f.otf":          "font/otf",
		"f.eot":          "application/vnd.ms-fontobject",
		"m.wasm":         "application/wasm",
		"d.bin":          "application/octet-stream",
		"/path/to/x.png": "image/png",
	}
	for in, want := range cases {
		got, ok := Lookup(in)
		if !ok || got != want {
			t.Errorf("Lookup(%q) = %q,%v; want %q", in, got, ok, want)
		}
	}
}

func TestLookupBareExt(t *testing.T) {
	got, ok := Lookup("json")
	if !ok || got != "application/json" {
		t.Errorf("Lookup(json) = %q,%v", got, ok)
	}
	got, ok = Lookup(".png")
	if !ok || got != "image/png" {
		t.Errorf("Lookup(.png) = %q,%v", got, ok)
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("file.unknownext"); ok {
		t.Error("expected unknown ext to fail")
	}
	if _, ok := Lookup(""); ok {
		t.Error("expected empty to fail")
	}
}

func TestContentType(t *testing.T) {
	cases := map[string]string{
		"txt":              "text/plain; charset=utf-8",
		"html":             "text/html; charset=utf-8",
		"json":             "application/json; charset=utf-8",
		"text/html":        "text/html; charset=utf-8",
		"application/json": "application/json; charset=utf-8",
		"png":              "image/png",
		"image/png":        "image/png",
		"application/xml":  "application/xml; charset=utf-8",
	}
	for in, want := range cases {
		got, ok := ContentType(in)
		if !ok || got != want {
			t.Errorf("ContentType(%q) = %q,%v; want %q", in, got, ok, want)
		}
	}
}

func TestContentTypePreservesCharset(t *testing.T) {
	got, ok := ContentType("text/html; charset=iso-8859-1")
	if !ok || got != "text/html; charset=iso-8859-1" {
		t.Errorf("ContentType = %q,%v", got, ok)
	}
}

func TestExtension(t *testing.T) {
	cases := map[string]string{
		"text/html":             "html",
		"application/json":      "json",
		"image/jpeg":            "jpg",
		"image/png":             "png",
		"application/pdf":       "pdf",
		"text/plain":            "txt",
		"application/json; q=1": "json",
	}
	for in, want := range cases {
		got, ok := Extension(in)
		if !ok || got != want {
			t.Errorf("Extension(%q) = %q,%v; want %q", in, got, ok, want)
		}
	}
	if _, ok := Extension("application/x-not-real"); ok {
		t.Error("expected unknown mime to fail")
	}
}

func TestCharset(t *testing.T) {
	yes := map[string]string{
		"text/plain":               "utf-8",
		"text/html":                "utf-8",
		"application/json":         "utf-8",
		"application/xml":          "utf-8",
		"application/javascript":   "utf-8",
		"application/vnd.api+json": "utf-8",
		"image/svg+xml":            "utf-8",
	}
	for in, want := range yes {
		got, ok := Charset(in)
		if !ok || got != want {
			t.Errorf("Charset(%q) = %q,%v; want %q", in, got, ok, want)
		}
	}
	if _, ok := Charset("image/png"); ok {
		t.Error("image/png should have no charset")
	}
	if _, ok := Charset("application/octet-stream"); ok {
		t.Error("octet-stream should have no charset")
	}
}
