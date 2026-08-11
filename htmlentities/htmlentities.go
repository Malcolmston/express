// Package htmlentities encodes and decodes HTML entities, a port of the npm
// "html-entities" library. It exists so that Go code ported from a Node codebase
// can keep calling a familiar encode/decode pair when rendering user-supplied
// text into HTML or when reading entity-laden text back out again, without
// pulling in a third-party dependency.
//
// Escaping HTML entities is the front line of defense against cross-site
// scripting: turning the characters & < > " ' into their named forms prevents
// user input from being interpreted as markup or from breaking out of an
// attribute value. The npm html-entities library grew popular precisely because
// it makes this safe by default while still allowing stricter modes that also
// escape non-ASCII or non-printable text for transports that are not UTF-8
// clean. This port keeps the same option surface and shares the library's own
// reference tables, which are generated straight out of the installed package
// into tables.go, so the two agree entity for entity rather than approximately.
//
// # Encoding
//
// Encode selects which characters to touch with EncodeOptions.Mode, which
// character set of named entities to draw on with EncodeOptions.Level, and the
// spelling of the fallback numeric reference with EncodeOptions.Numeric. The
// defaults match upstream: mode "specialChars", level "all", numeric "decimal".
//
//   - ModeSpecialChars (default) encodes only & < > " '.
//   - ModeNonASCII additionally encodes every code point above U+007F.
//   - ModeNonASCIIPrintable additionally encodes the non-printable control
//     characters (using upstream's exact, slightly ragged ranges) and U+007F up.
//   - ModeNonASCIIPrintableOnly is the same but leaves the five specials alone.
//   - ModeExtensive encodes essentially all ASCII punctuation as well, for
//     contexts where nothing but letters, digits and spaces may pass through.
//
// A character is emitted as the named entity the level's table offers for it and
// otherwise as a numeric reference (&#233; or, with NumericHexadecimal, &#xe9;).
// LevelXML restricts the table to the five XML entities, LevelHTML4 to the HTML
// 4 set, and LevelHTML5 / LevelAll to the full HTML 5 set, so
// Encode("é", EncodeOptions{Mode: ModeNonASCII}) yields "&eacute;" while the
// same call at LevelXML yields "&#233;".
//
// # Decoding
//
// Decode resolves entity references back to text, and is deliberately more
// permissive than Encode. DecodeOptions.Scope chooses how forgiving it is, and
// matches how a browser reads the two positions an entity can appear in:
//
//   - ScopeBody (the default for every level but XML) accepts the legacy named
//     references that have no trailing semicolon, so "&amp" decodes to "&", just
//     as an HTML body would read it.
//   - ScopeAttribute is the same except that a reference followed by "=" is left
//     alone, mirroring the HTML rule that keeps "?a&copy=1" intact.
//   - ScopeStrict (the default at LevelXML) requires the semicolon and leaves
//     "&amp" untouched.
//
// Numeric references are resolved with the spec's fix-ups rather than naively:
// the C1 range 0x80..0x9F is remapped to the Windows-1252 characters browsers
// substitute there, U+0000 becomes U+FFFD, and a value at or above U+10FFFF
// becomes U+FFFD. Anything that does not form a recognized reference — a bare
// ampersand, an unknown name like &unknownentity;, a malformed numeric like
// &#zz; — is left exactly as it was found. DecodeEntity applies the same
// resolution to a single reference given on its own.
//
// # Parity notes
//
// The port is a faithful port, not a subset: the modes, levels, scopes and
// reference tables all come from html-entities, and the body-scope matching
// patterns are the upstream regexes verbatim (they use nothing outside RE2, and
// Go alternation is leftmost-first as JavaScript's is). Two differences are
// inherent to the language rather than chosen. Upstream operates on UTF-16 code
// units, so a numeric reference to an unpaired surrogate such as &#xD800;
// produces a lone surrogate in JavaScript; a Go string cannot hold one, so this
// port yields U+FFFD. And where JavaScript coerces null and undefined to the
// empty string, Go simply takes a string.
package htmlentities

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// Encode modes, mirroring the html-entities EncodeMode union.
const (
	// ModeSpecialChars encodes only & < > " ' — the default.
	ModeSpecialChars = "specialChars"
	// ModeNonASCII also encodes every code point above U+007F.
	ModeNonASCII = "nonAscii"
	// ModeNonASCIIPrintable also encodes the non-printable control characters.
	ModeNonASCIIPrintable = "nonAsciiPrintable"
	// ModeNonASCIIPrintableOnly encodes the non-printables and non-ASCII but
	// leaves the five special characters alone.
	ModeNonASCIIPrintableOnly = "nonAsciiPrintableOnly"
	// ModeExtensive additionally encodes most ASCII punctuation.
	ModeExtensive = "extensive"
)

