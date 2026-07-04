package striptags

import "testing"

func TestStripAllTags(t *testing.T) {
	got := StripTags("<a href='#'>Hello</a> <b>world</b>")
	if got != "Hello world" {
		t.Fatalf("StripTags = %q; want %q", got, "Hello world")
	}
}

func TestStripKeepsPlainText(t *testing.T) {
	if got := StripTags("no tags here"); got != "no tags here" {
		t.Fatalf("StripTags = %q", got)
	}
}

func TestAllowedTags(t *testing.T) {
	got := StripTags("<p>Hello <b>world</b></p>", "b")
	want := "Hello <b>world</b>"
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}

func TestAllowedTagsWithBrackets(t *testing.T) {
	got := StripTags("<p>Hi <i>there</i></p>", "<i>")
	want := "Hi <i>there</i>"
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}

func TestAllowedPreservesAttributes(t *testing.T) {
	got := StripTags(`<a href="http://x.com">link</a>`, "a")
	want := `<a href="http://x.com">link</a>`
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}

func TestSelfClosing(t *testing.T) {
	got := StripTags("line1<br/>line2")
	if got != "line1line2" {
		t.Fatalf("StripTags = %q", got)
	}
	got = StripTags("line1<br/>line2", "br")
	if got != "line1<br/>line2" {
		t.Fatalf("StripTags allowed br = %q", got)
	}
}

func TestStripComments(t *testing.T) {
	got := StripTags("before<!-- a comment -->after")
	if got != "beforeafter" {
		t.Fatalf("StripTags = %q", got)
	}
}

func TestCommentWithTagsInside(t *testing.T) {
	got := StripTags("x<!-- <b>hidden</b> -->y")
	if got != "xy" {
		t.Fatalf("StripTags = %q", got)
	}
}

func TestUnclosedTagDropped(t *testing.T) {
	got := StripTags("keep this <unclosed")
	if got != "keep this " {
		t.Fatalf("StripTags = %q", got)
	}
}

func TestStrayLessThan(t *testing.T) {
	got := StripTags("a < b and c > d")
	// "< b and c >" looks like a tag and is stripped; "a " and " d" remain.
	if got != "a  d" {
		t.Fatalf("StripTags = %q; want %q", got, "a  d")
	}
}

func TestNestedDisallowedInsideAllowed(t *testing.T) {
	got := StripTags("<div><p>Text <span>here</span></p></div>", "p")
	want := "<p>Text here</p>"
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}

func TestClosingTagAllowed(t *testing.T) {
	got := StripTags("<b>bold</b><i>italic</i>", "b")
	want := "<b>bold</b>italic"
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}

func TestCaseInsensitiveAllowed(t *testing.T) {
	got := StripTags("<B>bold</B>", "b")
	want := "<B>bold</B>"
	if got != want {
		t.Fatalf("StripTags = %q; want %q", got, want)
	}
}
