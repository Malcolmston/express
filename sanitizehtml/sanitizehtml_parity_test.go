package sanitizehtml

import "testing"

// Upstream-parity vectors for apostrophecms/sanitize-html.
//
// Every input -> expected pair below is taken verbatim from the upstream test
// suite and default configuration, not invented:
//
//	https://raw.githubusercontent.com/apostrophecms/sanitize-html/main/test/test.js
//	https://raw.githubusercontent.com/apostrophecms/sanitize-html/main/index.js
//
// The upstream sanitizeHtml(html, options) signature maps onto this port's
// Sanitize(html, Options). Vectors are limited to behaviour this port's Options
// can express: the allowedTags/allowedAttributes allowlist model, text-
// preserving removal of disallowed tags, and the default nonTextTags content
// stripping (script/style/textarea/option). Upstream features the port does not
// implement (URL scheme validation, implicit tag auto-closing, disallowedTagsMode,
// transformTags, the allowedTags:false "allow everything" sentinel, custom
// nonTextTags) are recorded as gaps and are intentionally not asserted here.

// defaultAttrs returns the upstream default allowedAttributes, used when a
// vector overrides only allowedTags.
func defaultAttrs() map[string][]string {
	return DefaultOptions().AllowedAttributes
}

func TestParityDefaultAllowlist(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		opts Options
	}{
		// test.js: 'should pass through simple, well-formed markup'
		{"well-formed", "<div><p>Hello <b>there</b></p></div>", "<div><p>Hello <b>there</b></p></div>", DefaultOptions()},
		// test.js: 'should respect text nodes at top level'
		{"top-level-text", "Blah blah blah<p>Whee!</p>", "Blah blah blah<p>Whee!</p>", DefaultOptions()},
		// test.js: 'should reject markup not whitelisted...' (default: wiggly stripped, text kept)
		{"strip-unknown-keep-text", "<div><wiggly>Hello</wiggly></div>", "<div>Hello</div>", DefaultOptions()},
		// test.js: undefined/null/0/'' allowedTags -> strip all markup, keep text
		{"empty-allowed-tags", "<div><wiggly worms=\"ewww\">hello</wiggly></div>", "hello", Options{}},
		// test.js: 'should drop the attributes not whitelisted'
		{"drop-unlisted-attr", `<a href="foo.html" whizbang="whangle">foo</a>`, `<a href="foo.html">foo</a>`, DefaultOptions()},
		// test.js: 'should not filter if whitelisted...' custom attr allowlist
		{"custom-attr-allowlist", `<a href="foo.html" whizbang="whangle">foo</a>`, `<a href="foo.html" whizbang="whangle">foo</a>`,
			Options{AllowedTags: DefaultOptions().AllowedTags, AllowedAttributes: map[string][]string{"a": {"href", "whizbang"}}}},
		// test.js: custom allowedTags list, nested disallowed tag stripped
		{"custom-tag-allowlist", "<blue><red><green>Cheese</green></red></blue>", "<blue><green>Cheese</green></blue>",
			Options{AllowedTags: []string{"blue", "green"}}},
		// test.js: 'should preserve entities as such'
		{"preserve-entities", `<a name="&lt;silly&gt;">&lt;Kapow!&gt;</a>`, `<a name="&lt;silly&gt;">&lt;Kapow!&gt;</a>`, DefaultOptions()},
		// test.js: 'should remove comments'
		{"remove-comment", "<p><!-- Blah blah -->Whee</p>", "<p>Whee</p>", DefaultOptions()},
		// test.js: relative href with a colon after the fragment is kept
		{"relative-href-colon", `<a href="awesome.html#this:stuff">Hi</a>`, `<a href="awesome.html#this:stuff">Hi</a>`, DefaultOptions()},
		// test.js: plain http href kept
		{"http-href", `<a href="http://google.com/">Hi</a>`, `<a href="http://google.com/">Hi</a>`, DefaultOptions()},
		// test.js: relative href kept
		{"relative-href", `<a href="hello.html">Hi</a>`, `<a href="hello.html">Hi</a>`, DefaultOptions()},
		// test.js: 'should retain the content of fibble elements by default'
		{"fibble-keep-content", "<fibble>Nifty</fibble><p>Paragraph</p>", "Nifty<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should retain allowed tags within a fibble element...'
		{"fibble-nested-allowed", "<fibble>Ni<em>f</em>ty</fibble><p>Paragraph</p>", "Ni<em>f</em>ty<p>Paragraph</p>", DefaultOptions()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in, c.opts); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParityNonTextTags(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		opts Options
	}{
		// test.js: 'should drop the content of script elements'
		{"drop-script", `<script>alert("ruhroh!");</script><p>Paragraph</p>`, "<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should drop the content of style elements'
		{"drop-style", "<style>.foo { color: blue; }</style><p>Paragraph</p>", "<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should drop the content of textarea elements'
		{"drop-textarea", "<textarea>Nifty</textarea><p>Paragraph</p>", "<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should drop the content of option elements'
		{"drop-option", "<select><option>one</option><option>two</option></select><p>Paragraph</p>", "<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should drop the content of textarea elements but keep the closing parent tag, when nested'
		{"drop-nested-textarea", "<p>Paragraph<textarea>Nifty</textarea></p>", "<p>Paragraph</p>", DefaultOptions()},
		// test.js: 'should preserve textarea content if textareas are allowed'
		{"keep-allowed-textarea", "<textarea>Nifty</textarea><p>Paragraph</p>", "<textarea>Nifty</textarea><p>Paragraph</p>",
			Options{AllowedTags: []string{"textarea", "p"}, AllowedAttributes: defaultAttrs()}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in, c.opts); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}
