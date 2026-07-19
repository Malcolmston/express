package hotp

import (
	"encoding/hex"
	"testing"
)

// Upstream-parity vectors for the counter-based HOTP primitive.
//
// Two independent canonical sources are encoded here:
//
//  1. RFC 4226 "HOTP: An HMAC-Based One-Time Password Algorithm", Appendix D
//     ("HOTP Algorithm: Test Values"). Secret is the ASCII string
//     "12345678901234567890" (20 bytes); HMAC-SHA1; counters 0..9. The 6-digit
//     column is quoted verbatim from the RFC; the 7- and 8-digit columns are
//     derived from the RFC's "HOTP (decimal)" column (the full 31-bit dynamic
//     truncation value) by reduction modulo 10^7 and 10^8.
//     https://www.rfc-editor.org/rfc/rfc4226#appendix-D
//
//  2. hectorm/otpauth (the npm "otpauth" library this package ports), test
//     suite test/test.mjs at branch master. Each entry pairs the case's raw
//     secret (given here as the suite's hex encoding of the secret bytes) with
//     the SHA-1 HOTP.generate output for the specified counter and digit count.
//     Only the SHA-1 cases are reproduced, since this port implements the
//     standard HMAC-SHA1 path exclusively.
//     https://raw.githubusercontent.com/hectorm/otpauth/master/test/test.mjs

// --- Source 1: RFC 4226 Appendix D -----------------------------------------

func TestParityRFC4226AppendixD6Digit(t *testing.T) {
	secret := []byte("12345678901234567890")
	want := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	for counter, expected := range want {
		if got := Generate(secret, uint64(counter), 6); got != expected {
			t.Errorf("counter %d: got %q, want %q", counter, got, expected)
		}
	}
}

func TestParityRFC4226AppendixDDigitVariants(t *testing.T) {
	secret := []byte("12345678901234567890")
	// Derived from RFC 4226 Appendix D "HOTP (decimal)" column.
	cases := []struct {
		counter uint64
		want7   string
		want8   string
	}{
		{0, "4755224", "84755224"},
		{1, "4287082", "94287082"},
		{2, "7359152", "37359152"},
		{3, "6969429", "26969429"},
		{4, "0338314", "40338314"},
		{5, "8254676", "68254676"},
		{6, "8287922", "18287922"},
		{7, "2162583", "82162583"},
		{8, "3399871", "73399871"},
		{9, "5520489", "45520489"},
	}
	for _, c := range cases {
		if got := Generate(secret, c.counter, 7); got != c.want7 {
			t.Errorf("counter %d 7-digit: got %q, want %q", c.counter, got, c.want7)
		}
		if got := Generate(secret, c.counter, 8); got != c.want8 {
			t.Errorf("counter %d 8-digit: got %q, want %q", c.counter, got, c.want8)
		}
	}
}

// TestParityDefaultDigits verifies the port's documented default: digits <= 0
// is treated as the canonical 6-digit length, matching otpauth's default.
func TestParityDefaultDigits(t *testing.T) {
	secret := []byte("12345678901234567890")
	if got := Generate(secret, 0, 0); got != "755224" {
		t.Errorf("default digits: got %q, want %q", got, "755224")
	}
	if got := Generate(secret, 0, -1); got != "755224" {
		t.Errorf("negative digits: got %q, want %q", got, "755224")
	}
}

// --- Source 2: hectorm/otpauth test/test.mjs (SHA-1 cases) ------------------

