package escaperegexp

import (
	"regexp"
	"testing"
)

func TestEscapeRegExp(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello", "hello"},
		{"dot", "a.b", `a\.b`},
		{"all metachars", `|\{}()[]^$+*?.`, `\|\\\{\}\(\)\[\]\^\$\+\*\?\.`},
		{"dash", "a-b", `a\x2db`},
		{"how much wood", "How much wood would a woodchuck chuck?", `How much wood would a woodchuck chuck\?`},
		{"parens", "foo (bar)", `foo \(bar\)`},
		{"empty", "", ""},
		{"unicode preserved", "héllo.wörld", `héllo\.wörld`},
		{"multiple dashes", "a-b-c", `a\x2db\x2dc`},
		{"price", "$5.00", `\$5\.00`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeRegExp(tt.in)
			if got != tt.want {
				t.Fatalf("EscapeRegExp(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestEscapeMatchesLiterally verifies that the escaped string, when compiled as
// a regexp, matches the original input literally.
func TestEscapeMatchesLiterally(t *testing.T) {
	inputs := []string{
		"hello",
		"a.b",
		`|\{}()[]^$+*?.`,
		"How much wood would a woodchuck chuck?",
		"foo (bar) [baz] {qux}",
		"$5.00 + $10.00 = $15.00",
		"a-b-c-d",
		"héllo.wörld",
		"1*2*3?4",
	}
	for _, in := range inputs {
		escaped := EscapeRegExp(in)
		re, err := regexp.Compile(escaped)
		if err != nil {
			t.Fatalf("regexp.Compile(%q) error: %v", escaped, err)
		}
		loc := re.FindString(in)
		if loc != in {
			t.Fatalf("escaped pattern %q did not match input %q literally (got %q)", escaped, in, loc)
		}
		// Also ensure anchored full match.
		full := regexp.MustCompile("^" + escaped + "$")
		if !full.MatchString(in) {
			t.Fatalf("anchored escaped pattern did not full-match input %q", in)
		}
	}
}
