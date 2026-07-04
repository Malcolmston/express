package globtoregexp

import "testing"

func TestGlobToRegExpDefault(t *testing.T) {
	// Default: globstar OFF, so "*" matches everything including separators.
	tests := []struct {
		glob    string
		input   string
		matches bool
	}{
		{"*.js", "foo.js", true},
		{"*.js", "foo/bar.js", true}, // default * crosses separators
		{"*.js", "foo.txt", false},
		{"foo", "foo", true},
		{"foo", "foobar", false},
		{"a?b", "a?b", true},  // ? is literal when not extended
		{"a?b", "axb", false}, // ? does not match a single char by default
		{"a.b", "a.b", true},  // . is literal
		{"a.b", "axb", false},
		{"/foo/*", "/foo/bar", true},
		{"/foo/*", "/foo/bar/baz", true}, // default crosses separators
	}
	for _, tt := range tests {
		re, err := GlobToRegExp(tt.glob)
		if err != nil {
			t.Fatalf("GlobToRegExp(%q) error: %v", tt.glob, err)
		}
		if got := re.MatchString(tt.input); got != tt.matches {
			t.Errorf("glob %q input %q: got match=%v want %v (pattern %q)", tt.glob, tt.input, got, tt.matches, re.String())
		}
	}
}

func TestGlobToRegExpGlobStar(t *testing.T) {
	opts := Options{GlobStar: true}
	tests := []struct {
		glob    string
		input   string
		matches bool
	}{
		{"*.js", "foo.js", true},
		{"*.js", "foo/bar.js", false}, // single * does not cross separators
		{"a/**/b", "a/x/y/b", true},
		{"a/**/b", "a/x/b", true},
		{"a/**/b", "a/b", true}, // globstar matches zero segments
		{"a/*/b", "a/x/b", true},
		{"a/*/b", "a/x/y/b", false},
		{"/foo/*", "/foo/bar", true},
		{"/foo/*", "/foo/bar/baz", false},
		{"**", "anything/at/all", true},
	}
	for _, tt := range tests {
		re, err := GlobToRegExpOpts(tt.glob, opts)
		if err != nil {
			t.Fatalf("GlobToRegExpOpts(%q) error: %v", tt.glob, err)
		}
		if got := re.MatchString(tt.input); got != tt.matches {
			t.Errorf("glob %q input %q: got match=%v want %v (pattern %q)", tt.glob, tt.input, got, tt.matches, re.String())
		}
	}
}

func TestGlobToRegExpExtended(t *testing.T) {
	opts := Options{Extended: true, GlobStar: true}
	tests := []struct {
		glob    string
		input   string
		matches bool
	}{
		{"a?c", "abc", true}, // ? matches single char in extended mode
		{"a?c", "ac", false},
		{"[abc].js", "a.js", true},
		{"[abc].js", "d.js", false},
		{"*.{js,ts}", "foo.js", true},
		{"*.{js,ts}", "foo.ts", true},
		{"*.{js,ts}", "foo.md", false},
	}
	for _, tt := range tests {
		re, err := GlobToRegExpOpts(tt.glob, opts)
		if err != nil {
			t.Fatalf("GlobToRegExpOpts(%q) error: %v", tt.glob, err)
		}
		if got := re.MatchString(tt.input); got != tt.matches {
			t.Errorf("glob %q input %q: got match=%v want %v (pattern %q)", tt.glob, tt.input, got, tt.matches, re.String())
		}
	}
}

func TestGlobToRegExpFlags(t *testing.T) {
	re, err := GlobToRegExpOpts("*.JS", Options{GlobStar: true, Flags: "i"})
	if err != nil {
		t.Fatal(err)
	}
	if !re.MatchString("foo.js") {
		t.Errorf("case-insensitive flag: expected %q to match foo.js", re.String())
	}

	// "g" flag disables anchoring.
	re2, err := GlobToRegExpOpts("bar", Options{Flags: "g"})
	if err != nil {
		t.Fatal(err)
	}
	if !re2.MatchString("foobarbaz") {
		t.Errorf("g flag should disable anchoring; pattern %q did not match substring", re2.String())
	}
	if re2.String() != "bar" {
		t.Errorf("expected unanchored pattern %q", re2.String())
	}
}

func TestGlobToRegExpPattern(t *testing.T) {
	// Verify the exact generated pattern for a couple of cases.
	re, _ := GlobToRegExp("*.js")
	if re.String() != `^.*\.js$` {
		t.Errorf("default *.js pattern = %q, want %q", re.String(), `^.*\.js$`)
	}
	re2, _ := GlobToRegExpOpts("*.js", Options{GlobStar: true})
	if re2.String() != `^([^/]*)\.js$` {
		t.Errorf("globstar *.js pattern = %q, want %q", re2.String(), `^([^/]*)\.js$`)
	}
}
