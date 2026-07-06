// Package htmlentities encodes and decodes HTML entities, providing a subset
// of the behavior of the npm "html-entities" library. It exists so that Go code
// ported from a Node codebase can keep calling a familiar encode/decode pair
// when rendering user-supplied text into HTML or when reading entity-laden text
// back out again, without pulling in a third-party dependency.
//
// Escaping HTML entities is the front line of defense against cross-site
// scripting: turning the characters & < > " ' into their named forms prevents
// user input from being interpreted as markup or from breaking out of an
// attribute value. The npm html-entities library grew popular precisely because
// it makes this safe by default while still allowing a stricter mode that also
// escapes non-ASCII text for transports that are not UTF-8 clean. This port
// keeps that same two-mode split and matches the library's output for the
// characters and entities it supports.
//
// Encode supports two modes selected via EncodeOptions.Mode. The default,
// "specialChars", encodes only the five characters & < > " ' as the named
// entities &amp; &lt; &gt; &quot; and &apos;, leaving all other runes
// (including non-ASCII text such as accented letters) untouched. The "nonAscii"
// mode does everything specialChars does and additionally rewrites every rune
// with a code point above 127 as a decimal numeric entity (for example é
// becomes &#233;), which is useful when the output must be pure ASCII. Any
// unrecognized Mode value falls back to specialChars behavior.
//
// Decode is the inverse and is intentionally more permissive than Encode. It
// resolves the named entities in a built-in table (the five specials plus a
// selection of common ones such as &copy;, &nbsp;, &mdash; and several typographic
// and currency symbols) as well as decimal (&#233;) and hexadecimal (&#xe9; or
// &#Xe9;) numeric references. A leading fast path returns the input unchanged
// when it contains no ampersand. Anything that does not form a recognized entity
// is left exactly as-is: a bare or trailing ampersand, an unknown name like
// &unknownentity;, and malformed numerics like &#zz; all pass through untouched,
// and the scanner only looks a bounded distance ahead for the terminating
// semicolon so stray ampersands never swallow following text.
//
// Parity with the Node package is partial by design. The named-entity table is
// a curated subset rather than the full HTML5 entity set, so Decode will leave
// less common named entities unresolved, and Encode never emits named forms
// beyond the five specials. Because Encode and Decode agree on those five
// characters, specialChars-mode output round-trips exactly through Decode, which
// is the common case for escaping and later unescaping application text.
// DecodeOptions is accepted for API compatibility but currently has no effect.
package htmlentities

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// EncodeOptions configures Encode. Mode may be "specialChars" (default) or
// "nonAscii".
type EncodeOptions struct {
	// Mode selects the encoding strategy: "specialChars" or "nonAscii".
	Mode string
}

// DecodeOptions configures Decode. It is currently a placeholder for future
// options and may be passed as the zero value.
type DecodeOptions struct {
	// Scope is reserved for future use and currently has no effect.
	Scope string
}

// specialCharEntities maps the five special characters to their named entity.
var specialCharEntities = map[rune]string{
	'&':  "&amp;",
	'<':  "&lt;",
	'>':  "&gt;",
	'"':  "&quot;",
	'\'': "&apos;",
}

// namedEntities maps a reasonable set of named entities to their rune value,
// used by Decode.
var namedEntities = map[string]rune{
	"amp":    '&',
	"lt":     '<',
	"gt":     '>',
	"quot":   '"',
	"apos":   '\'',
	"nbsp":   '\u00a0',
	"copy":   '©',
	"reg":    '®',
	"trade":  '™',
	"hellip": '…',
	"mdash":  '—',
	"ndash":  '–',
	"lsquo":  '‘',
	"rsquo":  '’',
	"ldquo":  '“',
	"rdquo":  '”',
	"laquo":  '«',
	"raquo":  '»',
	"cent":   '¢',
	"pound":  '£',
	"yen":    '¥',
	"euro":   '€',
	"sect":   '§',
	"para":   '¶',
	"middot": '·',
	"deg":    '°',
	"plusmn": '±',
	"times":  '×',
	"divide": '÷',
	"frac12": '½',
	"frac14": '¼',
	"frac34": '¾',
}

// Encode converts special characters in s to HTML entities according to the
// supplied options. When no options are given, the "specialChars" mode is used.
func Encode(s string, opts ...EncodeOptions) string {
	mode := "specialChars"
	if len(opts) > 0 && opts[0].Mode != "" {
		mode = opts[0].Mode
	}
	nonAscii := mode == "nonAscii"

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if ent, ok := specialCharEntities[r]; ok {
			b.WriteString(ent)
			continue
		}
		if nonAscii && r > 127 {
			b.WriteString("&#")
			b.WriteString(strconv.Itoa(int(r)))
			b.WriteByte(';')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Decode converts HTML entities in s back to their literal characters. It
// handles named entities from a built-in table plus decimal and hexadecimal
// numeric references. Unrecognized entities are left unchanged.
func Decode(s string, opts ...DecodeOptions) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		// Find the terminating semicolon within a reasonable window.
		end := -1
		for j := i + 1; j < len(s) && j < i+34; j++ {
			if s[j] == ';' {
				end = j
				break
			}
			if s[j] == '&' || s[j] == ' ' {
				break
			}
		}
		if end == -1 {
			b.WriteByte('&')
			i++
			continue
		}
		body := s[i+1 : end]
		if r, ok := decodeEntityBody(body); ok {
			b.WriteRune(r)
			i = end + 1
			continue
		}
		b.WriteByte('&')
		i++
	}
	return b.String()
}

// decodeEntityBody decodes the text between '&' and ';' (exclusive), returning
// the resolved rune. It reports false if the body is not a recognized entity.
func decodeEntityBody(body string) (rune, bool) {
	if body == "" {
		return 0, false
	}
	if body[0] == '#' {
		if len(body) < 2 {
			return 0, false
		}
		var n int64
		var err error
		if body[1] == 'x' || body[1] == 'X' {
			n, err = strconv.ParseInt(body[2:], 16, 32)
		} else {
			n, err = strconv.ParseInt(body[1:], 10, 32)
		}
		if err != nil || n < 0 || n > utf8.MaxRune {
			return 0, false
		}
		return rune(n), true
	}
	if r, ok := namedEntities[body]; ok {
		return r, true
	}
	return 0, false
}
