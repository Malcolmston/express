// Package sanitizehtml sanitizes untrusted HTML by removing disallowed tags
// and attributes, mirroring a practical subset of the npm "sanitize-html"
// library. It is intended for the common case of accepting rich text from a
// user (comments, profile bios, CMS fields) and rendering it back into a page
// without opening the door to cross-site scripting or layout-breaking markup.
//
// Sanitization is driven by an allowlist rather than a blocklist, which is the
// safe default: only tags that appear in Options.AllowedTags survive, and on
// those surviving tags only the attributes named in Options.AllowedAttributes
// are kept. Everything else is discarded. Because the policy enumerates what is
// permitted instead of what is forbidden, a novel or misspelled tag is stripped
// automatically rather than slipping through. DefaultOptions returns a policy
// equivalent to the sanitize-html defaults, allowing common formatting tags
// plus href/name/target on <a> and the usual source attributes on <img>.
//
// The implementation tokenizes the input with the standard library only; no
// third-party HTML parser is used. The tokenizer scans the string once and
// emits a flat stream of text, start-tag, and end-tag tokens. HTML comments
// (<!-- ... -->) and declarations such as <!DOCTYPE html> are dropped during
// scanning, and a stray '<' that does not begin a valid tag is treated as
// literal text. Start tags are parsed into a tag name plus a list of
// attribute name/value pairs, with entity references in attribute values
// unescaped so that they can be re-escaped consistently on output.
//
// Disallowed tags are stripped in a way that preserves the reader's content:
// the tag's angle-bracket markup is removed but the text between the opening
// and closing tags is emitted verbatim, so "<foo>hello</foo>" becomes "hello".
// The two exceptions are <script> and <style>, whose entire raw contents are
// consumed and discarded rather than surfaced as text, preventing script
// source or CSS from leaking into the output. On surviving start tags each
// attribute is checked against AllowedAttributes (honoring the "*" key that
// applies to every tag), disallowed attributes are dropped, and the retained
// attribute values are re-escaped with html.EscapeString before serialization.
//
// Parity with the Node library is deliberately partial. This port covers the
// tag and attribute allowlist model, text-preserving tag removal, and the
// script/style content stripping that most callers rely on. It does not
// implement the richer sanitize-html features such as per-attribute value
// filtering, URL scheme validation, CSS style parsing, allowed classes,
// transformTags, or exclusiveFilter callbacks. Tag and attribute names are
// compared case-insensitively (lower-cased), matching the way browsers treat
// HTML, and output tag names are normalized to lower case.
package sanitizehtml

import (
	"html"
	"strings"
)

// Options controls how Sanitize filters HTML.
type Options struct {
	// AllowedTags is the set of tag names permitted in the output. A single
	// entry of "*" permits every tag. Tags not listed are stripped while their
	// text content is preserved.
	AllowedTags []string
	// AllowedAttributes maps a tag name to the attribute names permitted on
	// that tag. The special key "*" applies to all tags. Attributes not
	// permitted are removed from otherwise-allowed tags.
	AllowedAttributes map[string][]string
}

// DefaultOptions returns Options mirroring the sanitize-html defaults: a set of
// common formatting tags is allowed, href/name/target are permitted on <a>,
// and image source attributes are permitted on <img>.
func DefaultOptions() Options {
	return Options{
		AllowedTags: []string{
			"h1", "h2", "h3", "h4", "h5", "h6", "blockquote", "p", "a", "ul",
			"ol", "nl", "li", "b", "i", "strong", "em", "strike", "code", "hr",
			"br", "div", "table", "thead", "caption", "tbody", "tr", "th", "td",
			"pre", "span", "sub", "sup", "small", "u", "s", "abbr",
		},
		AllowedAttributes: map[string][]string{
			"a":   {"href", "name", "target"},
			"img": {"src", "srcset", "alt", "title", "width", "height"},
		},
	}
}

// tokenKind enumerates the kinds of tokens produced by the tokenizer.
type tokenKind int

