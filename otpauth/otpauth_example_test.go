package otpauth_test

import (
	"fmt"

	"github.com/malcolmston/express/otpauth"
)

// ExampleURL builds a TOTP otpauth:// URI from a fixed configuration. The
// Issuer and Account combine into the "Issuer:Account" label, and the secret
// plus optional parameters become the query string. Spaces in the issuer are
// percent-encoded as %20 to match the otpauth specification and authenticator
// apps. Parameters are emitted in a fixed order, and because this is a TOTP
// config a period is written but no counter. The resulting URI is what you
// would typically encode into a QR code during two-factor enrollment.
func ExampleURL() {
	c := otpauth.Config{
		Type:    "totp",
		Issuer:  "ACME Co",
		Account: "alice@example.com",
		Secret:  "JBSWY3DPEHPK3PXP",
		Digits:  6,
		Period:  30,
	}
	fmt.Println(otpauth.URL(c))
	// Output: otpauth://totp/ACME%20Co:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=ACME%20Co&digits=6&period=30
}

// ExampleURL_hotp builds a counter-based HOTP otpauth:// URI. Selecting the
// "hotp" type causes the initial counter value to be written into the query
// string and suppresses the period parameter, which only applies to
// time-based TOTP. All other fields behave the same as for TOTP. This is the
// form used by authenticator apps that advance a counter on each generated
// code rather than tracking wall-clock time.
func ExampleURL_hotp() {
	c := otpauth.Config{
		Type:    "hotp",
		Issuer:  "ACME",
		Account: "bob",
		Secret:  "JBSWY3DPEHPK3PXP",
		Digits:  6,
		Counter: 5,
	}
	fmt.Println(otpauth.URL(c))
	// Output: otpauth://hotp/ACME:bob?secret=JBSWY3DPEHPK3PXP&issuer=ACME&digits=6&counter=5
}

// ExampleParse reads an otpauth:// URI back into a Config, the inverse of URL.
// The label is split into issuer and account, and each recognized query
// parameter is decoded into its typed field. Percent-encoded values such as
// the %20 in the issuer are decoded automatically. Parse tolerates any
// parameter ordering and leaves absent optional fields at their zero values.
// A URI with a scheme other than otpauth, or with a non-integer digits,
// period or counter, would instead return an error.
func ExampleParse() {
	c, err := otpauth.Parse("otpauth://totp/ACME%20Co:alice@example.com?secret=JBSWY3DPEHPK3PXP&issuer=ACME%20Co&digits=6&period=30")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("type=%s issuer=%q account=%q secret=%s digits=%d period=%d\n",
		c.Type, c.Issuer, c.Account, c.Secret, c.Digits, c.Period)
	// Output: type=totp issuer="ACME Co" account="alice@example.com" secret=JBSWY3DPEHPK3PXP digits=6 period=30
}
