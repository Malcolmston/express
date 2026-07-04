package jwtdecode

import "testing"

// Known token: header {"alg":"HS256","typ":"JWT"}, payload {"sub":"1234567890","name":"John Doe","iat":1516239022}
const knownToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestDecodePayload(t *testing.T) {
	claims, err := Decode(knownToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims["sub"] != "1234567890" {
		t.Errorf("sub = %v", claims["sub"])
	}
	if claims["name"] != "John Doe" {
		t.Errorf("name = %v", claims["name"])
	}
	if claims["iat"].(float64) != 1516239022 {
		t.Errorf("iat = %v", claims["iat"])
	}
}

func TestDecodeHeader(t *testing.T) {
	h, err := DecodeHeader(knownToken)
	if err != nil {
		t.Fatal(err)
	}
	if h["alg"] != "HS256" {
		t.Errorf("alg = %v", h["alg"])
	}
	if h["typ"] != "JWT" {
		t.Errorf("typ = %v", h["typ"])
	}
}

func TestMalformed(t *testing.T) {
	cases := []string{"", "onlyonepart", "two.parts", "a.b.c.d", "bad.@@@.sig"}
	for _, c := range cases {
		if _, err := Decode(c); err == nil {
			t.Errorf("expected error for %q", c)
		}
	}
}
