package titlecase

import "testing"

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple sentence", "a simple test", "A Simple Test"},
		{"hello world", "hello world", "Hello World"},
		{"already title", "Hello World", "Hello World"},
		{"all caps", "HELLO WORLD", "Hello World"},
		{"camelCase", "helloWorld", "Hello World"},
		{"PascalCase", "HelloWorld", "Hello World"},
		{"snake_case", "foo_bar_baz", "Foo Bar Baz"},
		{"kebab-case", "foo-bar-baz", "Foo Bar Baz"},
		{"mixed separators", "foo_bar-baz qux", "Foo Bar Baz Qux"},
		{"acronym then word", "XMLHttpRequest", "Xml Http Request"},
		{"with digits", "version2Point0", "Version2 Point0"},
		{"leading spaces", "  hello   world  ", "Hello World"},
		{"single word", "test", "Test"},
		{"empty", "", ""},
		{"punctuation", "hello, world!", "Hello World"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TitleCase(tt.in)
			if got != tt.want {
				t.Fatalf("TitleCase(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
