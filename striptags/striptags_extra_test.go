package striptags

import (
	"fmt"
	"strings"
	"testing"
)

// TestStripTagsWithReplacement covers the third argument of the Node function,
// striptags(html, allowedTags, tagReplacement). Upstream test/index.js:
//
//	assert.equal(striptags('<b>bold</b>', [], '**'), '**bold**')
func TestStripTagsWithReplacement(t *testing.T) {
	cases := []struct {
		name        string
		in          string
		allowed     []string
		replacement string
		want        string
	}{
		{"bold-stars", "<b>bold</b>", nil, "**", "**bold**"},
		{"empty-replacement", "<b>bold</b>", nil, "", "bold"},
		{"space-replacement", "<p>a</p><p>b</p>", nil, " ", " a  b "},
		{"allowed-not-replaced", "<b>x</b><i>y</i>", []string{"b"}, "|", "<b>x</b>|y|"},
		{"comment-not-replaced", "a<!-- c -->b", nil, "|", "ab"},
		{"self-closing", "a<br/>b", nil, "\n", "a\nb"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := StripTagsWith(c.in, c.allowed, c.replacement); got != c.want {
				t.Errorf("StripTagsWith(%q, %v, %q) = %q; want %q", c.in, c.allowed, c.replacement, got, c.want)
			}
		})
	}
}

// TestAllowedTagSpecifierForms covers the allowed-tag specifier forms upstream
// accepts: a bracketed list in one string ("<a><p>"), individual bracketed
// names, closing-tag forms, and the bare-name convenience this port allows.
func TestAllowedTagSpecifierForms(t *testing.T) {
	const in = `<a href="/x">link</a> <p>para</p> <b>bold</b>`
	const want = `<a href="/x">link</a> <p>para</p> bold`
	forms := [][]string{
		{"<a><p>"},
		{"<a>", "<p>"},
		{"a", "p"},
		{"</a>", "</p>"},
		{"<A><P>"},
	}
	for _, allowed := range forms {
		t.Run(strings.Join(allowed, ","), func(t *testing.T) {
			if got := StripTags(in, allowed...); got != want {
				t.Errorf("StripTags(in, %v) = %q; want %q", allowed, got, want)
			}
		})
	}
}

// TestStripTagsMalformedInput feeds adversarial and malformed markup through
// the state machine to confirm it terminates and never panics, and pins the
// documented edge-case behaviors (nested angle brackets, quoted delimiters,
// unterminated tags, and "< " returning to plaintext).
func TestStripTagsMalformedInput(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"<", ""},
		{"<<<<", ""},
		{">", ">"},
		{"a < b", "a < b"},
		{"a <\nb", "a < b"},
		{"unterminated <b", "unterminated "},
		{"<<script>script>alert(1)<</script>/script>", "alert(1)"},
		{`<a href="a>b">t</a>`, "t"},
		{`<a href='a>b'>t</a>`, "t"},
		{"<!-- <b>x</b> -->y", "y"},
		{"<!-- unterminated comment", ""},
		// "<!--->" closes: on reaching '>' the two preceding buffered chars are
		// "--", which is the reference implementation's comment-close test.
		{"<!--->x", "x"},
		{"a<!---->b", "ab"},
	}
	for _, c := range cases {
		if got := StripTags(c.in); got != c.want {
			t.Errorf("StripTags(%q) = %q; want %q", c.in, got, c.want)
		}
	}
	// Long pathological input must not blow up.
	_ = StripTags(strings.Repeat("<", 10000) + strings.Repeat(">", 10000))
	_ = StripTags(strings.Repeat(`<a href="`, 5000))
}

// ExampleStripTagsWith shows replacing every removed tag with a marker, the
// three-argument form of the Node original.
func ExampleStripTagsWith() {
	fmt.Println(StripTagsWith("<p>Hello</p><p>World</p>", nil, " "))
	// Output:  Hello  World
}

// ExampleStripTags_allowedList shows keeping a subset of tags using the
// bracketed one-string form of the allowlist.
func ExampleStripTags_allowedList() {
	fmt.Println(StripTags(`<div><a href="/x">click</a> <em>me</em></div>`, "<a>"))
	// Output: <a href="/x">click</a> me
}
