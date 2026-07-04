package contentrange

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		unit             string
		start, end, size int64
		want             string
	}{
		{"bytes", 0, 499, 1234, "bytes 0-499/1234"},
		{"bytes", -1, 0, 1234, "bytes */1234"},
		{"bytes", 0, 499, -1, "bytes 0-499/*"},
		{"bytes", -1, -1, -1, "bytes */*"},
		{"", 0, 499, 1234, "bytes 0-499/1234"},
		{"items", 5, 9, 20, "items 5-9/20"},
	}
	for _, tt := range tests {
		if got := Format(tt.unit, tt.start, tt.end, tt.size); got != tt.want {
			t.Fatalf("Format(%q,%d,%d,%d) = %q, want %q",
				tt.unit, tt.start, tt.end, tt.size, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    ContentRange
		wantErr bool
	}{
		{"bytes 0-499/1234", ContentRange{Unit: "bytes", Start: 0, End: 499, Size: 1234, HasRange: true, HasSize: true}, false},
		{"bytes */1234", ContentRange{Unit: "bytes", Start: -1, End: -1, Size: 1234, HasRange: false, HasSize: true}, false},
		{"bytes 0-499/*", ContentRange{Unit: "bytes", Start: 0, End: 499, Size: -1, HasRange: true, HasSize: false}, false},
		{"bytes */*", ContentRange{Unit: "bytes", Start: -1, End: -1, Size: -1, HasRange: false, HasSize: false}, false},
		{"missingslash", ContentRange{}, true},
		{"noUnit", ContentRange{}, true},
		{"bytes bad/1234", ContentRange{}, true},
		{"bytes 0-499/xx", ContentRange{}, true},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("Parse(%q) expected error, got %+v", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("Parse(%q) unexpected error: %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("Parse(%q) = %+v, want %+v", tt.in, got, tt.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{
		"bytes 0-499/1234",
		"bytes */1234",
		"bytes 0-499/*",
		"bytes */*",
	}
	for _, in := range inputs {
		cr, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", in, err)
		}
		if got := cr.String(); got != in {
			t.Fatalf("round-trip %q -> %q", in, got)
		}
	}
}
