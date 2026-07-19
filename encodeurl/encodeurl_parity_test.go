package encodeurl

import "testing"

// Upstream-parity tests for the pillarjs/encodeurl npm package.
//
// Every vector below is taken verbatim from the original library's test suite
// and reference implementation:
//
//	https://raw.githubusercontent.com/pillarjs/encodeurl/master/test/test.js
//	https://raw.githubusercontent.com/pillarjs/encodeurl/master/index.js
//
// The upstream ENCODE_CHARS_REGEXP keeps the character class
// [\x21\x23-\x3B\x3D\x3F-\x5F\x61-\x7A\x7C\x7E] unencoded (note that '\',
// '^', and '|' are preserved), and encodes every other character via
// JavaScript's encodeURI over the matched substrings. The '%' character is
// preserved when it begins a valid "%XX" escape and encoded to "%25"
// otherwise.
//
// Two categories need translation from upstream's JavaScript string model
// (UTF-16 code units) to Go's (UTF-8 bytes):
//
//   - "above ASCII" vectors: upstream literals like '\x80' are the code point
//     U+0080, not a raw 0x80 byte, so the Go analog is the same code point
//     (written here with \u escapes), which encodes to the UTF-8 sequence
//     %C2%80 exactly as upstream expects.
//   - unpaired-surrogate vectors: a Go source string cannot hold an unpaired
//     surrogate, so the analog is invalid UTF-8, which Go decodes to U+FFFD
//     before encoding, matching upstream's surrogate-replacement behavior.

func TestParityUpstream(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// when url contains only allowed characters
		{"keep url the same", "http://localhost/foo/bar.html?fizz=buzz#readme", "http://localhost/foo/bar.html?fizz=buzz#readme"},
		{"ipv6 notation untouched", "http://[::1]:8080/foo/bar", "http://[::1]:8080/foo/bar"},
		{"backslashes untouched", "http:\\\\localhost\\foo\\bar.html", "http:\\\\localhost\\foo\\bar.html"},

		// when url contains invalid raw characters
		{"encode LF", "http://localhost/\nsnow.html", "http://localhost/%0Asnow.html"},
		{"encode FF", "http://localhost/\fsnow.html", "http://localhost/%0Csnow.html"},
		{"encode CR", "http://localhost/\rsnow.html", "http://localhost/%0Dsnow.html"},
		{"encode SP", "http://localhost/ snow.html", "http://localhost/%20snow.html"},
		{"encode NULL", "http://localhost/\x00snow.html", "http://localhost/%00snow.html"},

		// full ASCII set, one row per 16 code points (from upstream)
		{"ascii 0x00-0x0f", "/\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f", "/%00%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F"},
		{"ascii 0x10-0x1f", "/\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f", "/%10%11%12%13%14%15%16%17%18%19%1A%1B%1C%1D%1E%1F"},
		{"ascii 0x20-0x2f", "/\x20\x21\x22\x23\x24\x25\x26\x27\x28\x29\x2a\x2b\x2c\x2d\x2e\x2f", "/%20!%22#$%25&'()*+,-./"},
		{"ascii 0x30-0x3f", "/\x30\x31\x32\x33\x34\x35\x36\x37\x38\x39\x3a\x3b\x3c\x3d\x3e\x3f", "/0123456789:;%3C=%3E?"},
		{"ascii 0x40-0x4f", "/\x40\x41\x42\x43\x44\x45\x46\x47\x48\x49\x4a\x4b\x4c\x4d\x4e\x4f", "/@ABCDEFGHIJKLMNO"},
		{"ascii 0x50-0x5f", "/\x50\x51\x52\x53\x54\x55\x56\x57\x58\x59\x5a\x5b\x5c\x5d\x5e\x5f", "/PQRSTUVWXYZ[\\]^_"},
		{"ascii 0x60-0x6f", "/\x60\x61\x62\x63\x64\x65\x66\x67\x68\x69\x6a\x6b\x6c\x6d\x6e\x6f", "/%60abcdefghijklmno"},
		{"ascii 0x70-0x7f", "/\x70\x71\x72\x73\x74\x75\x76\x77\x78\x79\x7a\x7b\x7c\x7d\x7e\x7f", "/pqrstuvwxyz%7B|%7D~%7F"},

		// code points above ASCII encode as UTF-8 sequences (see file comment)
		{"above ascii U+0080-U+008F", "/", "/%C2%80%C2%81%C2%82%C2%83%C2%84%C2%85%C2%86%C2%87%C2%88%C2%89%C2%8A%C2%8B%C2%8C%C2%8D%C2%8E%C2%8F"},
		{"above ascii U+00A0-U+00AF", "/ ¡¢£¤¥¦§¨©ª«¬­®¯", "/%C2%A0%C2%A1%C2%A2%C2%A3%C2%A4%C2%A5%C2%A6%C2%A7%C2%A8%C2%A9%C2%AA%C2%AB%C2%AC%C2%AD%C2%AE%C2%AF"},
		{"above ascii U+00C0-U+00CF", "/ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏ", "/%C3%80%C3%81%C3%82%C3%83%C3%84%C3%85%C3%86%C3%87%C3%88%C3%89%C3%8A%C3%8B%C3%8C%C3%8D%C3%8E%C3%8F"},
		{"above ascii U+00F0-U+00FF", "/ðñòóôõö÷øùúûüýþÿ", "/%C3%B0%C3%B1%C3%B2%C3%B3%C3%B4%C3%B5%C3%B6%C3%B7%C3%B8%C3%B9%C3%BA%C3%BB%C3%BC%C3%BD%C3%BE%C3%BF"},

		// when url contains percent-encoded sequences
		{"keep percent for valid escape", "http://localhost/%20snow.html", "http://localhost/%20snow.html"},
		{"keep percent regardless of utf8 validity", "http://localhost/%F0snow.html", "http://localhost/%F0snow.html"},
		{"encode percent when not a valid sequence", "http://localhost/%foo%bar%zap%", "http://localhost/%25foo%bar%25zap%25"},

		// when url contains raw surrogate pairs (Go equivalent: astral rune)
		{"encode valid astral code point as utf8", "http://localhost/\U0001F47B snow.html", "http://localhost/%F0%9F%91%BB%20snow.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Encode(tt.in); got != tt.want {
				t.Fatalf("Encode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestParityUnpairedSurrogate mirrors upstream's "unpaired surrogate ->
// replacement character" behavior. Go source strings cannot hold unpaired
// surrogates, so the Go analog is invalid UTF-8, which decodes to U+FFFD and
// then encodes to its UTF-8 form %EF%BF%BD, exactly as upstream emits for a
// replaced surrogate.
func TestParityUnpairedSurrogate(t *testing.T) {
	// A lone 0xFF byte is invalid UTF-8 -> U+FFFD -> %EF%BF%BD.
	if got, want := Encode("http://localhost/\xff"), "http://localhost/%EF%BF%BD"; got != want {
		t.Fatalf("Encode(invalid at middle) = %q, want %q", got, want)
	}
	if got, want := Encode("\xfffoo"), "%EF%BF%BDfoo"; got != want {
		t.Fatalf("Encode(invalid at start) = %q, want %q", got, want)
	}
}
