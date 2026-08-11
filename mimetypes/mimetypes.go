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
// document and archive types, so results are identical across platforms. The
// table is a subset of mime-db, but every entry in it is taken *from* mime-db
// rather than chosen by hand, so a type the table knows resolves exactly as
// upstream resolves it. All functions are pure lookups over that table, do no
// I/O, and are safe for concurrent use.
//
// The forward direction is handled by Lookup, which accepts a full path, a bare
// filename, a ".ext" or a bare extension, extracts the lower-case extension and
// returns the mapped MIME type. The returned type never carries a charset.
// ContentType builds on Lookup (or accepts a MIME type directly when the input
// already contains a "/") and appends "; charset=<cs>" when the type has a
// declared charset and the value does not already mention one, giving you a
// value suitable for a Content-Type header in one call.
//
// Charset reports the charset that applies to a given MIME type. Following
// mime-db, that is the charset the type itself declares — UTF-8 for
// application/json, application/javascript and a list of XML- and JSON-based
// types, US-ASCII for the news types, 7-BIT for application/prs.cyn — with a
// blanket UTF-8 fallback for anything under text/. Note what this does *not*
// do: a structured suffix is not enough on its own, so application/xml,
// application/ld+json and image/svg+xml all report no charset, exactly as
// upstream does.
//
// The reverse direction is Extension, which takes a MIME type (optionally with
// parameters, which are ignored) and returns the *first* extension mime-db
// lists for it, without a leading dot. That is not always the most familiar
// spelling: image/jpeg maps back to "jpeg" rather than "jpg", audio/mpeg to
// "mpga" rather than "mp3", video/quicktime to "qt" and text/markdown to
// "markdown". Some types mime-db knows carry no extensions at all
// (text/javascript, audio/opus, audio/flac), and those report nothing.
//
// Edge cases follow a consistent rule: every function returns a boolean second
// value that is false when nothing is known. An empty string, an extension that
// is not in the table, or a MIME type with no known extension yields ("",
// false) rather than a guess, so callers can distinguish "unknown" from a real
// answer. Compared with the Node original the API surface is adapted to Go —
// results are (value, ok) pairs instead of a value-or-false union, and the
// bundled type table is a fixed subset rather than the full generated mime-db —
// but the lookup semantics, charset rules and extension choices mirror
// mime-types.
package mimetypes

import (
	"regexp"
	"strings"
)

// typeByExt maps a lower-case file extension (without dot) to a MIME type. Each
// value is what mime-types@2.1.35 resolves that extension to, including the
// cases where mime-db's source ranking picks a type one might not expect
// (".js" is application/javascript, ".opus" is audio/ogg, ".wav" is audio/wave).
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
	"ico":      "image/vnd.microsoft.icon",
	"bmp":      "image/bmp",
	"tiff":     "image/tiff",
	"tif":      "image/tiff",
	"pdf":      "application/pdf",
	"zip":      "application/zip",
	"gz":       "application/gzip",
	"tar":      "application/x-tar",
	"rar":      "application/vnd.rar",
	"7z":       "application/x-7z-compressed",
	"bz2":      "application/x-bzip2",
	"mp4":      "video/mp4",
	"m4v":      "video/x-m4v",
	"mpeg":     "video/mpeg",
	"mpg":      "video/mpeg",
	"mov":      "video/quicktime",
	"avi":      "video/x-msvideo",
	"webm":     "video/webm",
	"mp3":      "audio/mpeg",
	"m4a":      "audio/mp4",
	"wav":      "audio/wave",
	"ogg":      "audio/ogg",
	"oga":      "audio/ogg",
	"opus":     "audio/ogg",
	"flac":     "audio/x-flac",
	"woff":     "font/woff",
	"woff2":    "font/woff2",
	"ttf":      "font/ttf",
	"otf":      "font/otf",
	"eot":      "application/vnd.ms-fontobject",
	"wasm":     "application/wasm",
	"bin":      "application/octet-stream",
	"exe":      "application/x-msdos-program",
	"dll":      "application/x-msdownload",
	"doc":      "application/msword",
	"xls":      "application/vnd.ms-excel",
	"ppt":      "application/vnd.ms-powerpoint",
	"rtf":      "application/rtf",
	"yaml":     "text/yaml",
	"yml":      "text/yaml",
	"ics":      "text/calendar",
}

// extByType maps a MIME type to the first extension mime-db lists for it, which
// is the one mime-types reports. Types that mime-db knows but gives no
// extensions are deliberately absent.
var extByType = map[string]string{
	"application/gzip":              "gz",
	"application/javascript":        "js",
	"application/json":              "json",
	"application/ld+json":           "jsonld",
	"application/manifest+json":     "webmanifest",
	"application/msword":            "doc",
	"application/octet-stream":      "bin",
	"application/pdf":               "pdf",
	"application/rtf":               "rtf",
	"application/vnd.ms-excel":      "xls",
	"application/vnd.ms-fontobject": "eot",
	"application/vnd.ms-powerpoint": "ppt",
	"application/vnd.rar":           "rar",
	"application/wasm":              "wasm",
	"application/x-7z-compressed":   "7z",
	"application/x-bzip2":           "bz2",
	"application/x-msdos-program":   "exe",
	"application/x-msdownload":      "exe",
	"application/x-tar":             "tar",
	"application/xml":               "xml",
	"application/zip":               "zip",
	"audio/mp4":                     "m4a",
	"audio/mpeg":                    "mpga",
	"audio/ogg":                     "oga",
	"audio/wav":                     "wav",
	"audio/wave":                    "wav",
	"audio/x-flac":                  "flac",
	"font/otf":                      "otf",
	"font/ttf":                      "ttf",
	"font/woff":                     "woff",
	"font/woff2":                    "woff2",
	"image/bmp":                     "bmp",
	"image/gif":                     "gif",
	"image/jpeg":                    "jpeg",
	"image/png":                     "png",
	"image/svg+xml":                 "svg",
	"image/tiff":                    "tif",
	"image/vnd.microsoft.icon":      "ico",
	"image/webp":                    "webp",
	"image/x-icon":                  "ico",
	"text/calendar":                 "ics",
	"text/css":                      "css",
	"text/csv":                      "csv",
	"text/html":                     "html",
	"text/markdown":                 "markdown",
	"text/plain":                    "txt",
	"text/xml":                      "xml",
	"text/yaml":                     "yaml",
	"video/mp4":                     "mp4",
	"video/mpeg":                    "mpeg",
	"video/quicktime":               "qt",
	"video/webm":                    "webm",
	"video/x-m4v":                   "m4v",
	"video/x-msvideo":               "avi",
}

