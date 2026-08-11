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
// # URL attributes
//
// URL-bearing attributes are additionally checked against a scheme allowlist.
// The attributes named in Options.AllowedSchemesAppliedToAttributes must either
// be scheme-less (relative) or use one of Options.AllowedSchemes (by default
// http, https, ftp, mailto and tel). That list defaults to the full upstream
// set — href, src and cite plus the other historical URL sinks such as action,
// formaction, poster, xlink:href, background, longdesc, usemap, data, ping,
// dynsrc and lowsrc — because an attribute missing from it is an unchecked
// "javascript:" sink the moment a policy permits it. The srcset and imagesrcset
// attributes are filtered candidate by candidate on every tag. This is what
// stops "javascript:", "data:" and "vbscript:" URLs from surviving. Matching
// mirrors upstream's naughtyHref check: all characters in the range
// U+0000..U+0020 are stripped and HTML comments are removed before the scheme
// is extracted, so "jav&#9;ascript:" and "  JaVaScRiPt:" are both rejected.
// A nil AllowedSchemes means "use the defaults"; an explicitly empty (non-nil)
// slice rejects every scheme, and a slice containing "*" accepts every scheme.
//
// # Parsing
//
// The implementation tokenizes the input with the standard library only; no
// third-party HTML parser is used. The tokenizer scans the string once and emits
// a flat stream of text, start-tag and end-tag tokens. HTML comments
// (<!-- ... -->) and declarations such as <!DOCTYPE html> are dropped during
// scanning, and a stray '<' that does not begin a valid tag is treated as
// literal text. Start tags are parsed into a tag name plus a list of attribute
// name/value pairs, with entity references in attribute values unescaped so
// that they can be re-escaped consistently on output. Text nodes are likewise
// entity-decoded and then re-escaped, matching upstream's decodeEntities parser
// default, so "&lt;Kapow!&gt;" round-trips unchanged while a raw "<" in text
// becomes "&lt;".
//
// The HTML raw-text elements are recognized as such, because their content is
// text rather than markup and must never be re-emitted as live markup:
// <script>, <style>, <xmp>, <iframe>, <noembed>, <noframes> and <plaintext>
// hold raw text (no entity decoding), and <title> and <textarea> hold escapable
// raw text (entities are decoded). Output tags are balanced: a close tag that
// does not match the innermost open tag is dropped, and any tag still open at
// the end of the input is closed.
//
// # Disallowed tags
//
// Disallowed tags are handled according to Options.DisallowedTagsMode. The
// default, "discard", removes the tag's angle-bracket markup but emits the text
// between the opening and closing tags, so "<foo>hello</foo>" becomes "hello".
// "escape" renders the disallowed tag itself as visible text, "recursiveEscape"
// does the same for the tag and everything nested inside it, and
// "completelyDiscard" drops the tag together with all of its content. Any
// unrecognized mode behaves like "escape", as upstream's branch structure does.
// The exceptions in every mode are the non-text elements listed in
// Options.NonTextTags (<script>, <style>, <textarea>, <option> and <xmp> by
// default), whose text content is dropped along with the tag rather than
// surfaced as text, preventing script source, CSS or form-control text from
// leaking into the output.
//
// # Attributes
//
// On surviving start tags each attribute is checked against AllowedAttributes
// (honoring the "*" key that applies to every tag, and treating an entry
// containing "*" as an anchored glob such as "data-*"), disallowed attributes
// are dropped, the class attribute is filtered against Options.AllowedClasses
// when that option is set, and the retained attribute values are re-escaped
// before serialization. An attribute whose value is empty is dropped when it is
// a known non-boolean attribute (Options.NonBooleanAttributes) that is not
// listed in Options.AllowedEmptyAttributes; a boolean attribute is emitted as a
// bare name with no value. Tags named in Options.SelfClosing are emitted in the
// self-closing "<br />" form even when the input wrote them as "<br>".
//
// # Parity notes
//
// This port covers the tag/attribute allowlist model, URL scheme validation,
// srcset filtering, class filtering, the four disallowedTagsMode behaviors,
// configurable nonTextTags, selfClosing, allowedEmptyAttributes and
// nonBooleanAttributes lists, and textFilter. It does not implement
// transformTags, exclusiveFilter, allowedStyles CSS parsing,
// allowedIframeHostnames, allowedScriptHostnames, nestingLimit or
// enforceHtmlBoundary. Tag and attribute names are compared case-insensitively
// (lower-cased), matching the way browsers treat HTML, and output tag names are
// normalized to lower case. Differences that are deliberate rather than
// incidental are listed in the repository's API-DEVIATIONS.md.
package sanitizehtml

import (
	"html"
	"strings"
)

