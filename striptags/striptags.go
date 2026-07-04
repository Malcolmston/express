// Package striptags removes HTML tags from a string while keeping the text
// content, mirroring the behavior of the npm "striptags" library.
//
// By default every tag is stripped. When one or more allowed tag names are
// supplied, those tags (and their matching closing tags) are preserved while
// all other tags are removed. HTML comments are always stripped.
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
