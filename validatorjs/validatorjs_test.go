package validatorjs

import "testing"

type strCase struct {
	in   string
	want bool
}

func runStr(t *testing.T, name string, fn func(string) bool, cases []strCase) {
	t.Helper()
	for _, c := range cases {
		if got := fn(c.in); got != c.want {
			t.Errorf("%s(%q) = %v, want %v", name, c.in, got, c.want)
		}
	}
}

func TestIsEmail(t *testing.T) {
	runStr(t, "IsEmail", IsEmail, []strCase{
		{"foo@bar.com", true},
		{"foo.bar@example.co.uk", true},
		{"a+b@gmail.com", true},
		{"user_name@sub.domain.org", true},
		{"x@y.io", true},
		{"", false},
		{"plainaddress", false},
		{"@no-local.com", false},
		{"no-at-sign.com", false},
		{"foo@bar", false},
		{"foo@bar.c", false},
		{"foo@.com", false},
		{"foo@bar..com", false},
		{".foo@bar.com", false},
		{"foo.@bar.com", false},
		{"foo@bar.com ", false},
	})
}

func TestIsURL(t *testing.T) {
	runStr(t, "IsURL", IsURL, []strCase{
		{"http://example.com", true},
		{"https://example.com/path?q=1", true},
		{"https://sub.example.co.uk:8080/a/b", true},
		{"ftp://files.example.com", true},
		{"http://localhost", true},
		{"http://127.0.0.1:3000", true},
		{"", false},
		{"example.com", false},
		{"http://", false},
		{"http:// space.com", false},
		{"just a string", false},
		{"mailto:foo@bar.com", false},
		{"httptypo://example.com", false},
	})
}

func TestIsUUID(t *testing.T) {
	runStr(t, "IsUUID", IsUUID, []strCase{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"123e4567-e89b-12d3-a456-426614174000", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"A987FBC9-4BED-4078-8F07-9141BA07C9F3", true},
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716", false},
		{"550e8400e29b41d4a716446655440000", false},
		{"g50e8400-e29b-41d4-a716-446655440000", false},
	})
}

func TestIsIP(t *testing.T) {
	runStr(t, "IsIP", IsIP, []strCase{
		{"192.168.0.1", true},
		{"8.8.8.8", true},
		{"::1", true},
		{"2001:db8::1", true},
		{"", false},
		{"999.1.1.1", false},
		{"abc", false},
	})
}

func TestIsIPv4(t *testing.T) {
	runStr(t, "IsIPv4", IsIPv4, []strCase{
		{"0.0.0.0", true},
		{"255.255.255.255", true},
		{"10.0.0.1", true},
		{"", false},
		{"256.1.1.1", false},
		{"1.2.3", false},
		{"::1", false},
		{"2001:db8::1", false},
		{"1.2.3.4.5", false},
	})
}

func TestIsIPv6(t *testing.T) {
	runStr(t, "IsIPv6", IsIPv6, []strCase{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1ff:fe23:4567:890a", true},
		{"::", true},
		{"", false},
		{"192.168.0.1", false},
		{"not-an-ip", false},
		{"2001:db8:::1", false},
	})
}

func TestIsCreditCard(t *testing.T) {
	runStr(t, "IsCreditCard", IsCreditCard, []strCase{
		{"4111111111111111", true},    // Visa
		{"4012888888881881", true},    // Visa
		{"5500005555555559", true},    // MasterCard
		{"340000000000009", true},     // Amex
		{"6011000000000004", true},    // Discover
		{"4111 1111 1111 1111", true}, // spaces
		{"4111-1111-1111-1111", true}, // hyphens
		{"", false},
		{"1234567890123456", false}, // fails prefix/luhn
		{"4111111111111112", false}, // bad luhn
		{"411111111111", false},     // too short
		{"abcd", false},
	})
}

func TestIsJSON(t *testing.T) {
	runStr(t, "IsJSON", IsJSON, []strCase{
		{`{"a":1}`, true},
		{`[1,2,3]`, true},
		{`"string"`, true},
		{`123`, true},
		{`true`, true},
		{`null`, true},
		{"", false},
		{"   ", false},
		{`{a:1}`, false},
		{`{"a":}`, false},
		{`{`, false},
	})
}

func TestIsHexColor(t *testing.T) {
	runStr(t, "IsHexColor", IsHexColor, []strCase{
		{"#fff", true},
		{"#ffffff", true},
		{"#000000", true},
		{"#AbC123", true},
		{"", false},
		{"fff", false},
		{"#ff", false},
		{"#fffff", false},
		{"#gggggg", false},
		{"#1234567", false},
	})
}

func TestIsBase64(t *testing.T) {
	runStr(t, "IsBase64", IsBase64, []strCase{
		{"aGVsbG8=", true},
		{"Zm9vYmFy", true},
		{"TWFu", true},
		{"YQ==", true},
		{"", false},
		{"aGVsbG8", false}, // wrong length
		{"====", false},
		{"a!b@", false},
		{"aGVsbG8==", false},
	})
}

func TestIsAlpha(t *testing.T) {
	runStr(t, "IsAlpha", IsAlpha, []strCase{
		{"abc", true},
		{"ABCdef", true},
		{"Hello", true},
		{"", false},
		{"abc123", false},
		{"hello world", false},
		{"café", false},
	})
}