const (
	tokenText tokenKind = iota
	tokenStart
	tokenEnd
)

// attribute is a single parsed HTML attribute with its unescaped value.
type attribute struct {
	name  string
	value string
}

// token is a single unit produced by the tokenizer.
type token struct {
	kind        tokenKind
	data        string // text content for tokenText
	name        string // tag name for tokenStart/tokenEnd
	attrs       []attribute
	selfClosing bool
}

// Sanitize returns htmlStr with disallowed tags and attributes removed
// according to opts. Text content of disallowed tags is kept; the contents of
// <script> and <style> elements are removed entirely.
func Sanitize(htmlStr string, opts Options) string {
	allowedTags := make(map[string]bool, len(opts.AllowedTags))
	allowAll := false
	for _, t := range opts.AllowedTags {
		if t == "*" {
			allowAll = true
		}
		allowedTags[strings.ToLower(t)] = true
	}

	tokens := tokenize(htmlStr)

	var b strings.Builder
	b.Grow(len(htmlStr))
	for _, tok := range tokens {
		switch tok.kind {
		case tokenText:
			b.WriteString(tok.data)
		case tokenStart:
			if allowAll || allowedTags[tok.name] {
				b.WriteString(serializeStart(tok, opts))
			}
		case tokenEnd:
			if allowAll || allowedTags[tok.name] {
				b.WriteString("</")
				b.WriteString(tok.name)
				b.WriteByte('>')
			}
		}
	}
	return b.String()
}

// serializeStart renders an allowed start tag, keeping only permitted
// attributes and re-escaping their values.
func serializeStart(tok token, opts Options) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(tok.name)
	for _, a := range tok.attrs {
		if !attrAllowed(tok.name, a.name, opts) {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(a.name)
		b.WriteString(`="`)
		b.WriteString(html.EscapeString(a.value))
		b.WriteByte('"')
	}
	if tok.selfClosing {
		b.WriteString(" />")
	} else {
		b.WriteByte('>')
	}
	return b.String()
}

// attrAllowed reports whether attribute attrName is permitted on tag according
// to opts, honoring the "*" wildcard key that applies to every tag.
func attrAllowed(tag, attrName string, opts Options) bool {
	if list, ok := opts.AllowedAttributes["*"]; ok {
		for _, a := range list {
			if strings.ToLower(a) == attrName {
				return true
			}
		}
	}
	if list, ok := opts.AllowedAttributes[tag]; ok {
		for _, a := range list {
			if strings.ToLower(a) == attrName {
				return true
			}
		}
	}
	return false
}

// tokenize scans s into a flat list of text, start-tag, and end-tag tokens.
// Comments and declarations are dropped. Script and style element contents are
// consumed and discarded.
func tokenize(s string) []token {
	var tokens []token
	i := 0
	for i < len(s) {
		if s[i] != '<' {
			j := strings.IndexByte(s[i:], '<')
			if j == -1 {
				tokens = append(tokens, token{kind: tokenText, data: s[i:]})
				break
			}
			tokens = append(tokens, token{kind: tokenText, data: s[i : i+j]})
			i += j
			continue
		}

		// s[i] == '<'
		if strings.HasPrefix(s[i:], "<!--") {
			end := strings.Index(s[i+4:], "-->")
			if end == -1 {
				i = len(s)
			} else {
				i = i + 4 + end + 3
			}
			continue
		}
		if i+1 < len(s) && s[i+1] == '!' {
			// Declaration such as <!DOCTYPE ...>.
			gt := strings.IndexByte(s[i:], '>')
			if gt == -1 {
				i = len(s)
			} else {
				i += gt + 1
			}
			continue
		}
		if i+1 < len(s) && s[i+1] == '/' {
			gt := strings.IndexByte(s[i:], '>')
			if gt == -1 {
				tokens = append(tokens, token{kind: tokenText, data: s[i:]})
				break
			}
			name := parseTagName(s[i+2 : i+gt])
			if name != "" {
				tokens = append(tokens, token{kind: tokenEnd, name: name})
			}
			i += gt + 1
			continue
		}
		if i+1 < len(s) && isLetter(s[i+1]) {
			tok, next, ok := parseStartTag(s, i)
			if ok {
				tokens = append(tokens, tok)
				i = next
				// Discard raw contents of script/style elements.
				if !tok.selfClosing && (tok.name == "script" || tok.name == "style") {
					rest := s[i:]
					closeIdx := indexCloseTag(rest, tok.name)
					if closeIdx == -1 {
						i = len(s)
					} else {
						gt := strings.IndexByte(rest[closeIdx:], '>')
						if gt == -1 {
							i = len(s)
						} else {
							tokens = append(tokens, token{kind: tokenEnd, name: tok.name})
							i += closeIdx + gt + 1
						}
					}
				}
				continue
			}
		}
		// A '<' that does not begin a tag is literal text.
		tokens = append(tokens, token{kind: tokenText, data: "<"})
		i++
	}
	return tokens
}

