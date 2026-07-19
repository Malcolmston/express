package striptags

import "testing"

// Upstream parity tests for the npm library "ericnorris/striptags".
//
// Input -> expected-output vectors are taken verbatim from the original
// library's own test suite at version 3.2.0 (the classic hand-rolled
// state-machine implementation this Go port mirrors):
//
//	https://raw.githubusercontent.com/ericnorris/striptags/v3.2.0/test/striptags-test.js
//	https://raw.githubusercontent.com/ericnorris/striptags/v3.2.0/src/striptags.js
//
// The upstream signature is striptags(html, allowable_tags, tag_replacement).
// This port exposes StripTags(html, allowed ...string); tag_replacement is not
// supported (see notes), so the one tag_replacement vector is intentionally not
// encoded here. The module-definition, streaming-mode, and type-confusion tests
// are runtime-specific to Node/JS and do not apply to the Go API.

func TestParityNoOptionalParams(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// "should not strip invalid tags": '<' followed by a space is literal.
		{"invalid-tags", "lorem ipsum < a> < div>", "lorem ipsum < a> < div>"},
		// "should remove simple HTML tags"
		{"simple-tags", `<a href="">lorem <strong>ipsum</strong></a>`, "lorem ipsum"},
		// "should remove comments"
		{"comments", "<!-- lorem -- ipsum -- --> dolor sit amet", " dolor sit amet"},
		// "should strip tags within comments"
		{"tags-in-comments", "<!-- <strong>lorem ipsum</strong> --> dolor sit", " dolor sit"},
		// "should not fail with nested quotes"
		{"nested-quotes", `<article attr="foo 'bar'">lorem</article> ipsum`, "lorem ipsum"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := StripTags(c.in); got != c.want {
				t.Fatalf("StripTags(%q) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParityAllowedTags(t *testing.T) {
	// "should parse a string": allowed tag given as "<strong>".
	if got := StripTags("<strong>lorem ipsum</strong>", "<strong>"); got != "<strong>lorem ipsum</strong>" {
		t.Fatalf("string allowed = %q", got)
	}
	// "should take an array": allowed [strong, em] -> variadic.
	if got := StripTags("<strong>lorem <em>ipsum</em></strong>", "strong", "em"); got != "<strong>lorem <em>ipsum</em></strong>" {
		t.Fatalf("array allowed = %q", got)
	}
}

func TestParityAllowableTagsParameter(t *testing.T) {
	// "should leave attributes when allowing HTML"
	if got := StripTags(`<a href="https://example.com">lorem ipsum</a>`, "<a>"); got != `<a href="https://example.com">lorem ipsum</a>` {
		t.Fatalf("attributes = %q", got)
	}
	// "should strip extra < within tags"
	if got := StripTags("<div<>>lorem ipsum</div>", "<div>"); got != "<div>lorem ipsum</div>" {
		t.Fatalf("extra-lt = %q", got)
	}
	// "should strip <> within quotes"
	if got := StripTags(`<a href="<script>">lorem ipsum</a>`, "<a>"); got != `<a href="script">lorem ipsum</a>` {
		t.Fatalf("lt-in-quotes = %q", got)
	}
}
