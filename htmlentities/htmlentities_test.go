package htmlentities

import "testing"

func TestEncodeSpecialChars(t *testing.T) {
	got := Encode(`<a href="x">Tom & Jerry's</a>`)
	want := `&lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&apos;s&lt;/a&gt;`
	if got != want {
		t.Fatalf("Encode = %q; want %q", got, want)
	}
}

func TestEncodeLeavesAsciiAlone(t *testing.T) {
	if got := Encode("hello world 123"); got != "hello world 123" {
		t.Fatalf("Encode = %q", got)
	}
}

func TestEncodeSpecialCharsKeepsNonAscii(t *testing.T) {
	if got := Encode("café ©"); got != "café ©" {
		t.Fatalf("specialChars mode should not touch non-ascii; got %q", got)
	}
}

func TestEncodeNonAscii(t *testing.T) {
	got := Encode("café ©", EncodeOptions{Mode: "nonAscii"})
	want := "caf&#233; &#169;"
	if got != want {
		t.Fatalf("Encode nonAscii = %q; want %q", got, want)
	}
}

func TestEncodeNonAsciiAlsoEncodesSpecials(t *testing.T) {
	got := Encode("<é>", EncodeOptions{Mode: "nonAscii"})
	want := "&lt;&#233;&gt;"
	if got != want {
		t.Fatalf("Encode = %q; want %q", got, want)
	}
}

func TestDecodeNamed(t *testing.T) {
	got := Decode(`&lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&apos;s`)
	want := `<a href="x">Tom & Jerry's`
	if got != want {
		t.Fatalf("Decode = %q; want %q", got, want)
	}
}

func TestDecodeExtendedNamed(t *testing.T) {
	got := Decode("&copy; &nbsp;&mdash;")
	want := "© \u00a0—"
	if got != want {
		t.Fatalf("Decode = %q; want %q", got, want)
	}
}

func TestDecodeDecimal(t *testing.T) {
	if got := Decode("&#233;"); got != "é" {
		t.Fatalf("Decode decimal = %q", got)
	}
}

func TestDecodeHex(t *testing.T) {
	if got := Decode("&#xe9;"); got != "é" {
		t.Fatalf("Decode hex lowercase = %q", got)
	}
	if got := Decode("&#XE9;"); got != "é" {
		t.Fatalf("Decode hex uppercase = %q", got)
	}
}

func TestDecodeUnknownLeftAlone(t *testing.T) {
	if got := Decode("&unknownentity;"); got != "&unknownentity;" {
		t.Fatalf("unknown entity should be untouched; got %q", got)
	}
	if got := Decode("just & text"); got != "just & text" {
		t.Fatalf("bare ampersand should be untouched; got %q", got)
	}
	if got := Decode("a & b & c"); got != "a & b & c" {
		t.Fatalf("bare ampersands = %q", got)
	}
}

func TestDecodeTrailingAmp(t *testing.T) {
	if got := Decode("trailing&"); got != "trailing&" {
		t.Fatalf("trailing amp = %q", got)
	}
}

func TestDecodeInvalidNumeric(t *testing.T) {
	if got := Decode("&#zz;"); got != "&#zz;" {
		t.Fatalf("invalid numeric = %q", got)
	}
}

func TestRoundTripSpecialChars(t *testing.T) {
	inputs := []string{
		`Tom & Jerry`,
		`<div class="a">'quote'</div>`,
		`a & b < c > d " e ' f`,
	}
	for _, in := range inputs {
		if got := Decode(Encode(in)); got != in {
			t.Fatalf("round trip failed for %q; got %q", in, got)
		}
	}
}

func TestDecodeNoAmpersandFastPath(t *testing.T) {
	if got := Decode("plain text"); got != "plain text" {
		t.Fatalf("plain = %q", got)
	}
}