// indexCloseTag returns the index in s of the closing tag "</name" matched
// case-insensitively, or -1 if none is found.
func indexCloseTag(s, name string) int {
	needle := "</" + name
	return strings.Index(strings.ToLower(s), needle)
}

// parseStartTag parses a start tag beginning at s[start] ('<'). It returns the
// token, the index just past the tag, and whether parsing succeeded. Parsing
// fails (ok=false) for an unterminated tag, in which case the caller treats the
// '<' as literal text.
func parseStartTag(s string, start int) (token, int, bool) {
	i := start + 1
	nameStart := i
	for i < len(s) && isNameChar(s[i]) {
		i++
	}
	name := strings.ToLower(s[nameStart:i])

	tok := token{kind: tokenStart, name: name}
	for i < len(s) {
		for i < len(s) && isSpace(s[i]) {
			i++
		}
		if i >= len(s) {
			return token{}, 0, false
		}
		if s[i] == '>' {
			i++
			return tok, i, true
		}
		if s[i] == '/' {
			if i+1 < len(s) && s[i+1] == '>' {
				tok.selfClosing = true
				return tok, i + 2, true
			}
			i++
			continue
		}
		// Attribute name.
		attrStart := i
		for i < len(s) && !isSpace(s[i]) && s[i] != '=' && s[i] != '>' && s[i] != '/' {
			i++
		}
		attrName := strings.ToLower(s[attrStart:i])
		for i < len(s) && isSpace(s[i]) {
			i++
		}
		var value string
		if i < len(s) && s[i] == '=' {
			i++
			for i < len(s) && isSpace(s[i]) {
				i++
			}
			if i < len(s) && (s[i] == '"' || s[i] == '\'') {
				q := s[i]
				i++
				vStart := i
				for i < len(s) && s[i] != q {
					i++
				}
				value = s[vStart:i]
				if i < len(s) {
					i++
				}
			} else {
				vStart := i
				for i < len(s) && !isSpace(s[i]) && s[i] != '>' {
					i++
				}
				value = s[vStart:i]
			}
		}
		if attrName != "" {
			tok.attrs = append(tok.attrs, attribute{name: attrName, value: html.UnescapeString(value)})
		}
	}
	return token{}, 0, false
}

// parseTagName extracts a lower-cased tag name from raw tag inner text.
func parseTagName(s string) string {
	s = strings.TrimSpace(s)
	end := len(s)
	for i := 0; i < len(s); i++ {
		if !isNameChar(s[i]) {
			end = i
			break
		}
	}
	return strings.ToLower(s[:end])
}

// isLetter reports whether c is an ASCII letter.
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isNameChar reports whether c may appear in a tag or attribute name.
func isNameChar(c byte) bool {
	return isLetter(c) || (c >= '0' && c <= '9') || c == '-' || c == ':' || c == '_'
}

// isSpace reports whether c is HTML whitespace.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}