func TestIsAlphanumeric(t *testing.T) {
	runStr(t, "IsAlphanumeric", IsAlphanumeric, []strCase{
		{"abc123", true},
		{"ABC", true},
		{"123", true},
		{"", false},
		{"abc 123", false},
		{"abc-123", false},
		{"café1", false},
	})
}

func TestIsNumeric(t *testing.T) {
	runStr(t, "IsNumeric", IsNumeric, []strCase{
		{"123", true},
		{"+123", true},
		{"-123", true},
		{"007", true},
		{"", false},
		{"12.3", false},
		{"12a", false},
		{"++1", false},
	})
}

func TestIsInt(t *testing.T) {
	runStr(t, "IsInt", IsInt, []strCase{
		{"0", true},
		{"123", true},
		{"-123", true},
		{"+123", true},
		{"", false},
		{"01", false},
		{"12.0", false},
		{"abc", false},
		{"-", false},
	})
}

func TestIsFloat(t *testing.T) {
	runStr(t, "IsFloat", IsFloat, []strCase{
		{"1.5", true},
		{"-1.5", true},
		{"+0.5", true},
		{"123", true},
		{".5", true},
		{"5.", true},
		{"1e10", true},
		{"1.5e-3", true},
		{"", false},
		{"abc", false},
		{"1.2.3", false},
		{"1e", false},
	})
}

func TestIsMobilePhone(t *testing.T) {
	runStr(t, "IsMobilePhone", IsMobilePhone, []strCase{
		{"1234567", true},
		{"+12345678901", true},
		{"123456789012345", true},
		{"", false},
		{"123456", false},           // too short
		{"1234567890123456", false}, // too long
		{"123-456-7890", false},
		{"+", false},
	})
}

func TestIsSlug(t *testing.T) {
	runStr(t, "IsSlug", IsSlug, []strCase{
		{"hello-world", true},
		{"foo", true},
		{"a-b-c", true},
		{"post-123", true},
		{"", false},
		{"Hello-World", false},
		{"hello--world", false},
		{"-hello", false},
		{"hello-", false},
		{"hello_world", false},
	})
}

func TestIsStrongPassword(t *testing.T) {
	runStr(t, "IsStrongPassword", IsStrongPassword, []strCase{
		{"Abcdef1!", true},
		{"P@ssw0rd", true},
		{"Str0ng#Pass", true},
		{"", false},
		{"abcdefg", false},
		{"Abcdefg1", false}, // no symbol
		{"abcdef1!", false}, // no upper
		{"ABCDEF1!", false}, // no lower
		{"Abcdefg!", false}, // no number
		{"Ab1!", false},     // too short
	})
}

func TestIsMongoId(t *testing.T) {
	runStr(t, "IsMongoId", IsMongoId, []strCase{
		{"507f1f77bcf86cd799439011", true},
		{"AAAAAAAAAAAAAAAAAAAAAAAA", true},
		{"0123456789abcdef01234567", true},
		{"", false},
		{"507f1f77bcf86cd79943901", false},   // 23
		{"507f1f77bcf86cd7994390111", false}, // 25
		{"507f1f77bcf86cd79943901g", false},
	})
}

func TestIsJWT(t *testing.T) {
	runStr(t, "IsJWT", IsJWT, []strCase{
		{"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", true},
		{"a.b.c", true},
		{"abc-_.def-_.ghi-_", true},
		{"", false},
		{"a.b", false},
		{"a.b.c.d", false},
		{"a..c", false},
		{"a.b.c!", false},
	})
}

func TestIsSemVer(t *testing.T) {
	runStr(t, "IsSemVer", IsSemVer, []strCase{
		{"1.0.0", true},
		{"0.0.1", true},
		{"1.2.3-alpha", true},
		{"1.2.3-alpha.1", true},
		{"1.2.3+build.1", true},
		{"1.2.3-rc.1+build.99", true},
		{"", false},
		{"1.0", false},
		{"1", false},
		{"01.0.0", false},
		{"1.0.0-", false},
		{"v1.0.0", false},
	})
}

func TestContains(t *testing.T) {
	cases := []struct {
		s, sub string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello", "", true},
		{"hello", "xyz", false},
		{"", "a", false},
		{"abc", "abc", true},
	}
	for _, c := range cases {
		if got := Contains(c.s, c.sub); got != c.want {
			t.Errorf("Contains(%q, %q) = %v, want %v", c.s, c.sub, got, c.want)
		}
	}
}

func TestIsLength(t *testing.T) {
	cases := []struct {
		s        string
		min, max int
		want     bool
	}{
		{"hello", 1, 10, true},
		{"hello", 5, 5, true},
		{"", 0, 5, true},
		{"héllo", 5, 5, true}, // rune count, not bytes
		{"hello", 6, 10, false},
		{"hello", 1, 4, false},
		{"anything", 0, -1, true}, // no upper bound
		{"", 1, -1, false},
		{"世界", 2, 2, true},
	}
	for _, c := range cases {
		if got := IsLength(c.s, c.min, c.max); got != c.want {
			t.Errorf("IsLength(%q, %d, %d) = %v, want %v", c.s, c.min, c.max, got, c.want)
		}
	}
}
