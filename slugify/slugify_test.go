package slugify

import "testing"

func TestBasic(t *testing.T) {
	if got := Slugify("Hello World"); got != "Hello-World" {
		t.Errorf("got %q", got)
	}
}

func TestLowerAndAccents(t *testing.T) {
	if got := Slugify("Héllo World!", Options{Lower: true}); got != "hello-world" {
		t.Errorf("got %q", got)
	}
}

func TestTransliteration(t *testing.T) {
	if got := Slugify("Héllo"); got != "Hello" {
		t.Errorf("got %q", got)
	}
	if got := Slugify("Crème brûlée", Options{Lower: true}); got != "creme-brulee" {
		t.Errorf("got %q", got)
	}
}

func TestAmpersand(t *testing.T) {
	if got := Slugify("Salt & Pepper", Options{Lower: true}); got != "salt-and-pepper" {
		t.Errorf("got %q", got)
	}
}

func TestCustomSeparator(t *testing.T) {
	if got := Slugify("Hello World", Options{Separator: "_"}); got != "Hello_World" {
		t.Errorf("got %q", got)
	}
}

func TestStrict(t *testing.T) {
	// Underscore is a word char kept by default but removed under strict.
	if got := Slugify("foo_bar baz"); got != "foo_bar-baz" {
		t.Errorf("default got %q", got)
	}
	if got := Slugify("foo_bar baz", Options{Strict: true, Separator: "-"}); got != "foobar-baz" {
		t.Errorf("strict got %q", got)
	}
}

func TestCollapseSeparators(t *testing.T) {
	if got := Slugify("a   b"); got != "a-b" {
		t.Errorf("collapse whitespace got %q", got)
	}
	if got := Slugify("a---b"); got != "a-b" {
		t.Errorf("collapse literal separators got %q", got)
	}
}

func TestTrim(t *testing.T) {
	if got := Slugify("  padded  "); got != "padded" {
		t.Errorf("got %q", got)
	}
	if got := Slugify("  padded  ", Options{Trim: false, Separator: "-"}); got != "-padded-" {
		t.Errorf("no-trim got %q", got)
	}
}

func TestEmpty(t *testing.T) {
	if got := Slugify(""); got != "" {
		t.Errorf("got %q", got)
	}
}
