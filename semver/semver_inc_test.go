package semver

import "testing"

// TestIncAllReleaseLevels covers every release level npm semver's inc()
// accepts. The pre* levels and "release" were previously missing and returned
// an "unknown release level" error. Vectors follow node-semver's test/increments.js.
func TestIncAllReleaseLevels(t *testing.T) {
	cases := []struct {
		in      string
		release string
		want    string
	}{
		// Plain levels.
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "patch", "1.2.4"},
		// A prerelease is promoted rather than bumped.
		{"1.2.4-0", "patch", "1.2.4"},
		{"1.2.0-0", "minor", "1.2.0"},
		{"1.0.0-0", "major", "1.0.0"},
		{"1.2.3-4", "major", "2.0.0"},
		{"1.2.3-4", "minor", "1.3.0"},

		// pre* levels.
		{"1.2.3", "premajor", "2.0.0-0"},
		{"1.2.3", "preminor", "1.3.0-0"},
		{"1.2.3", "prepatch", "1.2.4-0"},
		{"1.2.0-1", "premajor", "2.0.0-0"},
		{"1.2.0-1", "preminor", "1.3.0-0"},
		{"1.2.0-1", "prepatch", "1.2.1-0"},

		// prerelease: acts like prepatch on a release version, otherwise bumps
		// the trailing numeric identifier.
		{"1.2.3", "prerelease", "1.2.4-0"},
		{"1.2.4-0", "prerelease", "1.2.4-1"},
		{"1.2.4-1", "prerelease", "1.2.4-2"},
		{"1.2.4-beta.0", "prerelease", "1.2.4-beta.1"},
		{"1.2.4-beta.9", "prerelease", "1.2.4-beta.10"},
		// No numeric identifier to bump: one is appended.
		{"1.2.4-beta", "prerelease", "1.2.4-beta.0"},
		{"1.2.4-alpha.beta", "prerelease", "1.2.4-alpha.beta.0"},

		// release strips the prerelease.
		{"1.2.3-beta.1", "release", "1.2.3"},
		{"1.2.3-0", "release", "1.2.3"},
	}
	for _, c := range cases {
		got, err := Inc(c.in, c.release)
		if err != nil {
			t.Errorf("Inc(%q, %q): %v", c.in, c.release, err)
			continue
		}
		if got != c.want {
			t.Errorf("Inc(%q, %q) = %q; want %q", c.in, c.release, got, c.want)
		}
	}
}

// TestIncWithIdentifier covers the identifier argument of npm semver's
// inc(version, release, identifier).
func TestIncWithIdentifier(t *testing.T) {
	cases := []struct {
		in         string
		release    string
		identifier string
		want       string
	}{
		{"1.2.3", "premajor", "beta", "2.0.0-beta.0"},
		{"1.2.3", "preminor", "beta", "1.3.0-beta.0"},
		{"1.2.3", "prepatch", "beta", "1.2.4-beta.0"},
		{"1.2.3", "prerelease", "beta", "1.2.4-beta.0"},
		{"1.2.4-beta.0", "prerelease", "beta", "1.2.4-beta.1"},
		// A different identifier restarts the counter.
		{"1.2.4-alpha.3", "prerelease", "beta", "1.2.4-beta.0"},
		// A non-numeric counter under the same identifier restarts too.
		{"1.2.4-beta.x", "prerelease", "beta", "1.2.4-beta.0"},
		// The identifier is ignored by the plain levels, as upstream does.
		{"1.2.3", "patch", "beta", "1.2.4"},
	}
	for _, c := range cases {
		got, err := IncWith(c.in, c.release, c.identifier)
		if err != nil {
			t.Errorf("IncWith(%q, %q, %q): %v", c.in, c.release, c.identifier, err)
			continue
		}
		if got != c.want {
			t.Errorf("IncWith(%q, %q, %q) = %q; want %q", c.in, c.release, c.identifier, got, c.want)
		}
	}
}

// TestIncErrors covers the error paths.
func TestIncErrors(t *testing.T) {
	if _, err := Inc("1.2.3", "release"); err == nil {
		t.Error(`Inc("1.2.3", "release"): want an error, 1.2.3 is not a prerelease`)
	}
	for _, level := range []string{"", "bogus", "PATCH", "pre"} {
		if _, err := Inc("1.2.3", level); err == nil {
			t.Errorf("Inc(1.2.3, %q): want an error", level)
		}
	}
	for _, v := range []string{"", "1.2", "not a version", "1.2.3.4"} {
		if _, err := Inc(v, "patch"); err == nil {
			t.Errorf("Inc(%q, patch): want an error", v)
		}
	}
}

// TestIncMonotonic confirms each increment produces a strictly higher version,
// which is the property callers actually depend on.
func TestIncMonotonic(t *testing.T) {
	levels := []string{"prerelease", "prerelease", "patch", "preminor", "prerelease", "minor", "premajor", "major"}
	cur := "1.2.3"
	for _, level := range levels {
		next, err := Inc(cur, level)
		if err != nil {
			t.Fatalf("Inc(%q, %q): %v", cur, level, err)
		}
		if Compare(next, cur) <= 0 {
			t.Fatalf("Inc(%q, %q) = %q which does not sort higher", cur, level, next)
		}
		cur = next
	}
}
