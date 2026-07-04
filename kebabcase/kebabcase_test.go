package kebabcase

import "testing"

func TestKebabCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"fooBar Baz", "foo-bar-baz"},
		{"fooBar", "foo-bar"},
		{"foo bar", "foo-bar"},
		{"foo_bar", "foo-bar"},
		{"foo-bar", "foo-bar"},
		{"FooBar", "foo-bar"},
		{"fooBarBaz", "foo-bar-baz"},
		{"foo123Bar", "foo123-bar"},
		{"FOOBar", "foobar"},
		{"foo   bar", "foo-bar"},
		{"__foo__bar__", "foo-bar"},
		{"--foo--bar--", "foo-bar"},
		{"foo", "foo"},
		{"FOO", "foo"},
		{"", ""},
		{"   ", ""},
		{"HelloWorld", "hello-world"},
		{"already-kebab-case", "already-kebab-case"},
		{"XMLHttpRequest", "xmlhttp-request"},
	}
	for _, c := range cases {
		if got := KebabCase(c.in); got != c.want {
			t.Errorf("KebabCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
