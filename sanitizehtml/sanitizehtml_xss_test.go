package sanitizehtml

import "testing"

// TestAttributeNameInjection covers attribute names that carry characters the
// permissive tokenizer accepts but that must never be echoed back into a start
// tag. The tokenizer ends an attribute name at whitespace, '=', '>' and '/',
// so a name can still contain a '<', a double quote, a single quote or a
// backtick. Before attrNameSafe
// existed, a policy with a "data-*" glob serialized those names verbatim, so
// `<p data-x<img src=x>` produced `<p data-x<img="">` -- raw markup inside a
// tag. Every case below must lose the offending attribute entirely.
func TestAttributeNameInjection(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"p"},
		AllowedAttributes: map[string][]string{"p": {"data-*"}},
	}
	cases := []struct{ name, in, want string }{
		{"angle-bracket", `<p data-x<img src=x>t</p>`, `<p>t</p>`},
		{"double-quote", `<p data-x"onmouseover=alert(1)>t</p>`, `<p>t</p>`},
		{"backtick", "<p data-x`onmouseover=alert(1)`>t</p>", `<p>t</p>`},
		{"single-quote", `<p data-x'onmouseover=alert(1)>t</p>`, `<p>t</p>`},
		// A well-formed data-* attribute is unaffected.
		{"clean", `<p data-x="1">t</p>`, `<p data-x="1">t</p>`},
		{"clean-dotted", `<p data-a.b="1">t</p>`, `<p data-a.b="1">t</p>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in, opts); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSrcsetFilteredOnEveryTag covers srcset candidate filtering on a tag other
// than <img>. Upstream sanitize-html keys the srcset check on the attribute
// name alone (`if (a === 'srcset')`), not on the element, so a <source
// srcset="javascript:..."> is filtered exactly like an <img srcset>. This port
// previously gated the check on tok.name == "img" and let the <source> form
// through unchanged.
func TestSrcsetFilteredOnEveryTag(t *testing.T) {
	opts := DefaultOptions()
	opts.AllowedTags = append(opts.AllowedTags, "source", "img")
	opts.AllowedAttributes["source"] = []string{"srcset", "type"}

	cases := []struct{ in, want string }{
		// <source> is a void element that the default SelfClosing list does not
		// name, so it is closed by an explicit end tag, as upstream does.
		{`<source srcset="javascript:alert(1) 1x">`, `<source></source>`},
		{`<source srcset="javascript:alert(1) 1x, /ok.png 2x">`, `<source srcset="/ok.png 2x"></source>`},
		{`<source srcset="/a.png 1x, /b.png 2x">`, `<source srcset="/a.png 1x, /b.png 2x"></source>`},
		{`<img srcset="vbscript:x 1x">`, `<img />`},
	}
	for _, c := range cases {
		if got := Sanitize(c.in, opts); got != c.want {
			t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// TestNestedTagObfuscation covers the classic "split tag" evasions from the
// OWASP XSS Filter Evasion Cheat Sheet. None of them may yield a live <script>
// element; whatever survives must be inert text.
func TestNestedTagObfuscation(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"doubled-open", `<<script>script>alert(1);//<</script>script>`, `&lt;script&gt;`},
		{"split-name", `<scr<script>ipt>alert(1)</script>`, `ipt&gt;alert(1)`},
		{"comment-wrapped", `<!--><script>alert(1)</script-->`, ``},
		{"svg-script", `<svg><script>alert(1)</script></svg>`, ``},
		{"uppercase-textarea", `<TEXTAREA><script>alert(1)</script></TEXTAREA>`, ``},
		// Outside XML a CDATA section is a bogus comment running to "]]>", so the
		// whole section is dropped rather than scanned for markup.
		{"cdata", `<![CDATA[<script>alert(1)</script>]]>`, ``},
		{"xlink-href", `<math><mi xlink:href="javascript:alert(1)">x</mi></math>`, `x`},
		// An unterminated END tag still closes its element rather than being
		// shown as text.
		{"unterminated-end-tag", `<p>a</p `, `<p>a</p>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Sanitize(c.in, DefaultOptions()); got != c.want {
				t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestEventHandlersAlwaysDropped confirms that inline event handlers never
// survive the default policy, on both allowed and disallowed elements.
func TestEventHandlersAlwaysDropped(t *testing.T) {
	opts := DefaultOptions()
	opts.AllowedTags = append(opts.AllowedTags, "img")
	cases := []struct{ in, want string }{
		{`<img src="/x.png" onerror="alert(1)">`, `<img src="/x.png" />`},
		{`<a href="http://x/" onclick="alert(1)">y</a>`, `<a href="http://x/">y</a>`},
		{`<a href="http://x/" ONMOUSEOVER="alert(1)">y</a>`, `<a href="http://x/">y</a>`},
		{`<p onload=alert(1)>y</p>`, `<p>y</p>`},
		{`<svg/onload=alert(1)>`, ``},
	}
	for _, c := range cases {
		if got := Sanitize(c.in, opts); got != c.want {
			t.Errorf("Sanitize(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}