// Disallowed-tag handling modes accepted by Options.DisallowedTagsMode. They
// mirror the upstream sanitize-html `disallowedTagsMode` option values.
const (
	// ModeDiscard removes the markup of a disallowed tag but keeps the text
	// between its opening and closing tags. This is the default.
	ModeDiscard = "discard"
	// ModeCompletelyDiscard removes a disallowed tag together with all of its
	// content.
	ModeCompletelyDiscard = "completelyDiscard"
	// ModeEscape renders a disallowed tag as escaped, visible text while still
	// sanitizing its content.
	ModeEscape = "escape"
	// ModeRecursiveEscape renders a disallowed tag and everything nested inside
	// it as escaped, visible text.
	ModeRecursiveEscape = "recursiveEscape"
)

// DefaultSchemes lists the URL schemes permitted by default in URL-bearing
// attributes, matching sanitize-html's defaults.allowedSchemes.
var DefaultSchemes = []string{"http", "https", "ftp", "mailto", "tel"}

// DefaultSchemeAttributes lists the attributes whose values are checked against
// the scheme allowlist, matching sanitize-html's
// defaults.allowedSchemesAppliedToAttributes. Every attribute absent from this
// list is an unchecked URL sink as soon as a policy allows it, which is why the
// list is deliberately broad rather than just href/src/cite.
var DefaultSchemeAttributes = []string{
	"href", "src", "cite",
	"action", "formaction", "data", "xlink:href",
	"poster", "background", "ping",
	"longdesc", "usemap", "codebase", "classid", "archive",
	"profile", "manifest", "itemid",
	"dynsrc", "lowsrc",
}

// DefaultNonTextTags lists the elements whose textual content is dropped along
// with the tag when the tag is disallowed, matching sanitize-html's nonTextTags
// default.
var DefaultNonTextTags = []string{"script", "style", "textarea", "option", "xmp"}

// DefaultSelfClosing lists the elements serialized in the "<br />" form,
// matching sanitize-html's defaults.selfClosing.
var DefaultSelfClosing = []string{"img", "br", "hr", "area", "base", "basefont", "input", "link", "meta", "col"}

// DefaultAllowedEmptyAttributes lists the attributes that survive with an empty
// value, matching sanitize-html's defaults.allowedEmptyAttributes.
var DefaultAllowedEmptyAttributes = []string{"alt"}

// DefaultNonBooleanAttributes lists the attributes that cannot be boolean, taken
// from sanitize-html's defaults.nonBooleanAttributes (itself derived from the
// HTML standard's attribute index). An empty value on one of these is dropped
// rather than emitted, unless the attribute is in AllowedEmptyAttributes.
var DefaultNonBooleanAttributes = []string{
	"abbr", "accept", "accept-charset", "accesskey", "action",
	"allow", "alt", "as", "autocapitalize", "autocomplete",
	"blocking", "charset", "cite", "class", "color", "cols",
	"colspan", "content", "contenteditable", "coords", "crossorigin",
	"data", "datetime", "decoding", "dir", "dirname", "download",
	"draggable", "enctype", "enterkeyhint", "fetchpriority", "for",
	"form", "formaction", "formenctype", "formmethod", "formtarget",
	"headers", "height", "hidden", "high", "href", "hreflang",
	"http-equiv", "id", "imagesizes", "imagesrcset", "inputmode",
	"integrity", "is", "itemid", "itemprop", "itemref", "itemtype",
	"kind", "label", "lang", "list", "loading", "low", "max",
	"maxlength", "media", "method", "min", "minlength", "name",
	"nonce", "optimum", "pattern", "ping", "placeholder", "popover",
	"popovertarget", "popovertargetaction", "poster", "preload",
	"referrerpolicy", "rel", "rows", "rowspan", "sandbox", "scope",
	"shape", "size", "sizes", "slot", "span", "spellcheck", "src",
	"srcdoc", "srclang", "srcset", "start", "step", "style",
	"tabindex", "target", "title", "translate", "type", "usemap",
	"value", "width", "wrap",
	// Event handlers
	"onauxclick", "onafterprint", "onbeforematch", "onbeforeprint",
	"onbeforeunload", "onbeforetoggle", "onblur", "oncancel",
	"oncanplay", "oncanplaythrough", "onchange", "onclick", "onclose",
	"oncontextlost", "oncontextmenu", "oncontextrestored", "oncopy",
	"oncuechange", "oncut", "ondblclick", "ondrag", "ondragend",
	"ondragenter", "ondragleave", "ondragover", "ondragstart",
	"ondrop", "ondurationchange", "onemptied", "onended",
	"onerror", "onfocus", "onformdata", "onhashchange", "oninput",
	"oninvalid", "onkeydown", "onkeypress", "onkeyup",
	"onlanguagechange", "onload", "onloadeddata", "onloadedmetadata",
	"onloadstart", "onmessage", "onmessageerror", "onmousedown",
	"onmouseenter", "onmouseleave", "onmousemove", "onmouseout",
	"onmouseover", "onmouseup", "onoffline", "ononline", "onpagehide",
	"onpageshow", "onpaste", "onpause", "onplay", "onplaying",
	"onpopstate", "onprogress", "onratechange", "onreset", "onresize",
	"onrejectionhandled", "onscroll", "onscrollend",
	"onsecuritypolicyviolation", "onseeked", "onseeking", "onselect",
	"onslotchange", "onstalled", "onstorage", "onsubmit", "onsuspend",
	"ontimeupdate", "ontoggle", "onunhandledrejection", "onunload",
	"onvolumechange", "onwaiting", "onwheel",
}

