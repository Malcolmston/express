// Package striptags removes HTML tags from a string while keeping the text
// content, mirroring the behavior of the npm "striptags" library. It is a small
// utility for turning a fragment of markup into plain text, for example when
// generating a preview, building a search index, or sanitizing a value that
// must not contain any HTML at all.
//
// By default every tag is stripped: the angle-bracket markup is removed and the
// text that sits between tags is preserved. When one or more allowed tag names
// are supplied to StripTags as variadic arguments, those tags (and their
// matching closing tags) are kept verbatim while all other tags are removed.
// Allowed tag names may be written either bare ("p") or wrapped in angle
// brackets ("<p>"); normalizeAllowed accepts either form, and comparison is
// case-insensitive.
//
// The algorithm is a single-pass character state machine rather than a full
// HTML parser. It moves between three states: plaintext (copying characters to
// the output), in-tag (buffering characters after a '<' until the matching
// '>'), and in-comment (consuming everything from "<!--" through "-->"). When a
// complete tag is buffered, its name is extracted and the tag is emitted only
// if the name is in the allowed set; otherwise it is dropped. This mirrors the
// original striptags, which is also a hand-rolled state machine and not a
// spec-compliant parser.
//
// Several edge cases follow the reference library's behavior. HTML comments are
// always removed regardless of the allowed list. A '<' immediately followed by
// a space or newline is not the start of a tag: the "< " is emitted literally
// and parsing returns to plaintext, so text like "a < b" survives untouched.
// While inside a tag, quotes are tracked so that a '<' or '>' appearing inside a
// quoted attribute value is not mistaken for a tag delimiter, and additional
// unquoted '<' characters merely increase a nesting depth (matched by later
// '>') rather than being copied out. An unterminated tag run at the end of the
// input (a trailing '<' with no closing '>') is discarded, matching striptags
// which drops trailing partial tags.
//
// This port intentionally keeps the surface small. Unlike some configurations
// of the Node package it does not replace stripped tags with a substitute
// string, does not decode or encode HTML entities, and does not attempt to
// validate attribute contents; the text between tags, including any entities,
// is passed through unchanged. The allowed-tag matching is purely by tag name,
// so attributes on an allowed tag are preserved exactly as written.
package striptags

import "strings"

// parser states.
const (
	statePlaintext = iota
	stateHTML
	stateComment
)

// StripTags removes HTML tags from html. Any tag name listed in allowed is
// preserved (both its opening and closing forms); all other tags are removed
// and their inner text is kept. HTML comments are always removed.
func StripTags(html string, allowed ...string) string {
	allowedSet := make(map[string]bool, len(allowed))
	for _, t := range allowed {
		allowedSet[strings.ToLower(normalizeAllowed(t))] = true
	}

	var out strings.Builder
	out.Grow(len(html))
	var tag strings.Builder

	state := statePlaintext
	depth := 0
	var inQuote byte // 0 when not inside a quoted attribute value

	for i := 0; i < len(html); i++ {
		c := html[i]
		switch state {
		case statePlaintext:
			if c == '<' {
				state = stateHTML
				tag.WriteByte(c)
			} else {
				out.WriteByte(c)
			}
		case stateHTML:
			switch c {
			case '<':
				// Ignore '<' inside a quoted value; otherwise it is a
				// nested delimiter that raises the depth (matched by a
				// later '>') rather than closing the tag.
				if inQuote != 0 {
					break
				}
				depth++
			case '>':
				if inQuote != 0 {
					break
				}
				if depth > 0 {
					depth--
					break
				}
				tag.WriteByte('>')
				full := tag.String()
				if allowedSet[tagName(full)] {
					out.WriteString(full)
				}
				tag.Reset()
				inQuote = 0
				state = statePlaintext
			case '"', '\'':
				if c == inQuote {
					inQuote = 0
				} else if inQuote == 0 {
					inQuote = c
				}
				tag.WriteByte(c)
			case '-':
				// "<!-" followed by '-' begins an HTML comment.
				if tag.String() == "<!-" {
					state = stateComment
				}
				tag.WriteByte(c)
			case ' ', '\n':
				// A bare "<" followed by whitespace is not a tag: emit the
				// "< " literally and return to plaintext.
				if tag.String() == "<" {
					out.WriteString("< ")
					tag.Reset()
					state = statePlaintext
					break
				}
				tag.WriteByte(c)
			default:
				tag.WriteByte(c)
			}
		case stateComment:
			if c == '>' {
				// Close only when the two characters before '>' are "--".
				if s := tag.String(); len(s) >= 2 && s[len(s)-2:] == "--" {
					state = statePlaintext
				}
				tag.Reset()
			} else {
				tag.WriteByte(c)
			}
		}
	}

	// Any unterminated '<' run is discarded, matching striptags which drops
	// trailing partial tags.
	return out.String()
}

// normalizeAllowed strips surrounding angle brackets from an allowed tag
// specifier such as "<p>" so that either form is accepted.
func normalizeAllowed(t string) string {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "<")
	t = strings.TrimSuffix(t, ">")
	t = strings.TrimPrefix(t, "/")
	return strings.TrimSpace(t)
}

// tagName extracts the lower-cased tag name from a full tag token such as
// "<a href=...>" or "</a>" or "<br/>".
func tagName(tag string) string {
	s := tag
	s = strings.TrimPrefix(s, "<")
	s = strings.TrimSuffix(s, ">")
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSpace(s)
	// The name ends at the first whitespace or self-closing slash.
	end := len(s)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '/' {
			end = i
			break
		}
	}
	return strings.ToLower(s[:end])
}
