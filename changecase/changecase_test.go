package changecase

import (
	"reflect"
	"testing"
)

func TestWords(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"fooBar", []string{"foo", "Bar"}},
		{"foo_bar-baz", []string{"foo", "bar", "baz"}},
		{"HTTPServer", []string{"HTTP", "Server"}},
		{"version1Point2", []string{"version1", "Point2"}},
		{"XMLHttpRequest", []string{"XML", "Http", "Request"}},
		{"  hello  world ", []string{"hello", "world"}},
	}
	for _, tt := range tests {
		if got := Words(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Words(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestConverters(t *testing.T) {
	const in = "fooBarBaz"
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"NoCase", NoCase(in), "foo bar baz"},
		{"CamelCase", CamelCase("foo-bar-baz"), "fooBarBaz"},
		{"PascalCase", PascalCase("foo_bar_baz"), "FooBarBaz"},
		{"SnakeCase", SnakeCase(in), "foo_bar_baz"},
		{"KebabCase", KebabCase(in), "foo-bar-baz"},
		{"ParamCase", ParamCase(in), "foo-bar-baz"},
		{"ConstantCase", ConstantCase(in), "FOO_BAR_BAZ"},
		{"DotCase", DotCase(in), "foo.bar.baz"},
		{"PathCase", PathCase(in), "foo/bar/baz"},
		{"HeaderCase", HeaderCase(in), "Foo-Bar-Baz"},
		{"CapitalCase", CapitalCase(in), "Foo Bar Baz"},
		{"SentenceCase", SentenceCase(in), "Foo bar baz"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestAcronyms(t *testing.T) {
	if got := SnakeCase("XMLHttpRequest"); got != "xml_http_request" {
		t.Errorf("SnakeCase acronym = %q", got)
	}
	if got := CamelCase("HTTP_SERVER"); got != "httpServer" {
		t.Errorf("CamelCase from constant = %q", got)
	}
}

func TestSwapCase(t *testing.T) {
	if got := SwapCase("Hello World"); got != "hELLO wORLD" {
		t.Errorf("SwapCase = %q", got)
	}
	if got := SwapCase("abc123XYZ"); got != "ABC123xyz" {
		t.Errorf("SwapCase mixed = %q", got)
	}
}

func TestEmpty(t *testing.T) {
	if CamelCase("") != "" || PascalCase("") != "" || SentenceCase("") != "" {
		t.Error("empty input")
	}
}

func BenchmarkSnakeCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SnakeCase("XMLHttpRequestParserV2")
	}
}
