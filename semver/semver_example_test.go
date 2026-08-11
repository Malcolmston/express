package semver_test

import (
	"fmt"

	"github.com/malcolmston/express/semver"
)

// ExampleSatisfies shows the range check that drives dependency resolution:
// does a concrete version fall inside a published constraint?
func ExampleSatisfies() {
	fmt.Println(semver.Satisfies("1.2.3", "^1.0.0"))
	fmt.Println(semver.Satisfies("2.0.0", "^1.0.0"))
	fmt.Println(semver.Satisfies("1.2.7", "~1.2.0"))
	fmt.Println(semver.Satisfies("1.3.0", "~1.2.0"))
	// Output:
	// true
	// false
	// true
	// false
}

// ExampleCompare shows the precedence rules from clause 11 of semver.org: a
// prerelease sorts below its release, numeric prerelease identifiers compare
// numerically, and build metadata is ignored entirely.
func ExampleCompare() {
	fmt.Println(semver.Compare("1.0.0-alpha", "1.0.0"))
	fmt.Println(semver.Compare("1.0.0-beta.2", "1.0.0-beta.11"))
	fmt.Println(semver.Compare("1.0.0+build.1", "1.0.0+build.999"))
	// Output:
	// -1
	// -1
	// 0
}

// ExampleSort shows ordering a list of versions in place.
func ExampleSort() {
	versions := []string{"1.10.0", "1.2.0", "1.0.0-rc.1", "1.0.0", "1.2.0-beta"}
	semver.Sort(versions)
	fmt.Println(versions)
	// Output: [1.0.0-rc.1 1.0.0 1.2.0-beta 1.2.0 1.10.0]
}

// ExampleMaxSatisfying shows picking the newest version a constraint allows,
// which is what a package manager does when it resolves a dependency.
func ExampleMaxSatisfying() {
	versions := []string{"1.0.0", "1.2.0", "1.9.9", "2.0.0"}
	best, ok := semver.MaxSatisfying(versions, "^1.0.0")
	fmt.Println(best, ok)
	// Output: 1.9.9 true
}

// ExampleInc shows bumping a version by release type.
func ExampleInc() {
	for _, release := range []string{"major", "minor", "patch", "prerelease"} {
		next, err := semver.Inc("1.2.3", release)
		fmt.Println(release, next, err)
	}
	// Output:
	// major 2.0.0 <nil>
	// minor 1.3.0 <nil>
	// patch 1.2.4 <nil>
	// prerelease 1.2.4-0 <nil>
}

// ExampleCoerce shows salvaging a semver-ish string that is not strictly valid.
func ExampleCoerce() {
	for _, s := range []string{"v1.2", "1", "release-3.4.5-rc"} {
		v, err := semver.Coerce(s)
		fmt.Printf("%-16q -> %v %v\n", s, v, err)
	}
	// Output:
	// "v1.2"           -> 1.2.0 <nil>
	// "1"              -> 1.0.0 <nil>
	// "release-3.4.5-rc" -> 3.4.5 <nil>
}
