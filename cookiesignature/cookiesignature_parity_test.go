package cookiesignature

import "testing"

// Parity tests transcribed from the upstream npm library tj/node-cookie-signature.
// Source (fetched 2026-07-19, branch "master"):
//   https://raw.githubusercontent.com/tj/node-cookie-signature/master/test/index.js
//   https://raw.githubusercontent.com/tj/node-cookie-signature/master/index.js
//
// Upstream uses HMAC-SHA256, standard base64 (with '+' and '/'), and strips
// trailing '=' padding: val + '.' + hmac.digest('base64').replace(/=+$/, '').
// unsign re-signs the tentative value and compares the whole string in
// constant time, returning the value or false. The vectors below are the real
// values and expected outputs from upstream's test/index.js.

// binaryKey is the Buffer.from("A0ABBC0C", 'hex') secret used upstream to
// exercise non-string (raw byte) keys. As a Go string it is those four bytes.
var binaryKey = string([]byte{0xA0, 0xAB, 0xBC, 0x0C})

func TestParitySign(t *testing.T) {
	// cookie.sign('hello', 'tobiiscool')
	//   => 'hello.DGDUkGlIkCzPz+C0B064FNgHdEjox7ch8tOBGslZ5QI'
	if got, want := Sign("hello", "tobiiscool"), "hello.DGDUkGlIkCzPz+C0B064FNgHdEjox7ch8tOBGslZ5QI"; got != want {
		t.Errorf("Sign(\"hello\", \"tobiiscool\") = %q, want %q", got, want)
	}
	// cookie.sign('hello', 'luna') should NOT equal the tobiiscool signature.
	if Sign("hello", "luna") == "hello.DGDUkGlIkCzPz+C0B064FNgHdEjox7ch8tOBGslZ5QI" {
		t.Error("Sign(\"hello\", \"luna\") must differ from the tobiiscool signature")
	}
}

func TestParitySignBinarySecret(t *testing.T) {
	// var key = Buffer.from("A0ABBC0C", 'hex');
	// cookie.sign('hello', key)
	//   => 'hello.hIvljrKw5oOZtHHSq5u+MlL27cgnPKX77y7F+x5r1to'
	if got, want := Sign("hello", binaryKey), "hello.hIvljrKw5oOZtHHSq5u+MlL27cgnPKX77y7F+x5r1to"; got != want {
		t.Errorf("Sign(\"hello\", binaryKey) = %q, want %q", got, want)
	}
}

func TestParityUnsign(t *testing.T) {
	// var val = cookie.sign('hello', 'tobiiscool');
	// cookie.unsign(val, 'tobiiscool').should.equal('hello');
	val := Sign("hello", "tobiiscool")
	if got, ok := Unsign(val, "tobiiscool"); !ok || got != "hello" {
		t.Errorf("Unsign(val, \"tobiiscool\") = (%q, %v), want (\"hello\", true)", got, ok)
	}
	// cookie.unsign(val, 'luna').should.be.false();
	if got, ok := Unsign(val, "luna"); ok || got != "" {
		t.Errorf("Unsign(val, \"luna\") = (%q, %v), want (\"\", false)", got, ok)
	}
}

func TestParityUnsignMalformed(t *testing.T) {
	pwd := "actual sekrit password"
	// cookie.unsign('fake unsigned data', pwd).should.be.false();
	if got, ok := Unsign("fake unsigned data", pwd); ok || got != "" {
		t.Errorf("Unsign(\"fake unsigned data\", pwd) = (%q, %v), want (\"\", false)", got, ok)
	}
	val := Sign("real data", pwd)
	// Each of these tampered forms must fail verification.
	cases := []string{
		"garbage" + val,
		"garbage." + val,
		val + ".garbage",
		val + "garbage",
	}
	for _, c := range cases {
		if got, ok := Unsign(c, pwd); ok || got != "" {
			t.Errorf("Unsign(%q, pwd) = (%q, %v), want (\"\", false)", c, got, ok)
		}
	}
}

func TestParityUnsignBinarySecret(t *testing.T) {
	// var key = Uint8Array.from([0xA0, 0xAB, 0xBC, 0x0C]);
	// cookie.unsign('hello.hIvljrKw5oOZtHHSq5u+MlL27cgnPKX77y7F+x5r1to', key)
	//   .should.equal('hello');
	signed := "hello.hIvljrKw5oOZtHHSq5u+MlL27cgnPKX77y7F+x5r1to"
	if got, ok := Unsign(signed, binaryKey); !ok || got != "hello" {
		t.Errorf("Unsign(%q, binaryKey) = (%q, %v), want (\"hello\", true)", signed, got, ok)
	}
}
