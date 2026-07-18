package semver

import (
	"reflect"
	"testing"
)

func TestParseValid(t *testing.T) {
	tests := []struct {
		in                  string
		major, minor, patch uint64
		pre, build          []string
	}{
		{"1.2.3", 1, 2, 3, nil, nil},
		{"v1.2.3", 1, 2, 3, nil, nil},
		{"=1.2.3", 1, 2, 3, nil, nil},
		{" 1.0.0 ", 1, 0, 0, nil, nil},
		{"0.0.0", 0, 0, 0, nil, nil},
		{"1.2.3-alpha.1", 1, 2, 3, []string{"alpha", "1"}, nil},
		{"1.2.3+build.5", 1, 2, 3, nil, []string{"build", "5"}},
		{"1.2.3-rc.1+exp.sha.5114f85", 1, 2, 3, []string{"rc", "1"}, []string{"exp", "sha", "5114f85"}},
		{"10.20.30", 10, 20, 30, nil, nil},
	}
	for _, tt := range tests {
		v, err := Parse(tt.in)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", tt.in, err)
		}
		if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
			t.Errorf("Parse(%q) = %d.%d.%d", tt.in, v.Major, v.Minor, v.Patch)
		}
		if !reflect.DeepEqual(v.Prerelease, tt.pre) {
			t.Errorf("Parse(%q) pre = %v, want %v", tt.in, v.Prerelease, tt.pre)
		}
		if !reflect.DeepEqual(v.Build, tt.build) {
			t.Errorf("Parse(%q) build = %v, want %v", tt.in, v.Build, tt.build)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "1", "1.2", "1.2.3.4", "01.2.3", "1.2.3-", "1.2.3+", "a.b.c", "1.2.3-01", "-1.2.3"} {
		if Valid(in) {
			t.Errorf("Valid(%q) = true, want false", in)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "2.0.1", -1},
		{"2.1.0", "2.0.9", 1},
		{"1.0.0", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-alpha.1", -1},
		{"1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-beta.2", -1},
		{"1.0.0-beta.2", "1.0.0-beta.11", -1},
		{"1.0.0-beta.11", "1.0.0-rc.1", -1},
		{"1.0.0-rc.1", "1.0.0", -1},
		{"1.0.0+build.1", "1.0.0+build.2", 0}, // build ignored
	}
	for _, tt := range tests {
		if got := Compare(tt.a, tt.b); got != tt.want {
			t.Errorf("Compare(%q,%q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestBoolHelpers(t *testing.T) {
	if !GT("2.0.0", "1.0.0") || GT("1.0.0", "2.0.0") {
		t.Error("GT")
	}
	if !GTE("2.0.0", "2.0.0") || !GTE("2.0.1", "2.0.0") {
		t.Error("GTE")
	}
	if !LT("1.0.0", "2.0.0") || LT("2.0.0", "1.0.0") {
		t.Error("LT")
	}
	if !LTE("2.0.0", "2.0.0") || !LTE("1.0.0", "2.0.0") {
		t.Error("LTE")
	}
	if !EQ("1.2.3", "1.2.3+build") || EQ("1.2.3", "1.2.4") {
		t.Error("EQ")
	}
	if !NEQ("1.2.3", "1.2.4") || NEQ("1.2.3", "1.2.3") {
		t.Error("NEQ")
	}
}

func TestSort(t *testing.T) {
	in := []string{"2.0.0", "1.0.0-alpha", "1.0.0", "1.2.0", "1.0.0-beta"}
	Sort(in)
	want := []string{"1.0.0-alpha", "1.0.0-beta", "1.0.0", "1.2.0", "2.0.0"}
	if !reflect.DeepEqual(in, want) {
		t.Errorf("Sort = %v, want %v", in, want)
	}
}

func TestInc(t *testing.T) {
	tests := []struct {
		in, rel, want string
	}{
		{"1.2.3", "major", "2.0.0"},
		{"1.2.3", "minor", "1.3.0"},
		{"1.2.3", "patch", "1.2.4"},
		{"1.2.3-alpha.1", "patch", "1.2.3"},
		{"0.9.9", "minor", "0.10.0"},
	}
	for _, tt := range tests {
		got, err := Inc(tt.in, tt.rel)
		if err != nil {
			t.Fatalf("Inc(%q,%q): %v", tt.in, tt.rel, err)
		}
		if got != tt.want {
			t.Errorf("Inc(%q,%q) = %q, want %q", tt.in, tt.rel, got, tt.want)
		}
	}
	if _, err := Inc("1.2.3", "bogus"); err == nil {
		t.Error("expected error for bad release level")
	}
}

func TestCoerce(t *testing.T) {
	tests := []struct{ in, want string }{
		{"v2", "2.0.0"},
		{"1.2", "1.2.0"},
		{"1.2.3", "1.2.3"},
		{"version 4.5.6 final", "4.5.6"},
		{"=1.2.3.4", "1.2.3"},
		{"~1.2", "1.2.0"},
	}
	for _, tt := range tests {
		got, err := Coerce(tt.in)
		if err != nil {
			t.Fatalf("Coerce(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("Coerce(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
	if _, err := Coerce("no digits here"); err == nil {
		t.Error("expected error for no digits")
	}
}

func TestCleanAndAccessors(t *testing.T) {
	if Clean(" v1.2.3 ") != "1.2.3" {
		t.Error("Clean")
	}
	if Clean("bogus") != "" {
		t.Error("Clean invalid")
	}
	if m, _ := Major("1.2.3"); m != 1 {
		t.Error("Major")
	}
	if m, _ := Minor("1.2.3"); m != 2 {
		t.Error("Minor")
	}
	if p, _ := Patch("1.2.3"); p != 3 {
		t.Error("Patch")
	}
	if pre, _ := Prerelease("1.2.3-rc.1"); pre != "rc.1" {
		t.Error("Prerelease")
	}
	if MustParse("1.2.3-a+b").Core().String() != "1.2.3" {
		t.Error("Core")
	}
}

func TestSatisfies(t *testing.T) {
	tests := []struct {
		v, r string
		want bool
	}{
		{"1.2.3", "^1.0.0", true},
		{"2.0.0", "^1.0.0", false},
		{"1.2.3", "~1.2.0", true},
		{"1.3.0", "~1.2.0", false},
		{"1.2.3", ">=1.0.0 <2.0.0", true},
		{"2.0.0", ">=1.0.0 <2.0.0", false},
		{"1.5.0", "1.x", true},
		{"2.5.0", "1.x", false},
		{"1.2.9", "1.2.x", true},
		{"1.3.0", "1.2.x", false},
		{"1.2.3", "*", true},
		{"1.2.3", "1.2.3 - 2.3.4", true},
		{"2.3.4", "1.2.3 - 2.3.4", true},
		{"2.3.5", "1.2.3 - 2.3.4", false},
		{"1.5.0", "1.x || 3.x", true},
		{"3.5.0", "1.x || 3.x", true},
		{"2.5.0", "1.x || 3.x", false},
		{"0.2.3", "^0.2.0", true},
		{"0.3.0", "^0.2.0", false},
		{"1.0.0", "=1.0.0", true},
		{"1.2.3", ">1.2.2", true},
		{"1.2.2", ">1.2.2", false},
		// prerelease semantics: only matches when comparator names same core pre
		{"1.2.3-alpha.1", "^1.2.3-alpha.0", true},
		{"1.2.4-alpha.1", "^1.2.3", false},
	}
	for _, tt := range tests {
		if got := Satisfies(tt.v, tt.r); got != tt.want {
			t.Errorf("Satisfies(%q,%q) = %v, want %v", tt.v, tt.r, got, tt.want)
		}
	}
}

func TestMaxMinSatisfying(t *testing.T) {
	versions := []string{"1.0.0", "1.2.0", "1.3.0", "2.0.0", "bogus"}
	if got, ok := MaxSatisfying(versions, "^1.0.0"); !ok || got != "1.3.0" {
		t.Errorf("MaxSatisfying = %q,%v", got, ok)
	}
	if got, ok := MinSatisfying(versions, "^1.0.0"); !ok || got != "1.0.0" {
		t.Errorf("MinSatisfying = %q,%v", got, ok)
	}
	if _, ok := MaxSatisfying(versions, "^9.0.0"); ok {
		t.Error("expected no match")
	}
}

func TestParseRangeString(t *testing.T) {
	r, err := ParseRange("^1.2.0")
	if err != nil {
		t.Fatal(err)
	}
	if r.String() != "^1.2.0" {
		t.Errorf("String = %q", r.String())
	}
	if !r.Test(MustParse("1.5.0")) {
		t.Error("Test")
	}
}

func BenchmarkCompare(b *testing.B) {
	x, y := MustParse("1.2.3-beta.2"), MustParse("1.2.3-beta.11")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = x.Compare(y)
	}
}

func BenchmarkSatisfies(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Satisfies("1.2.3", ">=1.0.0 <2.0.0")
	}
}
