package sanitizehtml

import (
	"strings"
	"testing"
)

// TestNaughtyHrefRejected covers the URL scheme allowlist. Before this check
// existed, every one of these inputs survived DefaultOptions() untouched, which
// is a stored-XSS hole for any application rendering the result. The evasion
// forms are the ones listed in the OWASP XSS Filter Evasion Cheat Sheet and
// handled by upstream sanitize-html's naughtyHref().
func TestNaughtyHrefRejected(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", `<a href="javascript:alert(1)">x</a>`, `<a>x</a>`},
		{"mixed-case", `<a href="JaVaScRiPt:alert(1)">x</a>`, `<a>x</a>`},
		{"entity-encoded", `<a href="java&#115;cript:alert(1)">x</a>`, `<a>x</a>`},
		{"leading-space", `<a href="  javascript:alert(1)">x</a>`, `<a>x</a>`},
		{"embedded-tab", "<a href=\"jav\tascript:alert(1)\">x</a>", `<a>x</a>`},
		{"embedded-newline", "<a href=\"jav\nascript:alert(1)\">x</a>", `<a>x</a>`},
		{"embedded-nul", "<a href=\"jav\x00ascript:alert(1)\">x</a>", `<a>x</a>`},
		{"embedded-comment", `<a href="jav<!-- foo -->ascript:alert(1)">x</a>`, `<a>x</a>`},
		{"data-url", `<a href="data:text/html;base64,PHNjcmlwdD4=">x</a>`, `<a>x</a>`},
		{"vbscript", `<a href="vbscript:msgbox(1)">x</a>`, `<a>x</a>`},

		// Legitimate URLs must still survive.
		{"http", `<a href="http://example.com/">x</a>`, `<a href="http://example.com/">x</a>`},
		{"https", `<a href="https://example.com/">x</a>`, `<a href="https://example.com/">x</a>`},
		{"mailto", `<a href="mailto:a@b.example">x</a>`, `<a href="mailto:a@b.example">x</a>`},
		{"tel", `<a href="tel:+15550100">x</a>`, `<a href="tel:+15550100">x</a>`},
		{"relative", `<a href="hello.html">x</a>`, `<a href="hello.html">x</a>`},
		{"relative-with-colon-later", `<a href="awesome.html#this:stuff">x</a>`, `<a href="awesome.html#this:stuff">x</a>`},
		{"fragment", `<a href="#top">x</a>`, `<a href="#top">x</a>`},
		{"protocol-relative", `<a href="//cdn.example/x">x</a>`, `<a href="//cdn.example/x">x</a>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in, DefaultOptions()); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestImageURLFiltering covers src/srcset filtering on <img>, which is not in
// the upstream default allowedTags but does have default allowedAttributes.
// srcset is filtered candidate-by-candidate, matching upstream.
func TestImageURLFiltering(t *testing.T) {
	opts := DefaultOptions()
	opts.AllowedTags = append(opts.AllowedTags, "img")
	cases := []struct{ in, want string }{
		{`<img src="javascript:alert(1)">`, `<img />`},
		{`<img src="/ok.png">`, `<img src="/ok.png" />`},
		{`<img srcset="javascript:alert(1) 1x, /ok.png 2x">`, `<img srcset="/ok.png 2x" />`},
		{`<img srcset="/a.png 1x, /b.png 2x">`, `<img srcset="/a.png 1x, /b.png 2x" />`},
	}
	for _, c := range cases {
		if got := Sanitize(c.in, opts); got != c.want {
			t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// TestZeroOptionsStillFiltersSchemes documents that a zero Options value (nil
// AllowedSchemes) uses DefaultSchemes rather than "allow everything", so a
// hand-built policy cannot accidentally opt out of URL filtering.
func TestZeroOptionsStillFiltersSchemes(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"a"},
		AllowedAttributes: map[string][]string{"a": {"href"}},
	}
	if got := Sanitize(`<a href="javascript:alert(1)">x</a>`, opts); got != `<a>x</a>` {
		t.Fatalf("Sanitize = %q; want %q", got, `<a>x</a>`)
	}
}

// TestAllowedSchemesOverrides covers the explicit forms of AllowedSchemes:
// "*" allows everything, an empty non-nil slice allows nothing, and
// AllowedSchemesByTag overrides the global list per tag.
func TestAllowedSchemesOverrides(t *testing.T) {
	base := func() Options {
		o := DefaultOptions()
		return o
	}

	all := base()
	all.AllowedSchemes = []string{"*"}
	if got := Sanitize(`<a href="javascript:x">y</a>`, all); got != `<a href="javascript:x">y</a>` {
		t.Errorf(`AllowedSchemes ["*"] = %q`, got)
	}

	none := base()
	none.AllowedSchemes = []string{}
	if got := Sanitize(`<a href="http://x/">y</a>`, none); got != `<a>y</a>` {
		t.Errorf("AllowedSchemes [] = %q", got)
	}
	// A relative URL has no scheme and is always fine.
	if got := Sanitize(`<a href="/y">y</a>`, none); got != `<a href="/y">y</a>` {
		t.Errorf("AllowedSchemes [] relative = %q", got)
	}

	byTag := base()
	byTag.AllowedSchemesByTag = map[string][]string{"a": {"ftp"}}
	if got := Sanitize(`<a href="http://x/">y</a>`, byTag); got != `<a>y</a>` {
		t.Errorf("AllowedSchemesByTag http = %q", got)
	}
	if got := Sanitize(`<a href="ftp://x/">y</a>`, byTag); got != `<a href="ftp://x/">y</a>` {
		t.Errorf("AllowedSchemesByTag ftp = %q", got)
	}
}

// TestAllowProtocolRelative covers the AllowProtocolRelative option, which
// defaults to true upstream and is expressed here as a *bool so that the zero
// Options value keeps the upstream default.
func TestAllowProtocolRelative(t *testing.T) {
	no := false
	opts := DefaultOptions()
	opts.AllowProtocolRelative = &no
	for _, in := range []string{`<a href="//evil.example/x">y</a>`, `<a href="\\evil.example\x">y</a>`} {
		if got := Sanitize(in, opts); got != `<a>y</a>` {
			t.Errorf("Sanitize(%q) = %q; want %q", in, got, `<a>y</a>`)
		}
	}
}

// TestAllowedClasses covers class filtering, including the "*" tag key, prefix
// globs, and the implicit permission of the class attribute that upstream
// grants when allowedClasses names a tag.
func TestAllowedClasses(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		classes map[string][]string
	}{
		{"keep-listed", `<p class="a b c">x</p>`, `<p class="a c">x</p>`, map[string][]string{"p": {"a", "c"}}},
		{"drop-all", `<p class="q">x</p>`, `<p>x</p>`, map[string][]string{"p": {"a"}}},
		{"glob", `<p class="col-1 row-2">x</p>`, `<p class="col-1">x</p>`, map[string][]string{"p": {"col-*"}}},
		{"wildcard-tag", `<p class="a">x</p><span class="a">y</span>`, `<p class="a">x</p><span class="a">y</span>`, map[string][]string{"*": {"a"}}},
		{"other-tag-unlisted", `<span class="a">y</span>`, `<span>y</span>`, map[string][]string{"p": {"a"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := DefaultOptions()
			opts.AllowedClasses = c.classes
			if got := Sanitize(c.in, opts); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestDisallowedTagsMode covers the four upstream disallowedTagsMode values.
func TestDisallowedTagsMode(t *testing.T) {
	const in = `<p>keep <bad attr="v">inner <b>bold</b></bad> tail</p>`
	cases := []struct {
		mode string
		want string
	}{
		{ModeDiscard, `<p>keep inner <b>bold</b> tail</p>`},
		{ModeCompletelyDiscard, `<p>keep  tail</p>`},
		// The escaped rendering is the NORMALIZED tag, so attr — which the
		// policy does not permit on <bad> — is gone from it too, matching
		// sanitize-html 2.17.6.
		{ModeEscape, `<p>keep &lt;bad&gt;inner <b>bold</b>&lt;/bad&gt; tail</p>`},
		{ModeRecursiveEscape, `<p>keep &lt;bad&gt;inner &lt;b&gt;bold&lt;/b&gt;&lt;/bad&gt; tail</p>`},
	}
	for _, c := range cases {
		t.Run(c.mode, func(t *testing.T) {
			opts := DefaultOptions()
			opts.DisallowedTagsMode = c.mode
			if got := Sanitize(in, opts); got != c.want {
				t.Errorf("mode %s: Sanitize = %q; want %q", c.mode, got, c.want)
			}
		})
	}
}

// TestNonTextTagsOption covers overriding the nonTextTags list: a tag named
// there has its raw content consumed with the tag rather than surfaced as text.
func TestNonTextTagsOption(t *testing.T) {
	opts := DefaultOptions()
	opts.NonTextTags = []string{"secret"}
	if got := Sanitize(`<p>a</p><secret>hidden</secret><script>shown()</script>`, opts); got != `<p>a</p>shown()` {
		t.Fatalf("Sanitize = %q", got)
	}
}

// TestSelfClosingOption covers serialization of void elements: a tag named in
// SelfClosing is written as "<br />" even when the source wrote "<br>". A stray
// "</br>" is the one end tag HTML turns into an element, so it becomes another
// <br />, matching sanitize-html 2.17.6.
func TestSelfClosingOption(t *testing.T) {
	opts := DefaultOptions()
	if got := Sanitize("a<br>b</br>c", opts); got != "a<br />b<br />c" {
		t.Fatalf("Sanitize = %q", got)
	}
	opts.SelfClosing = []string{}
	// <br> is still a void element; with nothing in SelfClosing it is closed by
	// an explicit end tag instead of the "<br />" form, as upstream does.
	if got := Sanitize("a<br>b", opts); got != "a<br></br>b" {
		t.Fatalf("Sanitize with empty SelfClosing = %q", got)
	}
}

// TestTextFilter covers the textFilter hook, which receives already-escaped
// text plus the name of the enclosing tag.
func TestTextFilter(t *testing.T) {
	opts := DefaultOptions()
	var seen []string
	opts.TextFilter = func(text, tag string) string {
		seen = append(seen, tag)
		return "[" + text + "]"
	}
	if got := Sanitize("top<p>inner</p>", opts); got != "[top]<p>[inner]</p>" {
		t.Fatalf("Sanitize = %q", got)
	}
	if len(seen) != 2 || seen[0] != "" || seen[1] != "p" {
		t.Fatalf("TextFilter tags = %#v", seen)
	}
}

// TestAttributePrefixGlob covers the "data-*" prefix form of an allowed
// attribute name, which upstream supports via allowedAttributes globs.
func TestAttributePrefixGlob(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"p"},
		AllowedAttributes: map[string][]string{"p": {"data-*"}},
	}
	got := Sanitize(`<p data-x="1" data-y="2" other="3">t</p>`, opts)
	if got != `<p data-x="1" data-y="2">t</p>` {
		t.Fatalf("Sanitize = %q", got)
	}
}

// TestSanitizeDoesNotPanic feeds deliberately malformed markup through the
// sanitizer to confirm it terminates without panicking.
func TestSanitizeDoesNotPanic(t *testing.T) {
	inputs := []string{
		"", "<", "<<<<", "</", "</>", "<!", "<!--", "<!-- unterminated",
		"<a", "<a href", "<a href=", `<a href="`, "<a href='x", "<a/",
		"<script", "<script>", "<script>x", "<style>a{b:c}", "<textarea",
		"<b</p>", "<p><p><p><p>", "</p></p>", "<p " + strings.Repeat("a", 1000),
		"<img src=x onerror=alert(1)>", "<svg/onload=alert(1)>",
		"<a href=javascript:alert(1)>x</a>",
	}
	for _, in := range inputs {
		_ = Sanitize(in, DefaultOptions())
		_ = Sanitize(in, Options{})
		for _, mode := range []string{ModeEscape, ModeRecursiveEscape, ModeCompletelyDiscard} {
			o := DefaultOptions()
			o.DisallowedTagsMode = mode
			_ = Sanitize(in, o)
		}
	}
}

// TestUnquotedJavascriptHref covers the unquoted attribute form, which the
// tokenizer parses separately from the quoted one.
func TestUnquotedJavascriptHref(t *testing.T) {
	if got := Sanitize(`<a href=javascript:alert(1)>x</a>`, DefaultOptions()); got != `<a>x</a>` {
		t.Fatalf("Sanitize = %q", got)
	}
	if got := Sanitize(`<a href='javascript:alert(1)'>x</a>`, DefaultOptions()); got != `<a>x</a>` {
		t.Fatalf("Sanitize = %q", got)
	}
}
