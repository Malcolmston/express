package kebabcase

import "testing"

// Upstream-parity vectors for the Go port of blakeembrey/change-case (the
// kebabCase function).
//
// Source: blakeembrey/change-case, packages/change-case/src/index.spec.ts
//
//	https://raw.githubusercontent.com/blakeembrey/change-case/main/packages/change-case/src/index.spec.ts
//
// Algorithm: packages/change-case/src/index.ts (split + noCase with delimiter "-")
//
//	https://raw.githubusercontent.com/blakeembrey/change-case/main/packages/change-case/src/index.ts
//
// The spec's `tests` table maps each input to a `kebabCase` expectation. Only
// the entries that use the DEFAULT options are reproduced here, because this
// port exposes a single no-options KebabCase entry point. Entries that depend
// on non-default options (separateNumbers, a custom delimiter, or
// prefix/suffix characters) are intentionally omitted; where the same input
// also appears with default options its default expectation is used.
//
// Values below are copied verbatim from the upstream spec; they are not
// invented.
func TestParityChangeCaseKebab(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},                             // spec: ""
		{"test", "test"},                     // spec: "test"
		{"test string", "test-string"},       // spec: "test string"
		{"Test String", "test-string"},       // spec: "Test String"
		{"TestV2", "test-v2"},                // spec: "TestV2" (default)
		{"_foo_bar_", "foo-bar"},             // spec: "_foo_bar_"
		{"version 1.2.10", "version-1-2-10"}, // spec: "version 1.2.10"
		{"version 1.21.0", "version-1-21-0"}, // spec: "version 1.21.0"
		{"V1Test", "v1-test"},                // spec: "V1Test" (default)
	}
	for _, c := range cases {
		if got := KebabCase(c.in); got != c.want {
			t.Errorf("KebabCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
