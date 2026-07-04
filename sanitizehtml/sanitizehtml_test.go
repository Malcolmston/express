package sanitizehtml

import "testing"

func TestStripDisallowedKeepText(t *testing.T) {
	opts := Options{AllowedTags: []string{"b"}}
	got := Sanitize("<div><b>bold</b> and <i>italic</i></div>", opts)
	want := "<b>bold</b> and italic"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestRemoveDisallowedAttributes(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"a"},
		AllowedAttributes: map[string][]string{"a": {"href"}},
	}
	got := Sanitize(`<a href="http://x.com" onclick="evil()">link</a>`, opts)
	want := `<a href="http://x.com">link</a>`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestScriptContentRemoved(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize(`<p>hi</p><script>alert('x < y')</script>`, opts)
	want := "<p>hi</p>"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestStyleContentRemoved(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize(`<p>hi</p><style>.a{color:red}</style>`, opts)
	want := "<p>hi</p>"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestWildcardAllowsAllTags(t *testing.T) {
	opts := Options{AllowedTags: []string{"*"}}
	got := Sanitize("<custom>x</custom>", opts)
	want := "<custom>x</custom>"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestWildcardAttributes(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"a", "p"},
		AllowedAttributes: map[string][]string{"*": {"class"}},
	}
	got := Sanitize(`<p class="c" id="i">t</p>`, opts)
	want := `<p class="c">t</p>`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestDefaultOptionsFormatting(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize(`<p>A <strong>b</strong> <em>c</em> <a href="/x">link</a></p>`, opts)
	want := `<p>A <strong>b</strong> <em>c</em> <a href="/x">link</a></p>`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestDefaultOptionsDropsUnknownAttr(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize(`<a href="/x" onmouseover="bad">y</a>`, opts)
	want := `<a href="/x">y</a>`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestCommentsRemoved(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize("a<!-- comment -->b", opts)
	if got != "ab" {
		t.Fatalf("Sanitize = %q", got)
	}
}

func TestDoctypeRemoved(t *testing.T) {
	opts := Options{AllowedTags: []string{"p"}}
	got := Sanitize("<!DOCTYPE html><p>x</p>", opts)
	if got != "<p>x</p>" {
		t.Fatalf("Sanitize = %q", got)
	}
}

func TestSelfClosingAllowed(t *testing.T) {
	opts := Options{AllowedTags: []string{"br", "img"}, AllowedAttributes: map[string][]string{"img": {"src"}}}
	got := Sanitize(`line<br/><img src="a.png" bad="x"/>`, opts)
	want := `line<br /><img src="a.png" />`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestLiteralLessThanKept(t *testing.T) {
	opts := Options{AllowedTags: []string{"b"}}
	got := Sanitize("2 < 3 is true", opts)
	if got != "2 < 3 is true" {
		t.Fatalf("Sanitize = %q", got)
	}
}

func TestAttributeValueReEscaped(t *testing.T) {
	opts := Options{
		AllowedTags:       []string{"a"},
		AllowedAttributes: map[string][]string{"a": {"title"}},
	}
	got := Sanitize(`<a title="Tom &amp; Jerry">x</a>`, opts)
	want := `<a title="Tom &amp; Jerry">x</a>`
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestNestedDisallowedTags(t *testing.T) {
	opts := Options{AllowedTags: []string{"p"}}
	got := Sanitize("<div><section><p>keep</p></section></div>", opts)
	want := "<p>keep</p>"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestUnterminatedTagAsText(t *testing.T) {
	opts := Options{AllowedTags: []string{"b"}}
	got := Sanitize("keep <b>x</b> and <notclosed", opts)
	want := "keep <b>x</b> and <notclosed"
	if got != want {
		t.Fatalf("Sanitize = %q; want %q", got, want)
	}
}

func TestScriptTagWithoutClosing(t *testing.T) {
	opts := DefaultOptions()
	got := Sanitize("<p>ok</p><script>never ends", opts)
	if got != "<p>ok</p>" {
		t.Fatalf("Sanitize = %q", got)
	}
}
