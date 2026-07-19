package escapehtml

import "testing"

// Upstream parity vectors extracted verbatim from the original npm library
// "escape-html" (component/escape-html) test suite:
//
//	https://raw.githubusercontent.com/component/escape-html/master/test/index.js
//	https://raw.githubusercontent.com/component/escape-html/master/index.js
//
// The upstream tests also cover non-string coercion (undefined -> "undefined",
// null -> "null", 42 -> "42", {} -> "[object Object]"). Those exercise
// JavaScript's `'' + string` value coercion and have no analogue for a Go
// func(string) string signature, so they are intentionally omitted here; all
// string-typed vectors below are reproduced exactly.

func TestParityUpstream(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		// contains '"'
		{`"`, "&quot;"},
		{`"bar`, "&quot;bar"},
		{`foo"`, "foo&quot;"},
		{`foo"bar`, "foo&quot;bar"},
		{`foo""bar`, "foo&quot;&quot;bar"},
		// contains '&'
		{"&", "&amp;"},
		{"&bar", "&amp;bar"},
		{"foo&", "foo&amp;"},
		{"foo&bar", "foo&amp;bar"},
		{"foo&&bar", "foo&amp;&amp;bar"},
		// contains '\''
		{"'", "&#39;"},
		{"'bar", "&#39;bar"},
		{"foo'", "foo&#39;"},
		{"foo'bar", "foo&#39;bar"},
		{"foo''bar", "foo&#39;&#39;bar"},
		// contains '<'
		{"<", "&lt;"},
		{"<bar", "&lt;bar"},
		{"foo<", "foo&lt;"},
		{"foo<bar", "foo&lt;bar"},
		{"foo<<bar", "foo&lt;&lt;bar"},
		// contains '>'
		{">", "&gt;"},
		{">bar", "&gt;bar"},
		{"foo>", "foo&gt;"},
		{"foo>bar", "foo&gt;bar"},
		{"foo>>bar", "foo&gt;&gt;bar"},
		// mixed
		{`&foo <> bar "fizz" l'a`, "&amp;foo &lt;&gt; bar &quot;fizz&quot; l&#39;a"},
	}
	for _, tt := range tests {
		if got := Escape(tt.in); got != tt.want {
			t.Errorf("Escape(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