// charsetByType is mime-db's complete set of types that declare a charset. Any
// other type falls back to the text/* rule in Charset, or to no charset at all.
var charsetByType = map[string]string{
	"application/3gpdash-qoe-report+xml":    "UTF-8",
	"application/beep+xml":                  "UTF-8",
	"application/elm+json":                  "UTF-8",
	"application/emergencycalldata.cap+xml": "UTF-8",
	"application/fhir+json":                 "UTF-8",
	"application/fhir+xml":                  "UTF-8",
	"application/im-iscomposing+xml":        "UTF-8",
	"application/javascript":                "UTF-8",
	"application/json":                      "UTF-8",
	"application/manifest+json":             "UTF-8",
	"application/msc-ivr+xml":               "UTF-8",
	"application/msc-mixer+xml":             "UTF-8",
	"application/news-checkgroups":          "US-ASCII",
	"application/news-groupinfo":            "US-ASCII",
	"application/pidf+xml":                  "UTF-8",
	"application/pidf-diff+xml":             "UTF-8",
	"application/poc-settings+xml":          "UTF-8",
	"application/prs.cyn":                   "7-BIT",
	"application/vnd.hl7cda+xml":            "UTF-8",
	"application/vnd.hl7v2+xml":             "UTF-8",
	"application/vnd.omads-email+xml":       "UTF-8",
	"application/vnd.omads-file+xml":        "UTF-8",
	"application/vnd.omads-folder+xml":      "UTF-8",
	"application/vnd.syncml+xml":            "UTF-8",
	"application/vnd.syncml.dm+wbxml":       "UTF-8",
	"application/vnd.syncml.dm+xml":         "UTF-8",
	"application/vnd.syncml.dmddf+xml":      "UTF-8",
	"application/vnd.syncml.dmtnds+xml":     "UTF-8",
	"application/vnd.wap.wbxml":             "UTF-8",
	"text/css":                              "UTF-8",
	"text/n3":                               "UTF-8",
	"text/parameters":                       "UTF-8",
	"text/provenance-notation":              "UTF-8",
	"text/turtle":                           "UTF-8",
	"text/vnd.debian.copyright":             "UTF-8",
	"text/vnd.esmertec.theme-descriptor":    "UTF-8",
	"text/vnd.sun.j2me.app-descriptor":      "UTF-8",
	"text/vnd.trolltech.linguist":           "UTF-8",
	"text/vtt":                              "UTF-8",
}

// extractType isolates the media type at the start of a header value, matching
// upstream's EXTRACT_TYPE_REGEXP: leading whitespace is skipped and the type
// ends at the first ';' or whitespace.
var extractType = regexp.MustCompile(`^[ \t\n\r\f\v]*([^;\s]*)(?:;|\s|$)`)

// textType matches the types that get the blanket UTF-8 fallback.
var textType = regexp.MustCompile(`(?i)^text/`)

// mediaTypeOf returns the lower-cased media type at the head of s, or "" when
// there is none.
func mediaTypeOf(s string) string {
	m := extractType.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
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

// Charset returns the charset for a MIME type, or false if none applies. The
// value is returned in mime-db's canonical casing ("UTF-8", "US-ASCII");
// ContentType lower-cases it when building a header value. Parameters on the
// input are ignored.
func Charset(mimeType string) (string, bool) {
	if mimeType == "" {
		return "", false
	}
	t := mediaTypeOf(mimeType)
	if cs, ok := charsetByType[t]; ok {
		return cs, true
	}
	if textType.MatchString(t) {
		return "UTF-8", true
	}
	return "", false
}

// ContentType returns a full Content-Type value for a MIME type or extension.
// If the input already contains a "/" it is treated as a MIME type; otherwise
// it is looked up as an extension. A "; charset=..." parameter is appended when
// the type declares a charset and the value does not already mention one. Like
// upstream, the input is not trimmed: whatever is passed in is echoed back.
func ContentType(typeOrExt string) (string, bool) {
	if typeOrExt == "" {
		return "", false
	}
	mime := typeOrExt
	if !strings.ContainsRune(typeOrExt, '/') {
		var ok bool
		mime, ok = Lookup(typeOrExt)
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

// Extension returns the first file extension (without a leading dot) that
// mime-db lists for a MIME type, ignoring any parameters. The boolean is false
// when the type is unknown or has no extensions.
func Extension(mimeType string) (string, bool) {
	if mimeType == "" {
		return "", false
	}
	ext, ok := extByType[mediaTypeOf(mimeType)]
	return ext, ok
}
