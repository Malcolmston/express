package changecase

// Upstream-parity vectors transcribed verbatim from the original
// blakeembrey/change-case test suite:
//
//	https://raw.githubusercontent.com/blakeembrey/change-case/main/packages/change-case/src/index.spec.ts
//
// The upstream `tests` table drives every transform from a shared list of
// inputs to an expected result per case. Only the rows that use the DEFAULT
// options are reproduced here, because this Go port exposes the no-option API
// surface. Rows that exercise upstream Options (delimiter, separateNumbers,
// prefixCharacters/suffixCharacters, mergeAmbiguousCharacters) are recorded as
// gaps in the sync notes rather than encoded as vectors.
//
// Transform name mapping (upstream field -> Go export):
//
//	camelCase->CamelCase, capitalCase->CapitalCase, constantCase->ConstantCase,
//	dotCase->DotCase, kebabCase->KebabCase, noCase->NoCase, pascalCase->PascalCase,
//	pathCase->PathCase, sentenceCase->SentenceCase, snakeCase->SnakeCase,
//	trainCase->HeaderCase. (pascalSnakeCase has no Go equivalent and is omitted.)

import "testing"

type parityResult struct {
	camel, capital, constant, dot, kebab, no, pascal, path, sentence, snake, train string
}

// parityCases are the default-option rows from the upstream TEST_CASES table.
var parityCases = []struct {
	in string
	r  parityResult
}{
	{"", parityResult{"", "", "", "", "", "", "", "", "", "", ""}},
	{"test", parityResult{"test", "Test", "TEST", "test", "test", "test", "Test", "test", "Test", "test", "Test"}},
	{"test string", parityResult{"testString", "Test String", "TEST_STRING", "test.string", "test-string", "test string", "TestString", "test/string", "Test string", "test_string", "Test-String"}},
	{"Test String", parityResult{"testString", "Test String", "TEST_STRING", "test.string", "test-string", "test string", "TestString", "test/string", "Test string", "test_string", "Test-String"}},
	{"TestV2", parityResult{"testV2", "Test V2", "TEST_V2", "test.v2", "test-v2", "test v2", "TestV2", "test/v2", "Test v2", "test_v2", "Test-V2"}},
	{"_foo_bar_", parityResult{"fooBar", "Foo Bar", "FOO_BAR", "foo.bar", "foo-bar", "foo bar", "FooBar", "foo/bar", "Foo bar", "foo_bar", "Foo-Bar"}},
	{"version 1.2.10", parityResult{"version_1_2_10", "Version 1 2 10", "VERSION_1_2_10", "version.1.2.10", "version-1-2-10", "version 1 2 10", "Version_1_2_10", "version/1/2/10", "Version 1 2 10", "version_1_2_10", "Version-1-2-10"}},
	{"version 1.21.0", parityResult{"version_1_21_0", "Version 1 21 0", "VERSION_1_21_0", "version.1.21.0", "version-1-21-0", "version 1 21 0", "Version_1_21_0", "version/1/21/0", "Version 1 21 0", "version_1_21_0", "Version-1-21-0"}},
	{"V1Test", parityResult{"v1Test", "V1 Test", "V1_TEST", "v1.test", "v1-test", "v1 test", "V1Test", "v1/test", "V1 test", "v1_test", "V1-Test"}},
}

func TestParityChangeCase(t *testing.T) {
	transforms := []struct {
		name string
		fn   func(string) string
		want func(parityResult) string
	}{
		{"CamelCase", CamelCase, func(r parityResult) string { return r.camel }},
		{"CapitalCase", CapitalCase, func(r parityResult) string { return r.capital }},
		{"ConstantCase", ConstantCase, func(r parityResult) string { return r.constant }},
		{"DotCase", DotCase, func(r parityResult) string { return r.dot }},
		{"KebabCase", KebabCase, func(r parityResult) string { return r.kebab }},
		{"NoCase", NoCase, func(r parityResult) string { return r.no }},
		{"PascalCase", PascalCase, func(r parityResult) string { return r.pascal }},
		{"PathCase", PathCase, func(r parityResult) string { return r.path }},
		{"SentenceCase", SentenceCase, func(r parityResult) string { return r.sentence }},
		{"SnakeCase", SnakeCase, func(r parityResult) string { return r.snake }},
		{"HeaderCase", HeaderCase, func(r parityResult) string { return r.train }},
	}
	for _, tc := range parityCases {
		for _, tr := range transforms {
			want := tr.want(tc.r)
			if got := tr.fn(tc.in); got != want {
				t.Errorf("%s(%q) = %q, want %q", tr.name, tc.in, got, want)
			}
		}
	}
}

// TestParityParamCaseAlias confirms ParamCase mirrors upstream kebabCase, which
// the original library exposes as a paramCase alias.
func TestParityParamCaseAlias(t *testing.T) {
	for _, tc := range parityCases {
		if got := ParamCase(tc.in); got != tc.r.kebab {
			t.Errorf("ParamCase(%q) = %q, want %q", tc.in, got, tc.r.kebab)
		}
	}
}
