// Package mimetypes provides MIME type lookups by filename or extension and
// reverse lookups from MIME type to extension. It is a port of the npm
// mime-types package, the module Express uses behind res.type/res.contentType
// and that content-negotiation middleware relies on, implemented using only
// the Go standard library. It answers the everyday questions "what
// Content-Type should I send for this file?" and, going the other way, "what
// file extension does this MIME type correspond to?".
//
// The package is useful anywhere you serve files, set response headers, or map
// uploads to a canonical type. Rather than depending on the operating system's
// mime.types database (whose contents vary by machine), it ships a curated,
// self-contained table covering the common web, image, video, audio, font,
// document and archive types, so results are identical across platforms. All
// functions are pure lookups over that table, do no I/O, and are safe for
// concurrent use.
//
// The forward direction is handled by Lookup, which accepts a full path, a bare
// filename, a ".ext" or a bare extension, extracts the lower-case extension and
// returns the mapped MIME type. The returned type never carries a charset.
// ContentType builds on Lookup (or accepts a MIME type directly when the input
// already contains a "/") and appends "; charset=utf-8" for textual types,
// giving you a value suitable for a Content-Type header in one call. Charset
// reports the charset that applies to a given MIME type, treating every text/*
// type plus JSON, XML, JavaScript and any "+json"/"+xml" structured-suffix
// type as UTF-8.
//
// The reverse direction is Extension, which takes a MIME type (optionally with
// parameters, which are ignored) and returns its canonical extension without a
// leading dot. Where several extensions share a type, the most common one is
// chosen as canonical: image/jpeg maps back to "jpg", text/plain to "txt", and
// application/javascript to "js". Because the table also records alias types
// such as text/javascript and image/vnd.microsoft.icon, those resolve to the
// same canonical extensions as their primaries.
//
// Edge cases follow a consistent rule: every function returns a boolean second
// value that is false when nothing is known. An empty string, an extension that
// is not in the table, or a MIME type with no known extension yields ("",
// false) rather than a guess, so callers can distinguish "unknown" from a real
// answer. Compared with the Node original the API surface is adapted to Go —
// results are (value, ok) pairs instead of a value-or-false union, and the
// bundled type table is a fixed curated subset rather than the full generated
// mime-db — but the lookup semantics, charset rules and canonical extension
// choices mirror mime-types.
package mimetypes

import (
	"strings"
)

// typeByExt maps a lower-case file extension (without dot) to a MIME type.
var typeByExt = map[string]string{
	"html":     "text/html",
	"htm":      "text/html",
	"shtml":    "text/html",
	"css":      "text/css",
	"js":       "text/javascript",
	"mjs":      "text/javascript",
	"json":     "application/json",
	"map":      "application/json",
	"xml":      "application/xml",
	"txt":      "text/plain",
	"text":     "text/plain",
	"conf":     "text/plain",
	"log":      "text/plain",
	"csv":      "text/csv",
	"md":       "text/markdown",
	"markdown": "text/markdown",
	"png":      "image/png",
	"jpg":      "image/jpeg",
	"jpeg":     "image/jpeg",
	"jpe":      "image/jpeg",
	"gif":      "image/gif",
	"svg":      "image/svg+xml",
	"svgz":     "image/svg+xml",
	"webp":     "image/webp",
	"ico":      "image/x-icon",
	"bmp":      "image/bmp",
	"tiff":     "image/tiff",
	"tif":      "image/tiff",
	"pdf":      "application/pdf",
	"zip":      "application/zip",
	"gz":       "application/gzip",
	"tgz":      "application/gzip",
	"tar":      "application/x-tar",
	"rar":      "application/vnd.rar",
	"7z":       "application/x-7z-compressed",
	"bz2":      "application/x-bzip2",
	"mp4":      "video/mp4",
	"m4v":      "video/mp4",
	"mpeg":     "video/mpeg",
	"mpg":      "video/mpeg",
	"mov":      "video/quicktime",
	"avi":      "video/x-msvideo",
	"webm":     "video/webm",
	"mp3":      "audio/mpeg",
	"m4a":      "audio/mp4",
	"wav":      "audio/wav",
	"ogg":      "audio/ogg",
	"oga":      "audio/ogg",
	"opus":     "audio/opus",
	"flac":     "audio/flac",
	"woff":     "font/woff",
	"woff2":    "font/woff2",
	"ttf":      "font/ttf",
	"otf":      "font/otf",
	"eot":      "application/vnd.ms-fontobject",
	"wasm":     "application/wasm",
	"bin":      "application/octet-stream",
	"exe":      "application/octet-stream",
	"dll":      "application/octet-stream",
	"doc":      "application/msword",
	"xls":      "application/vnd.ms-excel",
	"ppt":      "application/vnd.ms-powerpoint",
	"rtf":      "application/rtf",
	"yaml":     "text/yaml",
	"yml":      "text/yaml",
	"ics":      "text/calendar",
}

