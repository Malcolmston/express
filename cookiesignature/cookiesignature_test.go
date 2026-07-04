package cookiesignature

import (
	"strings"
	"testing"
)

func TestSignFormat(t *testing.T) {
	got := Sign("hello", "tobiiscool")
	if !strings.HasPrefix(got, "hello.") {
		t.Fatalf("expected prefix hello., got %q", got)
	}
	if strings.HasSuffix(got, "=") {
		t.Fatalf("signature should not contain padding: %q", got)
	}
}

func TestSignDeterministic(t *testing.T) {
	a := Sign("hello", "secret")
	b := Sign("hello", "secret")
	if a != b {
		t.Fatalf("sign not deterministic: %q != %q", a, b)
	}
}

func TestSignDifferentSecret(t *testing.T) {
	if Sign("hello", "s1") == Sign("hello", "s2") {
		t.Fatal("different secrets produced same signature")
	}
}

func TestUnsignValid(t *testing.T) {
	signed := Sign("hello", "tobiiscool")
	v, ok := Unsign(signed, "tobiiscool")
	if !ok || v != "hello" {
		t.Fatalf("expected hello,true got %q,%v", v, ok)
	}
}

func TestUnsignWrongSecret(t *testing.T) {
	signed := Sign("hello", "tobiiscool")
	v, ok := Unsign(signed, "wrong")
	if ok || v != "" {
		t.Fatalf("expected \"\",false got %q,%v", v, ok)
	}
}

func TestUnsignTampered(t *testing.T) {
	signed := Sign("hello", "secret")
	v, ok := Unsign("world"+signed[strings.LastIndex(signed, "."):], "secret")
	if ok || v != "" {
		t.Fatalf("tampered value should fail, got %q,%v", v, ok)
	}
}

func TestUnsignNoSeparator(t *testing.T) {
	v, ok := Unsign("nodot", "secret")
	if ok || v != "" {
		t.Fatalf("expected failure for no separator, got %q,%v", v, ok)
	}
}

func TestUnsignValueWithDot(t *testing.T) {
	// Values may contain dots; only the last dot separates the signature.
	signed := Sign("a.b.c", "secret")
	v, ok := Unsign(signed, "secret")
	if !ok || v != "a.b.c" {
		t.Fatalf("expected a.b.c,true got %q,%v", v, ok)
	}
}

func TestUnsignEmptyValue(t *testing.T) {
	signed := Sign("", "secret")
	v, ok := Unsign(signed, "secret")
	if !ok || v != "" {
		t.Fatalf("expected \"\",true got %q,%v", v, ok)
	}
}
