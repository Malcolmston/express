package encodeurl

import "testing"

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain path unchanged", "/foo/bar", "/foo/bar"},
		{"space encoded", "/foo bar", "/foo%20bar"},
		{"already encoded space stays", "/foo%20bar", "/foo%20bar"},
		{"valid two-hex escape preserved", "foo%bar", "foo%bar"},
		{"lone percent (non-hex follows)", "foo%gbar", "foo%25gbar"},
		{"lone percent at end", "foo%", "foo%25"},
		{"percent single hex non-hex", "%2G", "%252G"},
		{"percent one hex at end kept", "%2", "%2"},
		{"query preserved", "/search?q=1&r=2", "/search?q=1&r=2"},
		{"unicode encoded", "/€", "/%E2%82%AC"},
		{"accent encoded", "/café", "/caf%C3%A9"},
		{"double quote encoded", "/a\"b", "/a%22b"},
		{"backslash encoded", "/a\\b", "/a%5Cb"},
		{"hash preserved", "/a#b", "/a#b"},
		{"dollar preserved", "/a$b", "/a$b"},
		{"reserved preserved", "/a[b]c", "/a[b]c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Encode(tt.in); got != tt.want {
				t.Fatalf("Encode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEncodeDoesNotDoubleEncode(t *testing.T) {
	once := Encode("/a b%20c")
	twice := Encode(once)
	if once != twice {
		t.Fatalf("re-encoding changed value: %q -> %q", once, twice)
	}
	if once != "/a%20b%20c" {
		t.Fatalf("Encode = %q, want %q", once, "/a%20b%20c")
	}
}

func TestEncodeInvalidUTF8(t *testing.T) {
	got := Encode("/\xff")
	want := "/%EF%BF%BD" // U+FFFD
	if got != want {
		t.Fatalf("Encode(invalid) = %q, want %q", got, want)
	}
}
