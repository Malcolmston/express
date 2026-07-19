package otpauth

// Upstream parity tests for the otpauth:// URI layer.
//
// Every input -> expected-output vector below is copied verbatim from the
// canonical test fixtures of the original npm library "hectorm/otpauth"
// (the toString.output[0] canonical form for each case), fetched from:
//
//	https://raw.githubusercontent.com/hectorm/otpauth/master/test/test.mjs
//
// The relevant upstream semantics (HOTP/TOTP toString and URI.parse/stringify)
// live in:
//
//	https://raw.githubusercontent.com/hectorm/otpauth/master/src/uri.js
//	https://raw.githubusercontent.com/hectorm/otpauth/master/src/hotp.js
//	https://raw.githubusercontent.com/hectorm/otpauth/master/src/totp.js
//
// Upstream toString() emits, in order: the "ISSUER:LABEL" (or "LABEL") path,
// then the query parameters issuer (when the issuer is non-empty), secret,
// algorithm, digits, and finally counter (HOTP) or period (TOTP). Spaces are
// percent-encoded as %20. This Go port maps upstream's "label" onto
// Config.Account and upstream's "issuer" onto Config.Issuer.

import (
	"testing"
)

// Secrets (base32, no spaces) taken from the corresponding upstream fixtures.
const (
	secret00 = "6OWYXIW7QEYH34MFXCCXPZUBQDTIXBSX5GPKX4MSU2W6NHFNY2DOTEVK5OILVXN33GB6HN4QHHYLDN4AFTZZNH476KG3RAWESDUKZNHQW2KJLYMLTBHNJNPSTW33J4MAWWKNHPA"
	secret01 = "E3HL4LDW4KZL3ZV5RPWZPJPQT6KJLWFBFXQ2HJHSX2KZNSEA6KU3HERRY2J7HMNFTHYL5AVLE3BJJ2FPSHY3RKUU4SQIP4ESW64WERGXWXPLLUNL5K53DRVW4643A"
	secret02 = "6OU2XBPNRSZTDXMS46VKJ4VCTC26ZBFJPHUK3LHRXSF27Y4LR7X37PPJVGU53LHCSOGMFPPDSWB7JB5YQLWLBF7TXS2JD4M7QSDNFL7QTSIJP4E7TKPUF4EUTSUPDNEZU7M3Z2MRWLZ3RA5I5OLLV4ULQOEQ"
	secret03 = "OR6O5BU2ZCD6PPEJ6OB2LKW5SXUZ7LJM6KS3ND7PX664ZOTWZOY6JJN24KX3N2FPVPT3BA7RXO6ISLJN26MOLF4O6GDK3AHTQ6S3XY4PW7UITDRA6OUZPCGVU7Z2HHE34KL2G"
	secret04 = "ZC6HDZFVQHH2TWMO6CV3ZPXRVGW3BRVH6G2JDOPCW255LBOWVHXIRC7AUSJMRAGWSXR33I7HWGE5PDOGVLHZXZVLXIQ7FA5MXQ3MTO7LQC4WDRMG6CV2LJO6WUZA"
	secret05 = "E3NK2X7FWS3ONJ5BPTZ3FD5Q3CWNDDDV6C5LRCHRVOC2L2EHV3TZHC265OOJ2M7SW2H3B4NPWOVCXVFM6GE3TD7QWK5ZX4ESWCKO7CFC3ODPDCNCQPRIPN7DS6WTRZMGSXZKJGEK"
)

// parityVector pairs a Config (the fields upstream's constructor was given)
// with the exact canonical otpauth:// URI upstream's toString() produces.
type parityVector struct {
	name string
	cfg  Config
	uri  string
}