// extByType maps a MIME type to its canonical extension (without dot). Where
// multiple extensions share a type the canonical, most common one is chosen.
var extByType = map[string]string{
	"text/html":                     "html",
	"text/css":                      "css",
	"application/javascript":        "js",
	"text/javascript":               "js",
	"application/json":              "json",
	"application/xml":               "xml",
	"text/xml":                      "xml",
	"text/plain":                    "txt",
	"text/csv":                      "csv",
	"text/markdown":                 "md",
	"image/png":                     "png",
	"image/jpeg":                    "jpg",
	"image/gif":                     "gif",
	"image/svg+xml":                 "svg",
	"image/webp":                    "webp",
	"image/x-icon":                  "ico",
	"image/vnd.microsoft.icon":      "ico",
	"image/bmp":                     "bmp",
	"image/tiff":                    "tiff",
	"application/pdf":               "pdf",
	"application/zip":               "zip",
	"application/gzip":              "gz",
	"application/x-tar":             "tar",
	"video/mp4":                     "mp4",
	"video/mpeg":                    "mpeg",
	"video/quicktime":               "mov",
	"video/webm":                    "webm",
	"audio/mpeg":                    "mp3",
	"audio/mp4":                     "m4a",
	"audio/wav":                     "wav",
	"audio/ogg":                     "ogg",
	"audio/opus":                    "opus",
	"audio/flac":                    "flac",
	"font/woff":                     "woff",
	"font/woff2":                    "woff2",
	"font/ttf":                      "ttf",
	"font/otf":                      "otf",
	"application/vnd.ms-fontobject": "eot",
	"application/wasm":              "wasm",
	"application/octet-stream":      "bin",
	"text/yaml":                     "yaml",
	"text/calendar":                 "ics",
	"application/msword":            "doc",
	"application/vnd.ms-excel":      "xls",
	"application/vnd.ms-powerpoint": "ppt",
	"application/rtf":               "rtf",
}

// nodeExtname returns the extension (including the leading dot) of the final
// path segment, mirroring Node.js path.extname: a leading-dot "dotfile" whose
// only dot is the first character of the base name has no extension. Unlike
// the standard library path.Ext, which reports ".json" for a base name like
// ".json", this treats such dotfiles as extension-less. Only "/" is treated as
// a separator, matching Node's behavior on POSIX (upstream mime-types relies on
// this, so "C:\\path\\to\\page.html" resolves via the final dot).
func nodeExtname(p string) string {
	base := p
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		base = p[i+1:]
	}
	dot := strings.LastIndexByte(base, '.')
	// dot == -1: no dot; dot == 0: leading-dot dotfile. Both have no extension.
	if dot <= 0 {
		return ""
	}
	return base[dot:]
}

// normalizeExt extracts the lower-case extension (no dot) from a path,
// filename, bare extension, or ".ext". It follows upstream mime-types, which
// resolves the extension via extname("x." + input): prefixing "x." makes bare
// extensions and ".ext" inputs resolve like a file name, while path-embedded
// dotfiles such as "/path/to/.json" correctly yield no extension.
func normalizeExt(pathOrExt string) string {
	if pathOrExt == "" {
		return ""
	}
	ext := nodeExtname("x." + pathOrExt)
	if ext == "" {
		return ""
	}
	return strings.ToLower(ext[1:])
}

// Lookup returns the MIME type for a filename, path, or extension. The
// returned type never includes a charset. The boolean is false when no
// mapping is known.
func Lookup(pathOrExt string) (string, bool) {
	ext := normalizeExt(pathOrExt)
	if ext == "" {
		return "", false
	}
	t, ok := typeByExt[ext]
	return t, ok
}

// Charset returns the charset for a MIME type, or false if none applies.
// text/* types and JSON, XML and JavaScript types use "UTF-8". The value is
// returned in the canonical upstream casing ("UTF-8"); ContentType lower-cases
// it when building a header value.
func Charset(mimeType string) (string, bool) {
	t := strings.ToLower(strings.TrimSpace(mimeType))
	if i := strings.IndexByte(t, ';'); i >= 0 {
		t = strings.TrimSpace(t[:i])
	}
	if strings.HasPrefix(t, "text/") {
		return "UTF-8", true
	}
	switch t {
	case "application/json", "application/javascript", "text/javascript",
		"application/xml", "application/ld+json", "application/manifest+json":
		return "UTF-8", true
	}
	if strings.HasSuffix(t, "+json") || strings.HasSuffix(t, "+xml") {
		return "UTF-8", true
	}
	return "", false
}

// ContentType returns a full Content-Type value for a MIME type or extension.
// A "; charset=utf-8" parameter is appended for textual types. If the input
// already contains a "/" it is treated as a MIME type; otherwise it is looked
// up as an extension.
func ContentType(typeOrExt string) (string, bool) {
	s := strings.TrimSpace(typeOrExt)
	if s == "" {
		return "", false
	}
	var mime string
	if strings.ContainsRune(s, '/') {
		mime = s
	} else {
		var ok bool
		mime, ok = Lookup(s)
		if !ok {
			return "", false
		}
	}
	if strings.Contains(mime, "charset") {
		return mime, true
	}
	if cs, ok := Charset(mime); ok {
		return mime + "; charset=" + strings.ToLower(cs), true
	}
	return mime, true
}

// Extension returns the canonical file extension (without a leading dot) for a
// MIME type, ignoring any parameters. The boolean is false when unknown.
func Extension(mimeType string) (string, bool) {
	t := strings.ToLower(strings.TrimSpace(mimeType))
	if i := strings.IndexByte(t, ';'); i >= 0 {
		t = strings.TrimSpace(t[:i])
	}
	ext, ok := extByType[t]
	return ext, ok
}
