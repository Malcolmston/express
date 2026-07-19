package dotenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	in := `# a comment
FOO=bar
BAZ = qux
EMPTY=
`
	m, err := Parse(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if m["FOO"] != "bar" {
		t.Errorf("FOO = %q", m["FOO"])
	}
	if m["BAZ"] != "qux" {
		t.Errorf("BAZ = %q", m["BAZ"])
	}
	if v, ok := m["EMPTY"]; !ok || v != "" {
		t.Errorf("EMPTY = %q ok=%v", v, ok)
	}
}

func TestParseExportPrefix(t *testing.T) {
	m, _ := Bytes([]byte("export TOKEN=abc123\n"))
	if m["TOKEN"] != "abc123" {
		t.Errorf("TOKEN = %q", m["TOKEN"])
	}
}

func TestParseQuotes(t *testing.T) {
	in := `SINGLE='no $expand \n literal'
DOUBLE="line1\nline2\ttab"
SPACES="  padded  "
HASHIN="a#b"
`
	m, _ := Parse(strings.NewReader(in))
	if m["SINGLE"] != `no $expand \n literal` {
		t.Errorf("SINGLE = %q", m["SINGLE"])
	}
	if m["DOUBLE"] != "line1\nline2\ttab" {
		t.Errorf("DOUBLE = %q", m["DOUBLE"])
	}
	if m["SPACES"] != "  padded  " {
		t.Errorf("SPACES = %q", m["SPACES"])
	}
	if m["HASHIN"] != "a#b" {
		t.Errorf("HASHIN = %q", m["HASHIN"])
	}
}

func TestParseTrailingComment(t *testing.T) {
	m, _ := Parse(strings.NewReader("KEY=value # trailing\n"))
	if m["KEY"] != "value" {
		t.Errorf("KEY = %q", m["KEY"])
	}
}

func TestParseBlankAndInvalid(t *testing.T) {
	m, _ := Parse(strings.NewReader("\n\n   \nNOEQUALS\n=noKey\nGOOD=1\n"))
	// Upstream dotenv's key/value separator is `\s*=`, whose `\s*` spans the
	// newline between "NOEQUALS" and "=noKey", so those two physical lines parse
	// as the single assignment NOEQUALS=noKey (verified against motdotla/dotenv
	// lib/main.js). The bare "=noKey" therefore is not a standalone entry.
	if len(m) != 2 || m["GOOD"] != "1" || m["NOEQUALS"] != "noKey" {
		t.Errorf("unexpected map: %#v", m)
	}
}

func TestLoadNoOverride(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	os.WriteFile(p, []byte("DOTENV_TEST_A=fromfile\nDOTENV_TEST_B=alsofile\n"), 0o644)

	os.Setenv("DOTENV_TEST_A", "preset")
	defer os.Unsetenv("DOTENV_TEST_A")
	os.Unsetenv("DOTENV_TEST_B")
	defer os.Unsetenv("DOTENV_TEST_B")

	if err := Load(p); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DOTENV_TEST_A"); got != "preset" {
		t.Errorf("A should not be overridden, got %q", got)
	}
	if got := os.Getenv("DOTENV_TEST_B"); got != "alsofile" {
		t.Errorf("B = %q", got)
	}
}

func TestLoadOverride(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	os.WriteFile(p, []byte("DOTENV_TEST_C=fromfile\n"), 0o644)
	os.Setenv("DOTENV_TEST_C", "preset")
	defer os.Unsetenv("DOTENV_TEST_C")

	if err := LoadOverride(p); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("DOTENV_TEST_C"); got != "fromfile" {
		t.Errorf("C should be overridden, got %q", got)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if err := Load(filepath.Join(t.TempDir(), "nope.env")); err == nil {
		t.Error("expected error for missing file")
	}
}