// parityVectors are the URI-layer cases the Go port can model (the port has no
// "issuerInLabel: false" concept, so those upstream cases are covered
// separately in the parse-only tests below).
var parityVectors = []parityVector{
	// Fixture 00 - SHA1, default label "OTPAuth", no issuer.
	{
		name: "00-hotp",
		cfg:  Config{Type: "hotp", Account: "OTPAuth", Secret: secret00, Algorithm: "SHA1", Digits: 6, Counter: 0},
		uri:  "otpauth://hotp/OTPAuth?secret=" + secret00 + "&algorithm=SHA1&digits=6&counter=0",
	},
	{
		name: "00-totp",
		cfg:  Config{Type: "totp", Account: "OTPAuth", Secret: secret00, Algorithm: "SHA1", Digits: 6, Period: 5},
		uri:  "otpauth://totp/OTPAuth?secret=" + secret00 + "&algorithm=SHA1&digits=6&period=5",
	},
	// Fixture 01 - SHA256.
	{
		name: "01-hotp",
		cfg:  Config{Type: "hotp", Account: "OTPAuth", Secret: secret01, Algorithm: "SHA256", Digits: 6, Counter: 0},
		uri:  "otpauth://hotp/OTPAuth?secret=" + secret01 + "&algorithm=SHA256&digits=6&counter=0",
	},
	{
		name: "01-totp",
		cfg:  Config{Type: "totp", Account: "OTPAuth", Secret: secret01, Algorithm: "SHA256", Digits: 6, Period: 10},
		uri:  "otpauth://totp/OTPAuth?secret=" + secret01 + "&algorithm=SHA256&digits=6&period=10",
	},
	// Fixture 02 - SHA512.
	{
		name: "02-hotp",
		cfg:  Config{Type: "hotp", Account: "OTPAuth", Secret: secret02, Algorithm: "SHA512", Digits: 6, Counter: 0},
		uri:  "otpauth://hotp/OTPAuth?secret=" + secret02 + "&algorithm=SHA512&digits=6&counter=0",
	},
	{
		name: "02-totp",
		cfg:  Config{Type: "totp", Account: "OTPAuth", Secret: secret02, Algorithm: "SHA512", Digits: 6, Period: 15},
		uri:  "otpauth://totp/OTPAuth?secret=" + secret02 + "&algorithm=SHA512&digits=6&period=15",
	},
	// Fixture 03 - issuer "ACME" present in label.
	{
		name: "03-hotp",
		cfg:  Config{Type: "hotp", Issuer: "ACME", Account: "OTPAuth", Secret: secret03, Algorithm: "SHA1", Digits: 6, Counter: 0},
		uri:  "otpauth://hotp/ACME:OTPAuth?issuer=ACME&secret=" + secret03 + "&algorithm=SHA1&digits=6&counter=0",
	},
	{
		name: "03-totp",
		cfg:  Config{Type: "totp", Issuer: "ACME", Account: "OTPAuth", Secret: secret03, Algorithm: "SHA1", Digits: 6, Period: 30},
		uri:  "otpauth://totp/ACME:OTPAuth?issuer=ACME&secret=" + secret03 + "&algorithm=SHA1&digits=6&period=30",
	},
	// Fixture 04 - custom label "Username", 7 digits, no issuer.
	{
		name: "04-hotp",
		cfg:  Config{Type: "hotp", Account: "Username", Secret: secret04, Algorithm: "SHA1", Digits: 7, Counter: 0},
		uri:  "otpauth://hotp/Username?secret=" + secret04 + "&algorithm=SHA1&digits=7&counter=0",
	},
	{
		name: "04-totp",
		cfg:  Config{Type: "totp", Account: "Username", Secret: secret04, Algorithm: "SHA1", Digits: 7, Period: 30},
		uri:  "otpauth://totp/Username?secret=" + secret04 + "&algorithm=SHA1&digits=7&period=30",
	},
	// Fixture 05 - issuer and label both contain spaces (encoded as %20), 8 digits.
	{
		name: "05-hotp",
		cfg:  Config{Type: "hotp", Issuer: "ACME Co", Account: "Firstname Lastname", Secret: secret05, Algorithm: "SHA1", Digits: 8, Counter: 0},
		uri:  "otpauth://hotp/ACME%20Co:Firstname%20Lastname?issuer=ACME%20Co&secret=" + secret05 + "&algorithm=SHA1&digits=8&counter=0",
	},
	{
		name: "05-totp",
		cfg:  Config{Type: "totp", Issuer: "ACME Co", Account: "Firstname Lastname", Secret: secret05, Algorithm: "SHA1", Digits: 8, Period: 30},
		uri:  "otpauth://totp/ACME%20Co:Firstname%20Lastname?issuer=ACME%20Co&secret=" + secret05 + "&algorithm=SHA1&digits=8&period=30",
	},
}