// rawTextElements are the HTML elements whose content the tokenizer must treat
// as raw text rather than markup, and in which character references are NOT
// decoded. Their content is captured verbatim up to the matching close tag.
var rawTextElements = map[string]bool{
	"script":    true,
	"style":     true,
	"xmp":       true,
	"iframe":    true,
	"noembed":   true,
	"noframes":  true,
	"plaintext": true,
}

// escapableRawTextElements are the elements whose content is text in which
// character references ARE decoded.
var escapableRawTextElements = map[string]bool{
	"title":    true,
	"textarea": true,
}

// voidElements are the HTML elements that cannot have content, so an HTML parser
// closes them the moment they open. This is a property of the language, not of
// the policy: Options.SelfClosing only decides how an element is serialized, and
// naming a normal element there does not stop it from having children.
var voidElements = map[string]bool{
	"area": true, "base": true, "basefont": true, "br": true, "col": true,
	"embed": true, "frame": true, "hr": true, "img": true, "input": true,
	"isindex": true, "keygen": true, "link": true, "meta": true, "param": true,
	"source": true, "track": true, "wbr": true,
}

// srcsetAttributes are the attributes whose value is a srcset candidate list,
// filtered URL by URL rather than checked as a single URL.
var srcsetAttributes = map[string]bool{
	"srcset":      true,
	"imagesrcset": true,
}

// Options controls how Sanitize filters HTML.
type Options struct {
	// AllowedTags is the set of tag names permitted in the output. A single
	// entry of "*" permits every tag. Tags not listed are handled according to
	// DisallowedTagsMode.
	AllowedTags []string
	// AllowedAttributes maps a tag name to the attribute names permitted on
	// that tag. The special key "*" applies to all tags. Attributes not
	// permitted are removed from otherwise-allowed tags. An attribute name may
	// contain "*", which matches any run of characters, e.g. "data-*".
	AllowedAttributes map[string][]string
	// AllowedClasses maps a tag name to the class names permitted in that tag's
	// class attribute; the special key "*" applies to all tags, and a class
	// entry containing "*" is a glob. Naming a tag here implicitly permits its
	// class attribute, as upstream does. When AllowedClasses is nil the class
	// attribute is not filtered beyond the normal attribute allowlist.
	AllowedClasses map[string][]string
	// AllowedSchemes lists the URL schemes accepted in the attributes named by
	// AllowedSchemesAppliedToAttributes. A nil slice means DefaultSchemes; an
	// empty non-nil slice rejects every scheme; a slice containing "*" accepts
	// every scheme. Scheme-less (relative) URLs are always accepted.
	AllowedSchemes []string
	// AllowedSchemesByTag overrides AllowedSchemes for specific tags, keyed by
	// tag name.
	AllowedSchemesByTag map[string][]string
	// AllowedSchemesAppliedToAttributes names the attributes whose values are
	// scheme-checked. A nil slice means DefaultSchemeAttributes. The srcset and
	// imagesrcset attributes are always checked candidate-by-candidate on every
	// tag, matching upstream, which keys that check on the attribute name alone.
	AllowedSchemesAppliedToAttributes []string
	// AllowProtocolRelative controls whether protocol-relative URLs ("//host/x")
	// are permitted. A nil pointer means true, the upstream default.
	AllowProtocolRelative *bool
	// NonTextTags names the elements whose text content is dropped with the tag
	// rather than surfaced as text when the tag is disallowed. A nil slice means
	// DefaultNonTextTags.
	NonTextTags []string
	// SelfClosing names the elements serialized in the "<br />" form. A nil
	// slice means DefaultSelfClosing.
	SelfClosing []string
	// AllowedEmptyAttributes names the attributes that survive with an empty
	// value. A nil slice means DefaultAllowedEmptyAttributes.
	AllowedEmptyAttributes []string
	// NonBooleanAttributes names the attributes that cannot be boolean, so that
	// an empty value on one of them is dropped. A nil slice means
	// DefaultNonBooleanAttributes; an entry of "*" applies to every attribute.
	NonBooleanAttributes []string
	// DisallowedTagsMode selects how disallowed tags are rendered. It accepts
	// ModeDiscard (the default when empty), ModeCompletelyDiscard, ModeEscape,
	// and ModeRecursiveEscape. Any other value behaves like ModeEscape.
	DisallowedTagsMode string
	// TextFilter, when non-nil, is called with each already-escaped text node
	// and the name of its enclosing tag ("" at top level). Its return value is
	// written to the output in place of the text. As upstream, it is not applied
	// to the content of a raw-text element that survives sanitization.
	TextFilter func(text, tagName string) string
}

