package semver

// Upstream-parity tests for the npm "semver" package (npm/node-semver).
//
// Every vector below is copied verbatim from node-semver's own test fixtures on
// the "main" branch. Only default-mode vectors are encoded here: node-semver
// entries carrying a truthy "loose" option, an "includePrerelease" option, or
// the prerelease/prepatch/preminor/premajor/release increment identifiers
// exercise modes this stdlib-only port does not implement (see the package doc)
// and are intentionally omitted; see the mid-file notes for the excluded blocks.
//
// Sources (raw.githubusercontent.com):
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/comparisons.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/equality.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/increments.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/invalid-versions.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/range-include.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/fixtures/range-exclude.js
//   https://raw.githubusercontent.com/npm/node-semver/main/test/functions/coerce.js
//   https://raw.githubusercontent.com/npm/node-semver/main/internal/constants.js

import (
	"strings"
	"testing"
)

// TestParityComparisons mirrors test/fixtures/comparisons.js: for each pair,
// version1 has strictly greater precedence than version2 (the loose flag in the
// upstream fixture does not affect these strict-valid versions).
func TestParityComparisons(t *testing.T) {
	pairs := [][2]string{
		{"0.0.0", "0.0.0-foo"},
		{"0.0.1", "0.0.0"},
		{"1.0.0", "0.9.9"},
		{"0.10.0", "0.9.0"},
		{"0.99.0", "0.10.0"},
		{"2.0.0", "1.2.3"},
		{"v0.0.0", "0.0.0-foo"},
		{"v0.0.1", "0.0.0"},
		{"v1.0.0", "0.9.9"},
		{"v0.10.0", "0.9.0"},
		{"v0.99.0", "0.10.0"},
		{"v2.0.0", "1.2.3"},
		{"0.0.0", "v0.0.0-foo"},
		{"0.0.1", "v0.0.0"},
		{"1.0.0", "v0.9.9"},
		{"0.10.0", "v0.9.0"},
		{"0.99.0", "v0.10.0"},
		{"2.0.0", "v1.2.3"},
		{"1.2.3", "1.2.3-asdf"},
		{"1.2.3", "1.2.3-4"},
		{"1.2.3", "1.2.3-4-foo"},
		{"1.2.3-5-foo", "1.2.3-5"},
		{"1.2.3-5", "1.2.3-4"},
		{"1.2.3-5-foo", "1.2.3-5-Foo"},
		{"3.0.0", "2.7.2+asdf"},
		{"1.2.3-a.10", "1.2.3-a.5"},
		{"1.2.3-a.b", "1.2.3-a.5"},
		{"1.2.3-a.b", "1.2.3-a"},
		{"1.2.3-a.b.c.10.d.5", "1.2.3-a.b.c.5.d.100"},
		{"1.2.3-r2", "1.2.3-r100"},
		{"1.2.3-r100", "1.2.3-R2"},
	}
	for _, p := range pairs {
		hi, lo := p[0], p[1]
		if !GT(hi, lo) {
			t.Errorf("GT(%q,%q) = false, want true", hi, lo)
		}
		if !LT(lo, hi) {
			t.Errorf("LT(%q,%q) = false, want true", lo, hi)
		}
		if EQ(hi, lo) {
			t.Errorf("EQ(%q,%q) = true, want false", hi, lo)
		}
	}
}

// TestParityEquality mirrors the strict-valid entries of test/fixtures/equality.js
// (version1 equals version2, build metadata ignored). Upstream entries with an
// internal space after the "v"/"=" prefix require loose mode and are omitted.
func TestParityEquality(t *testing.T) {
	pairs := [][2]string{
		{"1.2.3", "v1.2.3"},
		{"1.2.3", "=1.2.3"},
		{"1.2.3", " v1.2.3"},
		{"1.2.3", " =1.2.3"},
		{"1.2.3-0", "v1.2.3-0"},
		{"1.2.3-0", "=1.2.3-0"},
		{"1.2.3-1", "v1.2.3-1"},
		{"1.2.3-1", "=1.2.3-1"},
		{"1.2.3-beta", "v1.2.3-beta"},
		{"1.2.3-beta", "=1.2.3-beta"},
		{"1.2.3-beta+build", "1.2.3-beta+otherbuild"},
		{"1.2.3+build", "1.2.3+otherbuild"},
		{"  v1.2.3+build", "1.2.3+otherbuild"},
	}
	for _, p := range pairs {
		if !EQ(p[0], p[1]) {
			t.Errorf("EQ(%q,%q) = false, want true", p[0], p[1])
		}
	}
}

