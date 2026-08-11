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
// StripTagsWith exposes the third argument of the Node function,
// striptags(html, allowedTags, tagReplacement): the string written in place of
// every removed tag. StripTags is the same call with an empty replacement.
// Note that a replacement is emitted only for removed tags, not for removed
// comments, matching the reference implementation.
//
// NewStreamer is the port of striptags.init_streaming_mode: it returns a
// Streamer whose Write method may be called once per chunk, carrying the parser
// state across chunk boundaries so a tag split between two chunks is still
// recognised. StripTagsWith is one Write on a fresh Streamer.
//
// This port intentionally keeps the rest of the surface small. It does not
// decode or encode HTML entities and does not attempt to validate attribute
// contents; the text between tags, including any entities, is passed through
// unchanged. The allowed-tag matching is purely by tag name, so attributes on an
// allowed tag are preserved exactly as written.
//
// Because output is never escaped, striptags is a text-extraction utility and
// not a sanitizer: a '<' that survives as literal text survives verbatim. Use
// the sanitizehtml package when the result will be rendered as HTML.
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
//
// Each allowed entry may be a bare name ("p"), a bracketed name ("<p>"), or a
// single string listing several bracketed names ("<a><p>"), matching the string
// and array forms the Node original accepts for its allowable_tags argument.
func StripTags(html string, allowed ...string) string {
	return StripTagsWith(html, allowed, "")
}

// StripTagsWith removes HTML tags from html, writing tagReplacement in place of
// every removed tag. It mirrors the three-argument form of the Node function,
// striptags(html, allowedTags, tagReplacement). Removed HTML comments produce
// no replacement, matching the reference implementation.
func StripTagsWith(html string, allowed []string, tagReplacement string) string {
	return NewStreamer(allowed, tagReplacement).Write(html)
}

// Streamer strips tags from a sequence of chunks, carrying the parser state from
// one chunk to the next so that a tag split across a chunk boundary is still
// recognised. It is the port of striptags.init_streaming_mode(allowedTags,
// tagReplacement), which returns a closure over the same mutable state.
//
// A Streamer is not safe for concurrent use; make one per stream.
type Streamer struct {
	allowed     map[string]bool
	replacement string

	state   int
	tag     strings.Builder
	depth   int
	inQuote byte // 0 when not inside a quoted attribute value
}

// NewStreamer returns a Streamer that keeps the tags named in allowed (in any of
// the forms StripTags accepts) and writes replacement in place of every removed
// tag. It is the equivalent of striptags.init_streaming_mode.
func NewStreamer(allowed []string, tagReplacement string) *Streamer {
	set := make(map[string]bool, len(allowed))
	for _, entry := range allowed {
		for _, t := range normalizeAllowed(entry) {
			set[t] = true
		}
	}
	return &Streamer{allowed: set, replacement: tagReplacement, state: statePlaintext}
}

// Write feeds one chunk through the stripper and returns the output produced by
// that chunk. State left over at the end of the chunk — a tag or comment that
// has not been closed yet — is carried into the next call, so
// s.Write("<b") + s.Write(">x") keeps the same result as s.Write("<b>x").
func (s *Streamer) Write(html string) string {
	var out strings.Builder
	out.Grow(len(html))

	for i := 0; i < len(html); i++ {
		c := html[i]
		switch s.state {
		case statePlaintext:
			if c == '<' {
				s.state = stateHTML
				s.tag.WriteByte(c)
			} else {
				out.WriteByte(c)
			}
		case stateHTML:
			switch c {
			case '<':
				// Ignore '<' inside a quoted value; otherwise it is a
				// nested delimiter that raises the depth (matched by a
				// later '>') rather than closing the tag.
				if s.inQuote != 0 {
					break
				}
				s.depth++
			case '>':
				if s.inQuote != 0 {
					break
				}
				if s.depth > 0 {
					s.depth--
					break
				}
				s.tag.WriteByte('>')
				full := s.tag.String()
				if s.allowed[tagName(full)] {
					out.WriteString(full)
				} else {
					out.WriteString(s.replacement)
				}
				s.tag.Reset()
				s.inQuote = 0
				s.state = statePlaintext
			case '"', '\'':
				if c == s.inQuote {
					s.inQuote = 0
				} else if s.inQuote == 0 {
					s.inQuote = c
				}
				s.tag.WriteByte(c)
			case '-':
				// "<!-" followed by '-' begins an HTML comment.
				if s.tag.String() == "<!-" {
					s.state = stateComment
				}
				s.tag.WriteByte(c)
			case ' ', '\n':
				// A bare "<" followed by whitespace is not a tag: emit the
				// "< " literally and return to plaintext.
				if s.tag.String() == "<" {
					out.WriteString("< ")
					s.tag.Reset()
					s.state = statePlaintext
					break
				}
				s.tag.WriteByte(c)
			default:
				s.tag.WriteByte(c)
			}
		case stateComment:
			if c == '>' {
				// Close only when the two characters before '>' are "--".
				if t := s.tag.String(); len(t) >= 2 && t[len(t)-2:] == "--" {
					s.state = statePlaintext
				}
				s.tag.Reset()
			} else {
				s.tag.WriteByte(c)
			}
		}
	}

	// Any unterminated '<' run stays buffered: in the one-shot StripTagsWith
	// case that means it is discarded, matching striptags which drops trailing
	// partial tags; in streaming use the next chunk continues it.
	return out.String()
}

// normalizeAllowed turns one allowed-tag specifier into the lower-cased tag
// names it denotes. It accepts a bare name ("p"), a bracketed name ("<p>" or
// "</p>"), and a single string listing several bracketed names ("<a><p>"),
// covering both the string and the array form of the Node original's
// allowable_tags argument. Entries that name no tag yield nothing.
func normalizeAllowed(t string) []string {
	t = strings.TrimSpace(t)
	if t == "" {
		return nil
	}
	if !strings.Contains(t, "<") {
		return []string{strings.ToLower(strings.TrimSuffix(t, ">"))}
	}

	// Extract every "<name>" group, mirroring upstream's /<(\w*)>/g scan.
	var names []string
	for i := 0; i < len(t); i++ {
		if t[i] != '<' {
			continue
		}
		j := strings.IndexByte(t[i:], '>')
		if j < 0 {
			break
		}
		name := strings.TrimSpace(t[i+1 : i+j])
		name = strings.TrimPrefix(name, "/")
		name = strings.TrimSuffix(name, "/")
		if name != "" {
			names = append(names, strings.ToLower(strings.TrimSpace(name)))
		}
		i += j
	}
	return names
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
