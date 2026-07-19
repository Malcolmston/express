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
		{"all caps preserved", "HELLO WORLD", "HELLO WORLD"},
		{"camelCase preserved", "helloWorld", "HelloWorld"},
		{"PascalCase preserved", "HelloWorld", "HelloWorld"},
		{"snake_case", "foo_bar_baz", "Foo_bar_baz"},
		{"kebab-case", "foo-bar-baz", "Foo-Bar-Baz"},
		{"mixed separators", "foo-bar-baz qux", "Foo-Bar-Baz Qux"},
		{"acronym preserved", "XMLHttpRequest", "XMLHttpRequest"},
		{"with digits", "version2Point0", "Version2Point0"},
		{"leading spaces preserved", "  hello   world  ", "  Hello   World  "},
		{"single word", "test", "Test"},
		{"empty", "", ""},
		{"punctuation preserved", "hello, world!", "Hello, World!"},
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
