// Package escapehtml escapes special characters in a string for safe inclusion
// in HTML.
//
// It is a port of the npm package "escape-html". It escapes the five
// characters &, <, >, ", and ' into their HTML entity equivalents.
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