// Entity table levels, mirroring the html-entities Level union. LevelAll is an
// alias for LevelHTML5, as it is upstream.
const (
	LevelXML   = "xml"
	LevelHTML4 = "html4"
	LevelHTML5 = "html5"
	LevelAll   = "all"
)

// Numeric reference spellings for Encode's fallback.
const (
	NumericDecimal     = "decimal"
	NumericHexadecimal = "hexadecimal"
)

// Decode scopes, mirroring the html-entities DecodeScope union.
const (
	// ScopeStrict requires a trailing semicolon on every reference.
	ScopeStrict = "strict"
	// ScopeBody accepts the legacy semicolon-less named references.
	ScopeBody = "body"
	// ScopeAttribute is ScopeBody except that a reference followed by "=" is
	// left alone.
	ScopeAttribute = "attribute"
)

// EncodeOptions configures Encode. The zero value means mode "specialChars",
// level "all" and decimal numeric references, matching the upstream defaults.
type EncodeOptions struct {
	// Mode selects which characters are encoded. See the Mode constants; an
	// unrecognized value falls back to ModeSpecialChars.
	Mode string
	// Level selects the named-entity table. See the Level constants; an
	// unrecognized value falls back to LevelAll.
	Level string
	// Numeric selects the spelling of the numeric fallback reference, either
	// NumericDecimal (default) or NumericHexadecimal.
	Numeric string
}

// DecodeOptions configures Decode and DecodeEntity. The zero value means level
// "all" and, since that is not LevelXML, scope "body".
type DecodeOptions struct {
	// Level selects the named-entity table. See the Level constants; an
	// unrecognized value falls back to LevelAll.
	Level string
	// Scope selects how strict reference matching is. See the Scope constants.
	// Empty means ScopeStrict at LevelXML and ScopeBody otherwise, as upstream.
	Scope string
}

// replacementChar is what a numeric reference to a disallowed code point
// resolves to.
const replacementChar = '�'

// entitiesFor returns the entity-spelling table for a level.
func entitiesFor(level string) map[string]string {
	switch level {
	case LevelXML:
		return entitiesXML
	case LevelHTML4:
		return entitiesHTML4
	default: // LevelHTML5, LevelAll, anything unrecognized
		return entitiesHTML5
	}
}

// charactersFor returns the character-to-named-entity table for a level.
func charactersFor(level string) map[string]string {
	switch level {
	case LevelXML:
		return charactersXML
	case LevelHTML4:
		return charactersHTML4
	default:
		return charactersHTML5
	}
}

// normalizeLevel maps an unrecognized level onto the upstream default.
func normalizeLevel(level string) string {
	switch level {
	case LevelXML, LevelHTML4, LevelHTML5, LevelAll:
		return level
	default:
		return LevelAll
	}
}

// ---------------------------------------------------------------------------
// Encode
// ---------------------------------------------------------------------------

