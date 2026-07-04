package otpauth

import (
	"strings"
	"testing"
)

func TestURLContains(t *testing.T) {
	c := Config{
		Type:    "totp",
		Issuer:  "ACME Co",
		Account: "alice@example.com",
		Secret:  "JBSWY3DPEHPK3PXP",
		Digits:  6,
		Period:  30,
	}
	u := URL(c)

	for _, sub := range []string{
		"otpauth://totp/ACME%20Co:alice@example.com",
		"secret=JBSWY3DPEHPK3PXP",
		"issuer=ACME%20Co",
	} {
		if !strings.Contains(u, sub) {
			t.Errorf("URL %q does not contain %q", u, sub)
		}
	}
}

func TestParseRoundTrip(t *testing.T) {
	c := Config{
		Type:    "totp",
		Issuer:  "ACME Co",
		Account: "alice@example.com",
		Secret:  "JBSWY3DPEHPK3PXP",
		Digits:  6,
		Period:  30,
	}
	u := URL(c)

	got, err := Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "totp" {
		t.Errorf("Type: got %q, want totp", got.Type)
	}
	if got.Issuer != "ACME Co" {
		t.Errorf("Issuer: got %q, want %q", got.Issuer, "ACME Co")
	}
	if got.Account != "alice@example.com" {
		t.Errorf("Account: got %q, want %q", got.Account, "alice@example.com")
	}
	if got.Secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("Secret: got %q, want %q", got.Secret, "JBSWY3DPEHPK3PXP")
	}
	if got.Digits != 6 {
		t.Errorf("Digits: got %d, want 6", got.Digits)
	}
	if got.Period != 30 {
		t.Errorf("Period: got %d, want 30", got.Period)
	}
}

func TestHOTPURL(t *testing.T) {
	c := Config{
		Type:    "hotp",
		Issuer:  "ACME",
		Account: "bob",
		Secret:  "JBSWY3DPEHPK3PXP",
		Counter: 5,
	}
	u := URL(c)
	if !strings.Contains(u, "counter=5") {
		t.Errorf("hotp URL %q should contain counter=5", u)
	}
	if strings.Contains(u, "period=") {
		t.Errorf("hotp URL %q should not contain period", u)
	}
}
