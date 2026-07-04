package express

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// exempt lists production package directories (paths relative to the repo root)
// that are permitted to ship without any _test.go file, each mapped to an honest
// reason. Keep this as small as possible: add an entry ONLY for a package that
// legitimately has nothing meaningful to test (for example a pure
// type-declaration package or generated code). New feature/util packages must
// ship tests instead of being exempted here.
var exempt = map[string]string{}

// skipTrees are top-level directory names whose subtrees are excluded from the
// per-package test coverage guard: version control metadata, CI config, docs,
// runnable examples, cross-runtime interop shims, and test fixtures.
var skipTrees = map[string]bool{
	".git":     true,
	".github":  true,
	"docs":     true,
	"examples": true,
	"interop":  true,
	"testdata": true,
}

// TestEveryPackageShipsTests enforces that every Go package under the module
// that contains production code (at least one non-_test.go .go file) also ships
// at least one _test.go file. It walks the repo tree from the working directory
// (the repo root at test time), skips the configured trees, any vendored
// package, and any directory that contains only _test.go files, then fails once
// with a sorted list of every offending package. Legitimately test-free
// packages must be recorded in the exempt map with a reason.
func TestEveryPackageShipsTests(t *testing.T) {
	// prodDirs holds package directories (relative paths) that contain
	// production code; hasTest records whether each also ships a test file.
	prodDirs := map[string]bool{}
	hasTest := map[string]bool{}

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel := filepath.ToSlash(filepath.Clean(path))

		if d.IsDir() {
			if rel == "." {
				return nil
			}
			top := strings.Split(rel, "/")[0]
			if skipTrees[top] {
				return fs.SkipDir
			}
			if d.Name() == "vendor" {
				return fs.SkipDir
			}
			return nil
		}

		name := d.Name()
		if !strings.HasSuffix(name, ".go") {
			return nil
		}

		dir := filepath.ToSlash(filepath.Dir(rel))
		// Guard against any vendored file the tree walk did not prune.
		if strings.Contains("/"+dir+"/", "/vendor/") {
			return nil
		}

		if strings.HasSuffix(name, "_test.go") {
			hasTest[dir] = true
		} else {
			prodDirs[dir] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking repo tree: %v", err)
	}

	var offenders []string
	checked := 0
	for dir := range prodDirs {
		checked++
		if hasTest[dir] {
			continue
		}
		if reason, ok := exempt[dir]; ok {
			t.Logf("exempt package %q shipped without tests: %s", dir, reason)
			continue
		}
		offenders = append(offenders, dir)
	}

	// Fail fast on stale exemptions so the map stays honest and minimal.
	for dir, reason := range exempt {
		if !prodDirs[dir] {
			t.Errorf("exempt entry %q (%s) is not a production package directory; remove it", dir, reason)
			continue
		}
		if hasTest[dir] {
			t.Errorf("exempt entry %q (%s) now ships tests; remove it from the exempt map", dir, reason)
		}
	}

	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Errorf("%d production package(s) ship no _test.go file (every shipped module/middleware/util package must ship tests); "+
			"add tests or, only if genuinely test-free, record them in the exempt map with a reason:\n  %s",
			len(offenders), strings.Join(offenders, "\n  "))
	}

	t.Logf("checked %d production package(s); %d exemption(s)", checked, len(exempt))

	// Sanity check: the guard must actually be exercising the tree.
	if checked == 0 {
		wd, _ := os.Getwd()
		t.Fatalf("no production packages found under %q; the coverage guard is not walking the module", wd)
	}
}
