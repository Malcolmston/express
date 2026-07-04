// Package mimetypes provides MIME type lookups by filename or extension and
// reverse lookups from MIME type to extension. It is a port of the npm
// mime-types package using only the Go standard library.
package mimetypes

import (
	"path"
	"strings"
)

// typeByExt maps a lower-case file extension (without dot) to a MIME type.
var typeByExt = map[string]string{
	"html":     "text/html",
	"htm":      "text/html",
	"shtml":    "text/html",
	"css":      "text/css",
	"js":       "application/javascript",
	"mjs":      "application/javascript",
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

// normalizeExt extracts the lower-case extension (no dot) from a path,
// filename, bare extension, or ".ext".
func normalizeExt(pathOrExt string) string {
	s := strings.ToLower(strings.TrimSpace(pathOrExt))
	if s == "" {
		return ""
	}
	// If it contains a path separator or a dot, treat it as a path/filename.
	if strings.ContainsAny(s, "/\\.") {
		ext := path.Ext(strings.ReplaceAll(s, "\\", "/"))
		if ext != "" {
			return strings.TrimPrefix(ext, ".")
		}
		// No dot-extension found; fall through treating whole thing as ext
		// only if it has no separators.
		if strings.ContainsAny(s, "/\\") {
			return ""
		}
	}
	return s
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
// text/* types and JSON, XML and JavaScript types use "utf-8".
func Charset(mimeType string) (string, bool) {
	t := strings.ToLower(strings.TrimSpace(mimeType))
	if i := strings.IndexByte(t, ';'); i >= 0 {
		t = strings.TrimSpace(t[:i])
	}
	if strings.HasPrefix(t, "text/") {
		return "utf-8", true
	}
	switch t {
	case "application/json", "application/javascript", "text/javascript",
		"application/xml", "application/ld+json", "application/manifest+json":
		return "utf-8", true
	}
	if strings.HasSuffix(t, "+json") || strings.HasSuffix(t, "+xml") {
		return "utf-8", true
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
		return mime + "; charset=" + cs, true
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