// DefaultOptions returns Options mirroring the sanitize-html defaults: a set of
// common formatting tags is allowed, href/name/target are permitted on <a>,
// image source attributes are permitted on <img>, and URL-bearing attributes
// are restricted to the http, https, ftp, mailto, and tel schemes.
func DefaultOptions() Options {
	return Options{
		// Mirrors sanitizeHtml.defaults.allowedTags in the upstream index.js
		// (categories derived from MDN element groups).
		AllowedTags: []string{
			"address", "article", "aside", "footer", "header",
			"h1", "h2", "h3", "h4", "h5", "h6", "hgroup",
			"main", "nav", "section",
			"blockquote", "dd", "div", "dl", "dt", "figcaption", "figure",
			"hr", "li", "menu", "ol", "p", "pre", "ul",
			"a", "abbr", "b", "bdi", "bdo", "br", "cite", "code", "data", "dfn",
			"em", "i", "kbd", "mark", "q",
			"rb", "rp", "rt", "rtc", "ruby",
			"s", "samp", "small", "span", "strong", "sub", "sup", "time", "u", "var", "wbr",
			"caption", "col", "colgroup", "table", "tbody", "td", "tfoot", "th",
			"thead", "tr",
		},
		AllowedAttributes: map[string][]string{
			"a":   {"href", "name", "target"},
			"img": {"src", "srcset", "alt", "title", "width", "height", "loading"},
		},
		AllowedSchemes:                    append([]string(nil), DefaultSchemes...),
		AllowedSchemesByTag:               map[string][]string{},
		AllowedSchemesAppliedToAttributes: append([]string(nil), DefaultSchemeAttributes...),
		NonTextTags:                       append([]string(nil), DefaultNonTextTags...),
		SelfClosing:                       append([]string(nil), DefaultSelfClosing...),
		AllowedEmptyAttributes:            append([]string(nil), DefaultAllowedEmptyAttributes...),
		NonBooleanAttributes:              append([]string(nil), DefaultNonBooleanAttributes...),
		DisallowedTagsMode:                ModeDiscard,
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
	selfClosing bool   // the source wrote "<x/>"
	rawText     bool   // start tag is a raw-text or escapable-raw-text element
	rawDecode   bool   // rawText content has its character references decoded
	rawContent  string // raw inner content captured for a rawText start tag
}

// policy is the resolved, lookup-friendly form of Options built once per call.
type policy struct {
	opts        Options
	allowAllTag bool
	tags        map[string]bool
	nonText     map[string]bool
	selfClosing map[string]bool
	schemeAttrs map[string]bool
	emptyOK     map[string]bool
	nonBoolean  map[string]bool
	anyNonBool  bool
	schemes     map[string]bool
	anyScheme   bool
	protoRel    bool
	mode        string
}

// newPolicy resolves opts into a policy, applying the upstream defaults for any
// field left nil.
func newPolicy(opts Options) *policy {
	p := &policy{opts: opts, mode: opts.DisallowedTagsMode, protoRel: true}
	if p.mode == "" {
		p.mode = ModeDiscard
	}
	if opts.AllowProtocolRelative != nil {
		p.protoRel = *opts.AllowProtocolRelative
	}

	p.tags = make(map[string]bool, len(opts.AllowedTags))
	for _, t := range opts.AllowedTags {
		if t == "*" {
			p.allowAllTag = true
		}
		p.tags[strings.ToLower(t)] = true
	}

	p.nonText = toSet(opts.NonTextTags, DefaultNonTextTags)
	p.selfClosing = toSet(opts.SelfClosing, DefaultSelfClosing)
	p.schemeAttrs = toSet(opts.AllowedSchemesAppliedToAttributes, DefaultSchemeAttributes)
	p.emptyOK = toSet(opts.AllowedEmptyAttributes, DefaultAllowedEmptyAttributes)
	p.nonBoolean = toSet(opts.NonBooleanAttributes, DefaultNonBooleanAttributes)
	p.anyNonBool = p.nonBoolean["*"]

	schemes := opts.AllowedSchemes
	if schemes == nil {
		schemes = DefaultSchemes
	}
	p.schemes = make(map[string]bool, len(schemes))
	for _, s := range schemes {
		if s == "*" {
			p.anyScheme = true
		}
		p.schemes[strings.ToLower(s)] = true
	}
	return p
}

// toSet lower-cases vals into a set, substituting def when vals is nil.
func toSet(vals, def []string) map[string]bool {
	if vals == nil {
		vals = def
	}
	m := make(map[string]bool, len(vals))
	for _, v := range vals {
		m[strings.ToLower(v)] = true
	}
	return m
}

// tagAllowed reports whether name survives sanitization.
func (p *policy) tagAllowed(name string) bool {
	return p.allowAllTag || p.tags[name]
}

// discarding reports whether the mode keeps disallowed tags' text (discard) or
// drops the whole subtree (completelyDiscard), as opposed to escaping it. The
// raw-text content rules upstream applies are gated on this same test.
func (p *policy) discarding() bool {
	return p.mode == ModeDiscard || p.mode == ModeCompletelyDiscard
}

// frame is one open element, tracked so that close tags can be matched, unclosed
// tags can be closed at the end, and TextFilter can be told its parent tag.
type frame struct {
	name    string
	allowed bool
	emitted bool // a plain end tag must be written when this frame closes
	escaped bool // the tag was rendered as escaped text, so its end tag is too
}

// Sanitize returns htmlStr with disallowed tags and attributes removed
// according to opts. By default the text content of disallowed tags is kept and
// the contents of disallowed non-text elements (<script>, <style>, <textarea>,
// <option>, <xmp>) are removed entirely; see Options.DisallowedTagsMode for the
// other behaviors. Sanitize never panics, and returns "" for empty input.
func Sanitize(htmlStr string, opts Options) string {
	p := newPolicy(opts)
	tokens := tokenize(htmlStr, p)

	var b strings.Builder
	b.Grow(len(htmlStr))

	// stack holds every currently open element, so that TextFilter can be told
	// which element a text node belongs to, close tags can be matched against
	// the innermost open element, and anything still open at the end of the
	// input can be closed.
	var stack []frame
	// escapeRoot is the stack index of the frame that opened a recursiveEscape
	// region, or -1 when we are not inside one. Everything nested inside that
	// region is rendered as text.
	escapeRoot := -1
	// skipDepth > 0 means we are inside a subtree that is being discarded whole.
	skipDepth := 0

	inEscape := func() bool { return escapeRoot >= 0 }
	parent := func() string {
		if len(stack) == 0 {
			return ""
		}
		return stack[len(stack)-1].name
	}

	// writeText emits an ordinary text node: decoded, re-escaped, then passed
	// through TextFilter. This is upstream's final `else` branch.
	writeText := func(raw string, decode bool) {
		if decode {
			raw = html.UnescapeString(raw)
		}
		text := escapeText(raw)
		if p.opts.TextFilter != nil {
			text = p.opts.TextFilter(text, parent())
		}
		b.WriteString(text)
	}

	// escapedTag renders a disallowed tag, and any raw-text content it holds, as
	// visible text.
	escapedTag := func(tok token) {
		b.WriteString(escapeText(serializeStart(tok, p)))
		if tok.rawText {
			b.WriteString(escapeText(decodeIf(tok.rawContent, tok.rawDecode)))
			b.WriteString(escapeText("</" + tok.name + ">"))
		}
	}

	for _, tok := range tokens {
		switch tok.kind {
		case tokenText:
			if skipDepth > 0 {
				continue
			}
			if inEscape() {
				b.WriteString(escapeText(html.UnescapeString(tok.data)))
				continue
			}
			if p.mode == ModeCompletelyDiscard && !p.tagAllowed(parent()) {
				// Upstream blanks the text of any element that is not itself
				// allowed in this mode, and at top level there is no element.
				continue
			}
			writeText(tok.data, true)

		case tokenStart:
			// A solidus in a start tag does not close a non-void element: an
			// HTML tokenizer ignores it. Whether the element can have content at
			// all is a property of the language, not of the policy.
			void := voidElements[tok.name]
			allowed := p.tagAllowed(tok.name)

			if skipDepth > 0 {
				if !void && !tok.rawText {
					skipDepth++
				}
				continue
			}
			if inEscape() {
				escapedTag(tok)
				if !void && !tok.rawText {
					stack = append(stack, frame{name: tok.name, allowed: allowed, escaped: true})
				}
				continue
			}

			if allowed {
				b.WriteString(serializeStart(tok, p))
				switch {
				case tok.rawText:
					b.WriteString(p.rawTextContent(tok))
					b.WriteString("</" + tok.name + ">")
				case void:
					// A void element is closed the moment it opens. When the
					// policy lists it in SelfClosing the "<br />" form already
					// said so; otherwise an explicit end tag follows at once.
					if !p.selfClosing[tok.name] {
						b.WriteString("</" + tok.name + ">")
					}
				default:
					// An element the policy lists in SelfClosing is written in
					// the "<x />" form and gets no end tag, but it is still a
					// normal element as far as parsing is concerned.
					stack = append(stack, frame{
						name:    tok.name,
						allowed: true,
						emitted: !p.selfClosing[tok.name],
					})
				}
				continue
			}

			// Disallowed start tag.
			switch {
			case p.discarding():
				if tok.rawText {
					// The tag is gone either way. Its content is dropped when
					// the element is a non-text element or the whole subtree is
					// being discarded, and surfaced as escaped text otherwise.
					if !p.nonText[tok.name] && p.mode != ModeCompletelyDiscard {
						stack = append(stack, frame{name: tok.name})
						writeText(tok.rawContent, tok.rawDecode)
						stack = stack[:len(stack)-1]
					}
					continue
				}
				if p.mode == ModeCompletelyDiscard {
					if !void {
						skipDepth = 1
					}
					continue
				}
				if p.nonText[tok.name] {
					// Not a raw-text element (<option>): its text is dropped,
					// but the tag's markup was going anyway.
					if !void {
						skipDepth = 1
					}
					continue
				}
				if !void {
					stack = append(stack, frame{name: tok.name})
				}

			default: // ModeEscape, ModeRecursiveEscape and any unrecognized mode
				escapedTag(tok)
				if void || tok.rawText {
					continue
				}
				stack = append(stack, frame{name: tok.name, escaped: true})
				if p.mode == ModeRecursiveEscape {
					escapeRoot = len(stack) - 1
				}
			}

		case tokenEnd:
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if len(stack) == 0 || stack[len(stack)-1].name != tok.name {
				// A close tag that does not match the innermost open element is
				// bad markup and is dropped. HTML makes two exceptions, which a
				// parser turns into elements: a stray "</p>" becomes an empty
				// <p> and a stray "</br>" becomes a <br>.
				if !inEscape() && (tok.name == "p" || tok.name == "br") {
					synth := token{kind: tokenStart, name: tok.name}
					switch {
					case p.tagAllowed(tok.name):
						b.WriteString(serializeStart(synth, p))
						if !p.selfClosing[tok.name] {
							b.WriteString("</" + tok.name + ">")
						}
					case !p.discarding():
						b.WriteString(escapeText(serializeStart(synth, p)))
						if !p.selfClosing[tok.name] {
							b.WriteString(escapeText("</" + tok.name + ">"))
						}
					}
				}
				continue
			}
			f := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if escapeRoot == len(stack) {
				escapeRoot = -1
			}
			switch {
			case f.emitted:
				b.WriteString("</" + f.name + ">")
			case f.escaped:
				b.WriteString(escapeText("</" + f.name + ">"))
			}
		}
	}

	// Close whatever is still open, innermost first. An IMPLIED close of a
	// disallowed tag that was rendered as text produces nothing, matching
	// upstream: only an end tag that was really written in the source is
	// escaped and echoed.
	for i := len(stack) - 1; i >= 0; i-- {
		f := stack[i]
		switch {
		case f.emitted:
			b.WriteString("</" + f.name + ">")
		case f.escaped && f.allowed:
			b.WriteString(escapeText("</" + f.name + ">"))
		}
	}
	return b.String()
}

// rawTextContent renders the captured content of an allowed raw-text element.
// The three rules are upstream's: <script> and <style> content is emitted
// verbatim (escaping it would break the code, and allowing those tags is already
// a decision to trust their content), <xmp> content has only its angle brackets
// escaped because entities in it are not decoded, and everything else — notably
// <textarea> — is decoded and then fully escaped so that no '<' can reopen a tag
// when a browser re-parses the output.
func (p *policy) rawTextContent(tok token) string {
	if p.discarding() {
		switch tok.name {
		case "script", "style":
			return tok.rawContent
		case "xmp":
			return escapeAngles(tok.rawContent)
		}
	}
	return escapeText(decodeIf(tok.rawContent, tok.rawDecode))
}

// decodeIf resolves character references when the element's content is
// escapable raw text.
func decodeIf(s string, decode bool) string {
	if decode {
		return html.UnescapeString(s)
	}
	return s
}

// escapeText escapes the three characters that are significant in HTML text
// content, matching upstream's escapeHtml(s, false).
func escapeText(s string) string {
	if !strings.ContainsAny(s, "&<>") {
		return s
	}
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// escapeAttr escapes an attribute value, matching upstream's
// escapeHtml(s, true). The apostrophe is deliberately not escaped: the value is
// always written inside double quotes.
func escapeAttr(s string) string {
	if !strings.ContainsAny(s, "&<>\"") {
		return s
	}
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// escapeAngles escapes only '<' and '>', leaving '&' alone so that content whose
// entities were never decoded is not double-encoded.
func escapeAngles(s string) string {
	if !strings.ContainsAny(s, "<>") {
		return s
	}
	return strings.NewReplacer("<", "&lt;", ">", "&gt;").Replace(s)
}

// serializeStart renders a start tag, keeping only permitted attributes and
// re-escaping their values. It is used both for output and, in the escape modes,
// to build the text that a disallowed tag is rendered as — upstream renders the
// normalized tag in both cases, not the original source.
func serializeStart(tok token, p *policy) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(tok.name)
	for _, a := range tok.attrs {
		if !attrNameSafe(a.name) || !attrAllowed(tok.name, a.name, p.opts) {
			continue
		}
		value := a.value
		if value == "" && !p.emptyOK[a.name] && (p.anyNonBool || p.nonBoolean[a.name]) {
			// A known non-boolean attribute with no value carries no meaning
			// and is dropped, as upstream does.
			continue
		}
		if a.name == "class" && p.opts.AllowedClasses != nil {
			value = filterClasses(tok.name, value, p.opts.AllowedClasses)
			if value == "" {
				continue
			}
		}
		if srcsetAttributes[a.name] {
			value = filterSrcset(value, a.name, p)
			if value == "" {
				continue
			}
		} else if p.schemeAttrs[a.name] && naughtyHref(tok.name, value, p) {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(a.name)
		switch {
		case value != "":
			b.WriteString(`="`)
			b.WriteString(escapeAttr(value))
			b.WriteByte('"')
		case p.emptyOK[a.name]:
			b.WriteString(`=""`)
		default:
			// A boolean attribute is emitted as a bare name.
		}
	}
	// Only the SelfClosing option decides the serialization; a solidus in the
	// source does not, matching upstream.
	if p.selfClosing[tok.name] {
		b.WriteString(" />")
	} else {
		b.WriteByte('>')
	}
	return b.String()
}

// attrNameSafe reports whether attrName can be serialized inside a start tag
// without altering its structure. The permissive tokenizer accepts an attribute
// name containing a '<', a double quote, a single quote or a backtick (browsers
// append those characters to the name with a parse error rather than ending it),
// and such a name would then be written back out verbatim. Every character that
// could terminate the name, the value, or the tag itself is rejected here so
// that a policy using a glob such as "data-*" cannot be tricked into echoing raw
// markup. This is stricter than upstream, whose check permits a quote and a
// backtick in an attribute name.
func attrNameSafe(attrName string) bool {
	if attrName == "" {
		return false
	}
	for i := 0; i < len(attrName); i++ {
		switch c := attrName[i]; {
		case isNameChar(c), c == '.':
		default:
			return false
		}
	}
	return true
}

// attrAllowed reports whether attribute attrName is permitted on tag according
// to opts, honoring the "*" wildcard key that applies to every tag, glob
// entries, and the implicit "class" permission granted by AllowedClasses.
func attrAllowed(tag, attrName string, opts Options) bool {
	if attrName == "class" && opts.AllowedClasses != nil {
		if _, ok := opts.AllowedClasses[tag]; ok {
			return true
		}
		if _, ok := opts.AllowedClasses["*"]; ok {
			return true
		}
	}
	if matchesNameList(attrName, opts.AllowedAttributes["*"]) {
		return true
	}
	return matchesNameList(attrName, opts.AllowedAttributes[tag])
}

// matchesNameList reports whether name matches any entry of list, where an entry
// containing "*" is an anchored glob: upstream compiles those entries into
// ^(...)$ with each "*" turned into ".*".
func matchesNameList(name string, list []string) bool {
	for _, a := range list {
		if globMatch(strings.ToLower(a), name) {
			return true
		}
	}
	return false
}

// globMatch reports whether name matches pattern, where "*" in pattern matches
// any run of characters and the match must cover the whole name.
func globMatch(pattern, name string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == name
	}
	parts := strings.Split(pattern, "*")
	// The first and last segments are anchored.
	if !strings.HasPrefix(name, parts[0]) {
		return false
	}
	name = name[len(parts[0]):]
	last := parts[len(parts)-1]
	for _, seg := range parts[1 : len(parts)-1] {
		i := strings.Index(name, seg)
		if i < 0 {
			return false
		}
		name = name[i+len(seg):]
	}
	return strings.HasSuffix(name, last) && len(name) >= len(last)
}

// filterClasses keeps only the class names permitted for tag, returning the
// space-joined survivors (empty when none survive).
func filterClasses(tag, value string, allowed map[string][]string) string {
	list := append(append([]string(nil), allowed["*"]...), allowed[tag]...)
	var kept []string
	for _, c := range strings.Fields(value) {
		for _, a := range list {
			if globMatch(a, c) {
				kept = append(kept, c)
				break
			}
		}
	}
	return strings.Join(kept, " ")
}

// filterSrcset drops the candidates of a srcset-style attribute whose URL uses a
// disallowed scheme, mirroring upstream's per-candidate srcset filtering. The
// scheme allowlist is keyed on the attribute name, as it is upstream.
func filterSrcset(value, attrName string, p *policy) string {
	var kept []string
	for _, cand := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(cand)
		if trimmed == "" {
			continue
		}
		url := trimmed
		if sp := strings.IndexAny(trimmed, " \t\n\f\r"); sp >= 0 {
			url = trimmed[:sp]
		}
		if naughtyHref(attrName, url, p) {
			continue
		}
		kept = append(kept, trimmed)
	}
	return strings.Join(kept, ", ")
}

// naughtyHref reports whether the URL value must be rejected for tag. It
// mirrors the upstream naughtyHref check: characters U+0000..U+0020 and HTML
// comments are stripped, a leading "scheme:" is extracted case-insensitively,
// and the scheme is tested against the tag-specific or global allowlist. URLs
// with no scheme are accepted, except protocol-relative ones when
// AllowProtocolRelative is false.
func naughtyHref(tag, value string, p *policy) bool {
	v := stripControl(value)
	v = stripComments(v)

	scheme, ok := leadingScheme(v)
	if !ok {
		// Protocol-relative URL: two or more '/' or '\' characters.
		if len(v) >= 2 && (v[0] == '/' || v[0] == '\\') && (v[1] == '/' || v[1] == '\\') {
			return !p.protoRel
		}
		return false
	}
	if list, has := p.opts.AllowedSchemesByTag[tag]; has {
		for _, s := range list {
			if s == "*" || strings.ToLower(s) == scheme {
				return false
			}
		}
		return true
	}
	if p.anyScheme {
		return false
	}
	return !p.schemes[scheme]
}

// stripControl removes every character in the range U+0000..U+0020, which
// browsers ignore in URL attributes and attackers use to break up "javascript:".
func stripControl(s string) string {
	return strings.Map(func(r rune) rune {
		if r <= 0x20 {
			return -1
		}
		return r
	}, s)
}

// stripComments removes "<!-- ... -->" runs, which browsers also ignore inside
// URL attributes. An unterminated "<!--" is left in place, as upstream leaves
// it: truncating there would hide a scheme that follows.
func stripComments(s string) string {
	for {
		start := strings.Index(s, "<!--")
		if start < 0 {
			return s
		}
		end := strings.Index(s[start+4:], "-->")
		if end < 0 {
			return s
		}
		s = s[:start] + s[start+4+end+3:]
	}
}

// leadingScheme returns the lower-cased scheme of s when it begins with
// "scheme:" using the /^([a-zA-Z][a-zA-Z0-9.\-+]*):/ grammar upstream uses.
func leadingScheme(s string) (string, bool) {
	if s == "" || !isLetter(s[0]) {
		return "", false
	}
	i := 1
	for i < len(s) {
		c := s[i]
		if isLetter(c) || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			i++
			continue
		}
		break
	}
	if i < len(s) && s[i] == ':' {
		return strings.ToLower(s[:i]), true
	}
	return "", false
}