// TestParityIncrement mirrors the major/minor/patch entries of
// test/fixtures/increments.js. Prerelease-family increments (prerelease,
// prepatch, preminor, premajor, release) and identifier arguments require an
// Inc signature this port does not expose and are omitted.
func TestParityIncrement(t *testing.T) {
	cases := []struct{ in, what, want string }{
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3-tag", "major", "2.0.0"},
		{"1.2.0-0", "patch", "1.2.0"},
		{"1.2.3-4", "major", "2.0.0"},
		{"1.2.3-4", "minor", "1.3.0"},
		{"1.2.3-4", "patch", "1.2.3"},
		{"1.2.3-alpha.0.beta", "major", "2.0.0"},
		{"1.2.3-alpha.0.beta", "minor", "1.3.0"},
		{"1.2.3-alpha.0.beta", "patch", "1.2.3"},
		{"1.2.0-1", "minor", "1.2.0"},
		{"1.0.0-1", "major", "1.0.0"},
	}
	for _, c := range cases {
		got, err := Inc(c.in, c.what)
		if err != nil {
			t.Errorf("Inc(%q,%q) error: %v", c.in, c.what, err)
			continue
		}
		if got != c.want {
			t.Errorf("Inc(%q,%q) = %q, want %q", c.in, c.what, got, c.want)
		}
	}
	// null results in the fixture: unknown release level or invalid version.
	for _, c := range []struct{ in, what string }{
		{"1.2.3", "fake"},
		{"fake", "major"},
	} {
		if _, err := Inc(c.in, c.what); err == nil {
			t.Errorf("Inc(%q,%q) = nil error, want error", c.in, c.what)
		}
	}
}

// TestParityInvalidVersions mirrors the string entries of
// test/fixtures/invalid-versions.js (non-string inputs such as regexps/objects
// are JS-specific and omitted). MAX_SAFE_INTEGER is 9007199254740991 and
// MAX_LENGTH is 256, from internal/constants.js.
func TestParityInvalidVersions(t *testing.T) {
	const maxSafeInteger = "9007199254740991"
	invalid := []string{
		strings.Repeat("1", 255) + ".0.0", // "too long" (MAX_LENGTH join)
		maxSafeInteger + "0.0.0",          // "too big"
		"0." + maxSafeInteger + "0.0",     // "too big"
		"0.0." + maxSafeInteger + "0",     // "too big"
		"hello, world",
		"xyz",
	}
	for _, v := range invalid {
		if Valid(v) {
			t.Errorf("Valid(%q) = true, want false", v)
		}
	}
}