var otpauthCases = []struct {
	idx     int
	hexKey  string
	counter uint64
	digits  int
	want    string
}{
	{0, "F3AD8BA2DF81307DF185B88577E68180E68B8657E99EABF192A6ADE69CADC686E992AAEB90BADDBBD983E3B79039F0B1B7802CF3969F9FF28DB882C490E8ACB4F0B69495E18B984ED4B5F29DB7B4F180B594D3BC", 10000000000, 6, "147664"},
	{3, "747CEE869AC887E7BC89F383A5AADD95E99FAD2CF2A5B68FEFBFBDCCBA76CBB1E4A5BAE2AFB6E8AFABE7B083F1BBBC892D2DD798E5978EF186AD80F387A5BBE38FB7E8898E20F3A99788D5A7F3A39C9BE297A3", 10000000000, 6, "361593"},
	{4, "C8BC71E4B581CFA9D98EF0ABBCBEF1A9ADB0C6A7F1B491B9E2B6BBD585D6A9EE888BE0A492C880D695E3BDA3E7B189D78DC6AACF9BE6ABBA21F283ACBC36C9BBEB80B961C586F0ABA5A5DEB532", 10000000000, 7, "8319983"},
	{5, "26DAAD5FE5B4B6E6A7A17CF3B28FB0D8ACD18C75F0BAB888F1AB85A5E887AEE7938B5EEB9C9D33F2B68FB0F1AFB3AA2BD4ACF189B98FF0B2BB9BF092B094EF88A2DB86F189A283E287B7E397AD38E58695F2A4988A", 10000000000, 8, "94726517"},
	{6, "747CEE869AC887E7BC89F383A5AADD95E99FAD2CF2A5B68FEFBFBDCCBA76CBB1E4A5BAE2AFB6E8AFABE7B083F1BBBC892D2DD798E5978EF186AD80F387A5BBE38FB7E8898E20F3A99788D5A7F3A39C9BE297A3", 10000000000, 6, "361593"},
	{7, "F3B78C92DAB5F2A98C9FD896C5A6C493EFBE9F5BE380ABE68D90F483858FECB7AB44E4AA9F55CBB5F1858892F0A59885E0A1A2E8B8ACF483B8B550F3A8A491E89BA9E1A59252ECBFACCEB5EFBFBDF0A782BB256A", 10000000000, 6, "606594"},
	{8, "F09D8C83E5888AC4B3D886F38784A64EE9AE8BEB8783C7BCCA9063E6A98DEAB3B42A52E98BACED8590CE96DC97F3B991A0F0AC81A6E8BBA866C6B7DAA8E69BB072C48DF195ADA9E6AD90F2BA97B0ECA388", 10000000000, 6, "290910"},
	{9, "7CEBB3A66BF2B1BB84E7988B41EA83A1F38CA9A52DF199BF99EFBFBDC3955FE3B399ED8C89E3BC89D987EF9F9C67D7B423EFBFBDE1B6AFF3B582ACF189AD94E399BFE3B594D48BDEBAEFBFBDF19F859DF2B2A596", 10000000000, 6, "851410"},
	{10, "7DD9BFE98DACE18484E1AEA1DDA5F3A4AFA0D299D6A2562EC7A0F1938D8EC2905BF29282A8F29A94B653F0BB8CBBE592A5D6B623F3ACA5BC73F0958780F182A7A7EAA3B8323F4BF1B7B1BE47", 10000000000, 6, "591041"},
	{11, "E1A486C7AD43D48EDFBBCD82D790CB85C2812367C59FE5A1ADC89A2DC39C5D7BD88AE4A5B178ED9299F0958CAFD2B3D491E8A29ED0A56CC5B131EEAC97CF83", 10000000000, 6, "717279"},
}

func TestParityOtpauthGenerate(t *testing.T) {
	for _, c := range otpauthCases {
		key, err := hex.DecodeString(c.hexKey)
		if err != nil {
			t.Fatalf("case %d: bad hex key: %v", c.idx, err)
		}
		if got := Generate(key, c.counter, c.digits); got != c.want {
			t.Errorf("case %d (counter=%d digits=%d): got %q, want %q", c.idx, c.counter, c.digits, got, c.want)
		}
	}
}

func TestParityOtpauthVerify(t *testing.T) {
	for _, c := range otpauthCases {
		key, err := hex.DecodeString(c.hexKey)
		if err != nil {
			t.Fatalf("case %d: bad hex key: %v", c.idx, err)
		}
		if !Verify(key, c.counter, c.want, c.digits) {
			t.Errorf("case %d: expected Verify to accept correct code %q", c.idx, c.want)
		}
		// A code for the neighboring counter must not verify against this one.
		if Verify(key, c.counter+1, c.want, c.digits) {
			t.Errorf("case %d: expected Verify to reject code %q at counter+1", c.idx, c.want)
		}
	}
}