// shouldEncode reports whether r is in the set the given mode encodes. The
// ranges are upstream's encodeRegExps translated from UTF-16 code units to code
// points; the gaps in the control-character ranges (U+0000, U+0009..U+0010 and
// U+0016 are absent) are upstream's and are reproduced deliberately.
func shouldEncode(mode string, r rune) bool {
	special := r == '<' || r == '>' || r == '\'' || r == '"' || r == '&'
	switch mode {
	case ModeNonASCII:
		return special || r >= 0x80
	case ModeNonASCIIPrintable:
		return special || nonPrintable(r) || r >= 0x7f
	case ModeNonASCIIPrintableOnly:
		return nonPrintable(r) || r >= 0x7f
	case ModeExtensive:
		return extensivePunct(r) || r >= 0x7f
	default: // ModeSpecialChars and anything unrecognized
		return special
	}
}

// nonPrintable reports whether r is in the control-character ranges the
// nonAsciiPrintable modes encode.
func nonPrintable(r rune) bool {
	return (r >= 0x01 && r <= 0x08) || (r >= 0x11 && r <= 0x15) || (r >= 0x17 && r <= 0x1f)
}

// extensivePunct reports whether r is in the ASCII ranges ModeExtensive encodes.
func extensivePunct(r rune) bool {
	switch {
	case r >= 0x01 && r <= 0x0c,
		r >= 0x0e && r <= 0x1f,
		r >= 0x21 && r <= 0x2c,
		r >= 0x2e && r <= 0x2f,
		r >= 0x3a && r <= 0x40,
		r >= 0x5b && r <= 0x60,
		r >= 0x7b && r <= 0x7d:
		return true
	}
	return false
}

