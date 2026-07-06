// Package escapehtml escapes the HTML-significant characters in a string so it
// can be safely embedded in HTML markup, a port of the npm "escape-html" package
// that Express uses when rendering error pages and other user-influenced text.
// It exposes the single Escape function and depends on only the Go standard
// library.
//
// Escaping is the baseline defense against HTML injection and cross-site
// scripting. Any time untrusted or merely unpredictable text — a form value, a
// filename, a database field, an error message — is placed into an HTML document
// or an attribute value, the characters that HTML treats specially must be turned
// into entities so a browser renders them as literal text instead of
// interpreting them as tags or attribute delimiters. Escape gives you exactly
// that conversion in one call.
//
// Escape rewrites the five characters that escape-html handles: & becomes &amp;,
// < becomes &lt;, > becomes &gt;, " becomes &quot;, and ' becomes &#39; (the
// numeric entity for the apostrophe). Ampersand is converted first in effect —
// because it is matched and replaced along with the others in a single pass —
// so an existing entity in the input has its leading & escaped too and nothing
// is double-interpreted. Escaping both quote characters means the result is safe
// to drop into either a single- or double-quoted attribute value, not just
// element text.
//
// The implementation is allocation-conscious and precise. It first scans for any
// of the five characters and, finding none, returns the original string
// unchanged with no allocation; this makes the common case of already-safe text
// cheap. When a special character is present it copies the safe prefix and then
// rewrites the remainder byte by byte. Because the five characters are all ASCII
// and the transform operates a byte at a time, multi-byte UTF-8 sequences are
// passed through untouched, so non-ASCII text such as accented letters or emoji
// survives verbatim. The empty string returns the empty string.
//
// Parity with the Node original is exact for the character set and the entity
// spellings: the same five characters map to the same entities, and text lacking
// them is returned as-is. The only differences are idiomatic — Escape takes and
// returns a Go string rather than coercing a JavaScript value, and it is a pure
// function with no configuration. Note the deliberately narrow scope: this is
// HTML body/attribute escaping only, not a URL, JavaScript-string, or CSS
// encoder, so it must not be relied upon to sanitize values destined for those
// other contexts.
package escapehtml

import "strings"

// Escape returns s with the HTML-significant characters &, <, >, ", and '
// replaced by their entity equivalents:
//
//	& -> &amp;
//	< -> &lt;
//	> -> &gt;
//	" -> &quot;
//	' -> &#39;
//
// If s contains none of these characters, it is returned unchanged.
func Escape(s string) string {
	idx := strings.IndexAny(s, "&<>\"'")
	if idx == -1 {
		return s
	}

	var b strings.Builder
	b.Grow(len(s) + 8)
	b.WriteString(s[:idx])
	for i := idx; i < len(s); i++ {
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