// TestParityCoerce mirrors the default (no-option) entries of
// test/functions/coerce.js. rtl and includePrerelease entries are omitted.
func TestParityCoerce(t *testing.T) {
	valid := []struct{ in, want string }{
		{".1", "1.0.0"},
		{".1.", "1.0.0"},
		{"..1", "1.0.0"},
		{".1.1", "1.1.0"},
		{"1.", "1.0.0"},
		{"1.0", "1.0.0"},
		{"1.0.0", "1.0.0"},
		{"0", "0.0.0"},
		{"0.0", "0.0.0"},
		{"0.0.0", "0.0.0"},
		{"0.1", "0.1.0"},
		{"0.0.1", "0.0.1"},
		{"0.1.1", "0.1.1"},
		{"1", "1.0.0"},
		{"1.2", "1.2.0"},
		{"1.2.3", "1.2.3"},
		{"1.2.3.4", "1.2.3"},
		{"13", "13.0.0"},
		{"35.12", "35.12.0"},
		{"35.12.18", "35.12.18"},
		{"35.12.18.24", "35.12.18"},
		{"v1", "1.0.0"},
		{"v1.2", "1.2.0"},
		{"v1.2.3", "1.2.3"},
		{"v1.2.3.4", "1.2.3"},
		{" 1", "1.0.0"},
		{"1 ", "1.0.0"},
		{"1 0", "1.0.0"},
		{"1 1", "1.0.0"},
		{"1.1 1", "1.1.0"},
		{"1.1-1", "1.1.0"},
		{"a1", "1.0.0"},
		{"a1a", "1.0.0"},
		{"1a", "1.0.0"},
		{"version 1", "1.0.0"},
		{"version1", "1.0.0"},
		{"version1.0", "1.0.0"},
		{"version1.1", "1.1.0"},
		{"42.6.7.9.3-alpha", "42.6.7"},
		{"v2", "2.0.0"},
		{"v3.4 replaces v3.3.1", "3.4.0"},
		{"4.6.3.9.2-alpha2", "4.6.3"},
		{strings.Repeat("1", 17) + ".2", "2.0.0"},
		{strings.Repeat("1", 17) + ".2.3", "2.3.0"},
		{"1." + strings.Repeat("2", 17) + ".3", "1.0.0"},
		{"1.2." + strings.Repeat("3", 17), "1.2.0"},
		{strings.Repeat("1", 17) + ".2.3.4", "2.3.4"},
		{"1." + strings.Repeat("2", 17) + ".3.4", "1.0.0"},
		{"1.2." + strings.Repeat("3", 17) + ".4", "1.2.0"},
		{strings.Repeat("1", 17) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("3", 16),
			strings.Repeat("2", 16) + "." + strings.Repeat("3", 16) + ".0"},
		{strings.Repeat("1", 16) + "." + strings.Repeat("2", 17) + "." + strings.Repeat("3", 16),
			strings.Repeat("1", 16) + ".0.0"},
		{strings.Repeat("1", 16) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("3", 17),
			strings.Repeat("1", 16) + "." + strings.Repeat("2", 16) + ".0"},
		{"11" + strings.Repeat(".1", 126), "11.1.1"},
		{strings.Repeat("1", 16), strings.Repeat("1", 16) + ".0.0"},
		{"a" + strings.Repeat("1", 16), strings.Repeat("1", 16) + ".0.0"},
		{strings.Repeat("1", 16) + ".2.3.4", strings.Repeat("1", 16) + ".2.3"},
		{"1." + strings.Repeat("2", 16) + ".3.4", "1." + strings.Repeat("2", 16) + ".3"},
		{"1.2." + strings.Repeat("3", 16) + ".4", "1.2." + strings.Repeat("3", 16)},
		{strings.Repeat("1", 16) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("3", 16),
			strings.Repeat("1", 16) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("3", 16)},
		{"1.2.3." + strings.Repeat("4", 252) + ".5", "1.2.3"},
		{"1.2.3." + strings.Repeat("4", 1024), "1.2.3"},
		{strings.Repeat("1", 17) + ".4.7.4", "4.7.4"},
	}
	for _, c := range valid {
		got, err := Coerce(c.in)
		if err != nil {
			t.Errorf("Coerce(%q) error: %v (want %q)", c.in, err, c.want)
			continue
		}
		if got != c.want {
			t.Errorf("Coerce(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	// Entries expected to coerce to null (cannot be coerced).
	null := []string{
		"",
		".",
		"version one",
		strings.Repeat("9", 16),
		strings.Repeat("1", 17),
		"a" + strings.Repeat("9", 16),
		"a" + strings.Repeat("1", 17),
		strings.Repeat("9", 16) + "a",
		strings.Repeat("1", 17) + "a",
		strings.Repeat("9", 16) + ".4.7.4",
		strings.Repeat("9", 16) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("3", 16),
		strings.Repeat("1", 16) + "." + strings.Repeat("9", 16) + "." + strings.Repeat("3", 16),
		strings.Repeat("1", 16) + "." + strings.Repeat("2", 16) + "." + strings.Repeat("9", 16),
	}
	for _, in := range null {
		if got, err := Coerce(in); err == nil {
			t.Errorf("Coerce(%q) = %q, nil error; want error (null)", in, got)
		}
	}
}

// TestParityRangeInclude mirrors the default (no-option) entries of
// test/fixtures/range-include.js: version should satisfy range.
func TestParityRangeInclude(t *testing.T) {
	cases := [][2]string{
		{"1.0.0 - 2.0.0", "1.2.3"},
		{"^1.2.3+build", "1.2.3"},
		{"^1.2.3+build", "1.3.0"},
		{"1.2.3-pre+asdf - 2.4.3-pre+asdf", "1.2.3"},
		{"1.2.3-pre+asdf - 2.4.3-pre+asdf", "1.2.3-pre.2"},
		{"1.2.3-pre+asdf - 2.4.3-pre+asdf", "2.4.3-alpha"},
		{"1.2.3+asdf - 2.4.3+asdf", "1.2.3"},
		{"1.0.0", "1.0.0"},
		{">=*", "0.2.4"},
		{"", "1.0.0"},
		{"*", "1.2.3"},
		{">=1.0.0", "1.0.0"},
		{">=1.0.0", "1.0.1"},
		{">=1.0.0", "1.1.0"},
		{">1.0.0", "1.0.1"},
		{">1.0.0", "1.1.0"},
		{"<=2.0.0", "2.0.0"},
		{"<=2.0.0", "1.9999.9999"},
		{"<=2.0.0", "0.2.9"},
		{"<2.0.0", "1.9999.9999"},
		{"<2.0.0", "0.2.9"},
		{">= 1.0.0", "1.0.0"},
		{">=  1.0.0", "1.0.1"},
		{">=   1.0.0", "1.1.0"},
		{"> 1.0.0", "1.0.1"},
		{">  1.0.0", "1.1.0"},
		{"<=   2.0.0", "2.0.0"},
		{"<= 2.0.0", "1.9999.9999"},
		{"<=  2.0.0", "0.2.9"},
		{"<    2.0.0", "1.9999.9999"},
		{"<\t2.0.0", "0.2.9"},
		{">=0.1.97", "0.1.97"},
		{"0.1.20 || 1.2.4", "1.2.4"},
		{">=0.2.3 || <0.0.1", "0.0.0"},
		{">=0.2.3 || <0.0.1", "0.2.3"},
		{">=0.2.3 || <0.0.1", "0.2.4"},
		{"||", "1.3.4"},
		{"2.x.x", "2.1.3"},
		{"1.2.x", "1.2.3"},
		{"1.2.x || 2.x", "2.1.3"},
		{"1.2.x || 2.x", "1.2.3"},
		{"x", "1.2.3"},
		{"2.*.*", "2.1.3"},
		{"1.2.*", "1.2.3"},
		{"1.2.* || 2.*", "2.1.3"},
		{"1.2.* || 2.*", "1.2.3"},
		{"2", "2.1.2"},
		{"2.3", "2.3.1"},
		{"~0.0.1", "0.0.1"},
		{"~0.0.1", "0.0.2"},
		{"~x", "0.0.9"},
		{"~2", "2.0.9"},
		{"~2.4", "2.4.0"},
		{"~2.4", "2.4.5"},
		{"~>3.2.1", "3.2.2"},
		{"~1", "1.2.3"},
		{"~>1", "1.2.3"},
		{"~> 1", "1.2.3"},
		{"~1.0", "1.0.2"},
		{"~ 1.0", "1.0.2"},
		{"~ 1.0.3", "1.0.12"},
		{">=1", "1.0.0"},
		{">= 1", "1.0.0"},
		{"<1.2", "1.1.1"},
		{"< 1.2", "1.1.1"},
		{"~v0.5.4-pre", "0.5.5"},
		{"~v0.5.4-pre", "0.5.4"},
		{"=0.7.x", "0.7.2"},
		{"<=0.7.x", "0.7.2"},
		{">=0.7.x", "0.7.2"},
		{"<=0.7.x", "0.6.2"},
		{"~1.2.1 >=1.2.3", "1.2.3"},
		{"~1.2.1 =1.2.3", "1.2.3"},
		{"~1.2.1 1.2.3", "1.2.3"},
		{"~1.2.1 >=1.2.3 1.2.3", "1.2.3"},
		{"~1.2.1 1.2.3 >=1.2.3", "1.2.3"},
		{">=1.2.1 1.2.3", "1.2.3"},
		{"1.2.3 >=1.2.1", "1.2.3"},
		{">=1.2.3 >=1.2.1", "1.2.3"},
		{">=1.2.1 >=1.2.3", "1.2.3"},
		{">=1.2", "1.2.8"},
		{"^1.2.3", "1.8.1"},
		{"^0.1.2", "0.1.2"},
		{"^0.1", "0.1.2"},
		{"^0.0.1", "0.0.1"},
		{"^1.2", "1.4.2"},
		{"^1.2 ^1", "1.4.2"},
		{"^1.2.3-alpha", "1.2.3-pre"},
		{"^1.2.0-alpha", "1.2.0-pre"},
		{"^0.0.1-alpha", "0.0.1-beta"},
		{"^0.0.1-alpha", "0.0.1"},
		{"^0.1.1-alpha", "0.1.1-beta"},
		{"^x", "1.2.3"},
		{"x - 1.0.0", "0.9.7"},
		{"x - 1.x", "0.9.7"},
		{"1.0.0 - x", "1.9.7"},
		{"1.x - x", "1.9.7"},
		{"<=7.x", "7.9.9"},
	}
	for _, c := range cases {
		if !Satisfies(c[1], c[0]) {
			t.Errorf("Satisfies(%q, %q) = false, want true", c[1], c[0])
		}
	}
}

// TestParityRangeExclude mirrors the default (no-option) entries of
// test/fixtures/range-exclude.js: version should not satisfy range.
func TestParityRangeExclude(t *testing.T) {
	cases := [][2]string{
		{"1.0.0 - 2.0.0", "2.2.3"},
		{"1.2.3+asdf - 2.4.3+asdf", "1.2.3-pre.2"},
		{"1.2.3+asdf - 2.4.3+asdf", "2.4.3-alpha"},
		{"^1.2.3+build", "2.0.0"},
		{"^1.2.3+build", "1.2.0"},
		{"^1.2.3", "1.2.3-pre"},
		{"^1.2", "1.2.0-pre"},
		{">1.2", "1.3.0-beta"},
		{"<=1.2.3", "1.2.3-beta"},
		{"^1.2.3", "1.2.3-beta"},
		{"=0.7.x", "0.7.0-asdf"},
		{">=0.7.x", "0.7.0-asdf"},
		{"<=0.7.x", "0.7.0-asdf"},
		{"1.0.0", "1.0.1"},
		{">=1.0.0", "0.0.0"},
		{">=1.0.0", "0.0.1"},
		{">=1.0.0", "0.1.0"},
		{">1.0.0", "0.0.1"},
		{">1.0.0", "0.1.0"},
		{"<=2.0.0", "3.0.0"},
		{"<=2.0.0", "2.9999.9999"},
		{"<=2.0.0", "2.2.9"},
		{"<2.0.0", "2.9999.9999"},
		{"<2.0.0", "2.2.9"},
		{">=0.1.97", "0.1.93"},
		{"0.1.20 || 1.2.4", "1.2.3"},
		{">=0.2.3 || <0.0.1", "0.0.3"},
		{">=0.2.3 || <0.0.1", "0.2.2"},
		{"2.x.x", "3.1.3"},
		{"1.2.x", "1.3.3"},
		{"1.2.x || 2.x", "3.1.3"},
		{"1.2.x || 2.x", "1.1.3"},
		{"2.*.*", "1.1.3"},
		{"2.*.*", "3.1.3"},
		{"1.2.*", "1.3.3"},
		{"1.2.* || 2.*", "3.1.3"},
		{"1.2.* || 2.*", "1.1.3"},
		{"2", "1.1.2"},
		{"2.3", "2.4.1"},
		{"~0.0.1", "0.1.0-alpha"},
		{"~0.0.1", "0.1.0"},
		{"~2.4", "2.5.0"},
		{"~2.4", "2.3.9"},
		{"~>3.2.1", "3.3.2"},
		{"~>3.2.1", "3.2.0"},
		{"~1", "0.2.3"},
		{"~>1", "2.2.3"},
		{"~1.0", "1.1.0"},
		{"<1", "1.0.0"},
		{">=1.2", "1.1.1"},
		{"~v0.5.4-beta", "0.5.4-alpha"},
		{"=0.7.x", "0.8.2"},
		{">=0.7.x", "0.6.2"},
		{"<0.7.x", "0.7.2"},
		{"<1.2.3", "1.2.3-beta"},
		{"=1.2.3", "1.2.3-beta"},
		{">1.2", "1.2.8"},
		{"^0.0.1", "0.0.2-alpha"},
		{"^0.0.1", "0.0.2"},
		{"^1.2.3", "2.0.0-alpha"},
		{"^1.2.3", "1.2.2"},
		{"^1.2", "1.1.9"},
		{"*", "not a version"},
		{">=2", "glorp"},
		{">=1.0.0 <1.1.0", "1.1.0"},
		{">=1.0.0 <1.1.0", "1.1.0-pre"},
		{">=1.0.0 <1.1.0-pre", "1.1.0-pre"},
	}
	for _, c := range cases {
		if Satisfies(c[1], c[0]) {
			t.Errorf("Satisfies(%q, %q) = true, want false", c[1], c[0])
		}
	}
}