// Encode converts characters in s to HTML entities according to opts. Passing no
// options means mode "specialChars" at level "all" with decimal numeric
// fallbacks, the upstream defaults. Characters the level has a named entity for
// are written in that form; the rest become numeric references.
func Encode(s string, opts ...EncodeOptions) string {
	var o EncodeOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if s == "" {
		return ""
	}
	characters := charactersFor(normalizeLevel(o.Level))
	isHex := o.Numeric == NumericHexadecimal

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if !shouldEncode(o.Mode, r) {
			b.WriteRune(r)
			continue
		}
		if ent, ok := characters[string(r)]; ok {
			b.WriteString(ent)
			continue
		}
		if isHex {
			b.WriteString("&#x")
			b.WriteString(strconv.FormatInt(int64(r), 16))
		} else {
			b.WriteString("&#")
			b.WriteString(strconv.Itoa(int(r)))
		}
		b.WriteByte(';')
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Decode
// ---------------------------------------------------------------------------

// The strict and attribute scopes match the same reference shapes at every
// level; only the body scope varies, and its pattern comes from bodyPattern.
const (
	strictPattern    = `&(?:#\d+|#[xX][\da-fA-F]+|[0-9a-zA-Z]+);`
	attributePattern = `&(?:#\d+|#[xX][\da-fA-F]+|[0-9a-zA-Z]+)[;=]?`
)

var (
	reCacheMu sync.Mutex
	reCache   = map[string]*regexp.Regexp{}
)

// decodeRE returns the compiled matcher for a level and scope, compiling each
// distinct pattern at most once.
func decodeRE(level, scope string) *regexp.Regexp {
	var pattern string
	switch scope {
	case ScopeStrict:
		pattern = strictPattern
	case ScopeAttribute:
		pattern = attributePattern
	default: // ScopeBody
		p, ok := bodyPattern[level]
		if !ok {
			p = bodyPattern[LevelHTML5]
		}
		pattern = p
	}
	reCacheMu.Lock()
	defer reCacheMu.Unlock()
	if re, ok := reCache[pattern]; ok {
		return re
	}
	re := regexp.MustCompile(pattern)
	reCache[pattern] = re
	return re
}

// defaultScope mirrors upstream's `scope = level === 'xml' ? 'strict' : 'body'`.
func defaultScope(level string) string {
	if level == LevelXML {
		return ScopeStrict
	}
	return ScopeBody
}

// normalizeScope maps an unrecognized scope onto the level's default.
func normalizeScope(level, scope string) string {
	switch scope {
	case ScopeStrict, ScopeBody, ScopeAttribute:
		return scope
	default:
		return defaultScope(level)
	}
}

// Decode converts every HTML entity reference in s back to the text it stands
// for. Passing no options means level "all" in "body" scope, so the legacy
// semicolon-less references decode too. A reference that is not recognized is
// left exactly as it appears.
func Decode(s string, opts ...DecodeOptions) string {
	var o DecodeOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if s == "" {
		return ""
	}
	level := normalizeLevel(o.Level)
	scope := normalizeScope(level, o.Scope)
	references := entitiesFor(level)
	re := decodeRE(level, scope)

	return re.ReplaceAllStringFunc(s, func(entity string) string {
		return decodedEntity(entity, references, scope == ScopeAttribute, scope == ScopeStrict)
	})
}

// DecodeEntity resolves a single entity reference, such as "&copy;" or "&#233;",
// and returns it unchanged when it is not recognized. It is the port of
// html-entities' decodeEntity, which — like this function — ignores the scope
// rules and only consults the level's table.
func DecodeEntity(entity string, opts ...DecodeOptions) string {
	var o DecodeOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	if entity == "" {
		return ""
	}
	return decodedEntity(entity, entitiesFor(normalizeLevel(o.Level)), false, false)
}

// decodedEntity resolves one matched reference, mirroring upstream's
// getDecodedEntity: an attribute-scope reference ending in "=" and a
// strict-scope reference with no ";" are left alone, a named reference is looked
// up in the level's table, and a numeric reference is parsed with the spec's
// remappings.
func decodedEntity(entity string, references map[string]string, isAttribute, isStrict bool) string {
	last := entity[len(entity)-1]
	if isAttribute && last == '=' {
		return entity
	}
	if isStrict && last != ';' {
		return entity
	}
	if replacement, ok := references[entity]; ok {
		return replacement
	}
	if len(entity) > 2 && entity[0] == '&' && entity[1] == '#' {
		return decodeNumeric(entity)
	}
	return entity
}

// decodeNumeric resolves a numeric character reference of the form &#123; or
// &#x7b;. A trailing semicolon is optional, since body scope may have matched
// without one.
func decodeNumeric(entity string) string {
	body := strings.TrimSuffix(entity[2:], ";")
	var n int64
	var err error
	if len(body) > 1 && (body[0] == 'x' || body[0] == 'X') {
		n, err = strconv.ParseInt(body[1:], 16, 64)
	} else {
		n, err = strconv.ParseInt(body, 10, 64)
	}
	switch {
	case err == nil:
	case errors.Is(err, strconv.ErrRange):
		// Too large for int64 is certainly out of the Unicode range; upstream
		// arrives at the same answer via a float parse.
		return string(replacementChar)
	default:
		// Upstream reaches String.fromCharCode(NaN) here, which is U+0000. Only
		// DecodeEntity can get this far: the Decode patterns require digits.
		return "\x00"
	}
	// Upstream tests `>= 0x10ffff`, so U+10FFFF itself is replaced too.
	if n >= 0x10ffff {
		return string(replacementChar)
	}
	if n > 0xffff {
		return string(rune(n))
	}
	if mapped, ok := numericUnicodeMap[int(n)]; ok {
		return string(mapped)
	}
	// A reference to an unpaired surrogate has no representation in a Go
	// string; upstream's UTF-16 output keeps the lone surrogate. See the
	// package documentation.
	if n >= 0xd800 && n <= 0xdfff {
		return string(replacementChar)
	}
	if n < 0 || n > utf8.MaxRune {
		return string(replacementChar)
	}
	return string(rune(n))
}