// tokenize scans s into a flat list of text, start-tag, and end-tag tokens.
// Comments and declarations are dropped. The content of a raw-text element is
// captured on its start token instead of being scanned for markup.
func tokenize(s string, p *policy) []token {
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
		if strings.HasPrefix(s[i:], "<![CDATA[") {
			// Outside XML a CDATA section is not markup: an HTML parser reads it
			// as a bogus comment running to "]]>", so the whole section — script
			// tags and all — is dropped rather than scanned.
			end := strings.Index(s[i+9:], "]]>")
			if end == -1 {
				i = len(s)
			} else {
				i = i + 9 + end + 3
			}
			continue
		}
		if i+1 < len(s) && (s[i+1] == '!' || s[i+1] == '?') {
			// A declaration such as <!DOCTYPE ...>, or a processing
			// instruction such as <?php ... ?>, which HTML treats as a bogus
			// comment. Both run to the next '>' and are dropped.
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
				// An end tag left unterminated at the end of the input still
				// closes its element, as an HTML parser flushes it; emitting the
				// fragment as text instead would show markup to the reader.
				if name := parseTagName(s[i+2:]); name != "" {
					tokens = append(tokens, token{kind: tokenEnd, name: name})
				}
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
				i = next
				// Capture the content of a raw-text element verbatim. A solidus
				// in the start tag does not make such an element empty: an HTML
				// tokenizer ignores it and still switches to the raw-text state.
				if rawTextElements[tok.name] || escapableRawTextElements[tok.name] {
					tok.rawText = true
					tok.selfClosing = false
					tok.rawDecode = escapableRawTextElements[tok.name]
					rest := s[i:]
					closeIdx := indexCloseTag(rest, tok.name)
					if closeIdx == -1 {
						tok.rawContent = rest
						i = len(s)
					} else {
						gt := strings.IndexByte(rest[closeIdx:], '>')
						if gt == -1 {
							tok.rawContent = rest
							i = len(s)
						} else {
							tok.rawContent = rest[:closeIdx]
							i += closeIdx + gt + 1
						}
					}
				}
				tokens = append(tokens, tok)
				continue
			}
			// A start tag that never closes runs to the end of the input. An
			// HTML parser discards such a tag along with everything it has
			// swallowed, and so does this tokenizer: emitting the fragment as
			// text would surface attribute values — including event-handler
			// source — as visible content.
			break
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