// TestParityStringify checks URL(cfg) reproduces upstream's canonical
// toString() output byte-for-byte.
func TestParityStringify(t *testing.T) {
	for _, v := range parityVectors {
		t.Run(v.name, func(t *testing.T) {
			got := URL(v.cfg)
			if got != v.uri {
				t.Errorf("URL mismatch\n got: %s\nwant: %s", got, v.uri)
			}
		})
	}
}

// TestParityParse checks Parse of upstream's canonical URI recovers the
// modelled fields (the inverse of stringify).
func TestParityParse(t *testing.T) {
	for _, v := range parityVectors {
		t.Run(v.name, func(t *testing.T) {
			got, err := Parse(v.uri)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if got.Type != v.cfg.Type {
				t.Errorf("Type: got %q want %q", got.Type, v.cfg.Type)
			}
			if got.Issuer != v.cfg.Issuer {
				t.Errorf("Issuer: got %q want %q", got.Issuer, v.cfg.Issuer)
			}
			if got.Account != v.cfg.Account {
				t.Errorf("Account: got %q want %q", got.Account, v.cfg.Account)
			}
			if got.Secret != v.cfg.Secret {
				t.Errorf("Secret: got %q want %q", got.Secret, v.cfg.Secret)
			}
			if got.Algorithm != v.cfg.Algorithm {
				t.Errorf("Algorithm: got %q want %q", got.Algorithm, v.cfg.Algorithm)
			}
			if got.Digits != v.cfg.Digits {
				t.Errorf("Digits: got %d want %d", got.Digits, v.cfg.Digits)
			}
			if got.Period != v.cfg.Period {
				t.Errorf("Period: got %d want %d", got.Period, v.cfg.Period)
			}
			if got.Counter != v.cfg.Counter {
				t.Errorf("Counter: got %d want %d", got.Counter, v.cfg.Counter)
			}
		})
	}
}

// TestParityParseIgnoresExtraParams mirrors upstream fixture 02's alternate
// HOTP form, which appends an unrecognized "&extra=0" parameter that the
// parser must ignore.
func TestParityParseExtraParam(t *testing.T) {
	uri := "otpauth://hotp/OTPAuth?secret=" + secret02 + "&algorithm=SHA512&digits=6&counter=0&extra=0"
	got, err := Parse(uri)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got.Secret != secret02 || got.Algorithm != "SHA512" || got.Digits != 6 || got.Counter != 0 {
		t.Errorf("unexpected parse of URI with extra param: %+v", got)
	}
}

// TestParityParseIssuerFromQuery mirrors upstream fixture 06
// (issuerInLabel: false): the label carries no issuer prefix but the issuer
// query parameter is present, so the issuer must be recovered from the query.
func TestParityParseIssuerFromQuery(t *testing.T) {
	uri := "otpauth://hotp/OTPAuth?issuer=ACME&secret=" + secret03 + "&algorithm=SHA1&digits=6&counter=0"
	got, err := Parse(uri)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got.Issuer != "ACME" {
		t.Errorf("Issuer: got %q want %q", got.Issuer, "ACME")
	}
	if got.Account != "OTPAuth" {
		t.Errorf("Account: got %q want %q", got.Account, "OTPAuth")
	}
}
