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
// always removed regardless of the allowed list. A stray '<' encountered while
// already inside a tag causes the previously buffered text to be flushed to the
// output as literal content and a new tag to begin, so malformed input degrades
// gracefully instead of swallowing text. An unterminated tag run at the end of
// the input (a trailing '<' with no closing '>') is discarded, matching
// striptags which drops trailing partial tags.
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
	for i := 0; i < len(html); i++ {
		c := html[i]
		switch state {
		case statePlaintext:
			if c == '<' {
				// Detect the start of an HTML comment.
				if strings.HasPrefix(html[i:], "<!--") {
					state = stateComment
					tag.Reset()
					tag.WriteString("<!--")
					i += 3
					continue
				}
				state = stateHTML
				tag.Reset()
				tag.WriteByte(c)
			} else {
				out.WriteByte(c)
			}
		case stateHTML:
			switch c {
			case '<':
				// A stray '<' inside what we thought was a tag: emit the
				// buffered text as plain content and restart the tag.
				out.WriteString(tag.String())
				tag.Reset()
				tag.WriteByte(c)
			case '>':
				tag.WriteByte(c)
				full := tag.String()
				if allowedSet[tagName(full)] {
					out.WriteString(full)
				}
				tag.Reset()
				state = statePlaintext
			default:
				tag.WriteByte(c)
			}
		case stateComment:
			tag.WriteByte(c)
			if c == '>' && strings.HasSuffix(tag.String(), "-->") {
				tag.Reset()
				state = statePlaintext
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
