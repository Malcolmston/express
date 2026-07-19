package camelcase

import "testing"

func TestCamelCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo-bar_baz", "fooBarBaz"},
		{"foo bar", "fooBar"},
		{"Foo Bar", "fooBar"},
		{"fooBar", "fooBar"},
		{"FOOBar", "fooBar"},
		{"foo-bar", "fooBar"},
		{"foo_bar", "fooBar"},
		{"--foo.bar", "fooBar"},
		{"foo-bar-baz", "fooBarBaz"},
		{"foo123", "foo123"},
		{"foo-123", "foo123"},
		{"1-foo-bar", "1FooBar"},
		{"FOO-BAR", "fooBar"},
		{"foo", "foo"},
		{"Foo", "foo"},
		{"", ""},
		{"   ", ""},
		{"___", "___"}, // upstream v9 preserves leading underscores
		{"HTTPServer", "httpServer"},
		{"a", "a"},
		{"A", "a"},
		{"foo bar  baz", "fooBarBaz"},
		{"  foo  ", "foo"},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPascalCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo-bar_baz", "FooBarBaz"},
		{"foo bar", "FooBar"},
		{"fooBar", "FooBar"},
		{"foo", "Foo"},
		{"", ""},
		{"1-foo-bar", "1FooBar"},
	}
	for _, c := range cases {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCamelCaseWith(t *testing.T) {
	if got := CamelCaseWith("foo-bar", Options{Pascal: false}); got != "fooBar" {
		t.Errorf("got %q, want fooBar", got)
	}
	if got := CamelCaseWith("foo-bar", Options{Pascal: true}); got != "FooBar" {
		t.Errorf("got %q, want FooBar", got)
	}
}

func TestUnicode(t *testing.T) {
	if got := CamelCase("straße-test"); got != "straßeTest" {
		t.Errorf("got %q, want straßeTest", got)
	}
}
