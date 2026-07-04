// Package htmlentities encodes and decodes HTML entities, providing a subset
// of the behavior of the npm "html-entities" library.
//
// Encode supports two modes: "specialChars" (the default) encodes only the
// characters & < > " ' as named entities, and "nonAscii" additionally encodes
// every non-ASCII rune (code point > 127) as a numeric entity. Decode
// understands a table of common named entities as well as decimal (&#NNN;) and
// hexadecimal (&#xHH;) numeric references.
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
