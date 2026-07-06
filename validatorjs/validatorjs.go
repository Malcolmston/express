// Package validatorjs is a standalone port of the popular npm "validatorjs"
// (and the closely related "validator.js") string validation library to
// idiomatic Go. It depends only on the Go standard library and does not import
// any part of the express package, so it can be used from any program that
// needs to answer simple "is this string a valid X?" questions.
//
// Reach for this package whenever you have a string of unknown provenance -
// form input, a query parameter, a config value, a webhook payload - and you
// need a cheap, allocation-light predicate that tells you whether it is a valid
// e-mail, URL, UUID, credit-card number, IP address, and so on. Every exported
// function takes a single string and returns a bool: true when the input
// satisfies the rule and false otherwise. There are no error returns and no
// panics on ordinary input, which makes the functions convenient to compose
// inside larger validation logic or to pass around as func(string) bool values.
//
// Under the hood most checks are implemented with precompiled package-level
// regular expressions (for example hexColorRe, slugRe, semVerRe, and the
// alpha/alphanumeric/numeric/int/float matchers), which keeps them fast and
// free of per-call compilation cost. A few rules use hand-written logic where a
// regexp would be awkward or wrong: IsEmail splits on the last "@" and
// separately bounds and validates the local part and the domain labels; IsURL
// parses with net/url and then checks the scheme and host; IsIPv4/IsIPv6 defer
// to net.ParseIP; IsJSON round-trips through encoding/json; and IsCreditCard
// strips separators, matches a known card-prefix pattern, and then verifies the
// Luhn checksum. Rune-aware rules such as IsLength and IsStrongPassword count
// runes with unicode/utf8 rather than bytes.
//
// The individual rules carry deliberate semantics and edge cases worth knowing.
// IsEmail enforces RFC-like length limits (254 overall, 64 for the local part,
// 253 for the domain) and rejects leading, trailing, or doubled dots in the
// local part; IsURL accepts only http, https, ftp, and ftps schemes and
// requires a non-empty host (an IP, "localhost", or a dotted domain); IsUUID
// accepts any version in the canonical 8-4-4-4-12 hex layout; IsBase64 requires
// a length that is a multiple of four; IsInt forbids superfluous leading zeros
// while IsNumeric does not; IsStrongPassword requires at least eight characters
// with a lowercase letter, an uppercase letter, a digit, and a symbol; and
// IsLength treats a negative max as "no upper bound". Unlike the npm library,
// these functions are the whole surface: this port does not expose validatorjs
// rule strings like "required|email|min:6" or pipe-delimited rule sets, nor the
// Validator object, its messages, or its option objects.
//
// Parity with the Node originals is close but not bit-for-bit. The goal is to
// match the practical acceptance and rejection behavior of validator.js's most
// commonly used validators for typical inputs, using the same algorithms
// (Luhn, semver.org's recommended regexp, the standard card prefixes) where it
// matters. It differs in that it does not implement locale/option arguments
// (for example IsMobilePhone is a generic "+ and 7-15 digits" check rather than
// a per-locale matcher, and IsEmail has no display-name or allow-IP options),
// and it omits the large catalog of niche validators and all of the sanitizers
// that validator.js ships. For express request/body/query validation with
// chainable per-field rules, see the sibling validator package instead.
package validatorjs

import (
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	// emailRe is a pragmatic e-mail matcher. It is intentionally close to the
	// default validator.js behavior: a local part, an "@", and a domain with
	// at least one dot and a TLD of two or more letters.
	emailUserRe = regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)

	uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

	base64Re = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{4})$`)

	alphaRe = regexp.MustCompile(`^[A-Za-z]+$`)

	alphanumericRe = regexp.MustCompile(`^[A-Za-z0-9]+$`)

	numericRe = regexp.MustCompile(`^[+-]?[0-9]+$`)

	intRe = regexp.MustCompile(`^[+-]?(?:0|[1-9][0-9]*)$`)

	floatRe = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

	mobilePhoneRe = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

	slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

	mongoIdRe = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

	base64URLSegmentRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

	// semVerRe follows the official semver.org recommended regular expression.
	semVerRe = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
)

// IsEmail reports whether s is a valid e-mail address. It accepts a local part
// followed by "@" and a domain that contains at least one dot and a top-level
// domain of two or more letters.
func IsEmail(s string) bool {
	if s == "" || len(s) > 254 {
		return false
	}
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return false
	}
	local := s[:at]
	domain := s[at+1:]
	if len(local) > 64 {
		return false
	}
	if !emailUserRe.MatchString(local) {
		return false
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") || strings.Contains(local, "..") {
		return false
	}
	return isDomain(domain)
}

func isDomain(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return false
			}
		}
	}
	// The TLD must be alphabetic and at least two characters long.
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return false
	}
	for _, r := range tld {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
			return false
		}
	}
	return true
}

// IsURL reports whether s is a valid URL with an http, https, ftp, or ftps
// scheme and a host component.
func IsURL(s string) bool {
	if s == "" || len(s) > 2083 {
		return false
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "ftp", "ftps":
	default:
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	// A host is valid if it is an IP address or a domain name.
	if net.ParseIP(host) != nil {
		return true
	}
	// Allow "localhost" and any dotted domain name.
	if host == "localhost" {
		return true
	}
	return isDomain(host)
}

// IsUUID reports whether s is a valid UUID of any version (1-5), matching the
// canonical 8-4-4-4-12 hexadecimal format.
func IsUUID(s string) bool {
	return uuidRe.MatchString(s)
}

// IsIP reports whether s is a valid IPv4 or IPv6 address.
func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

// IsIPv4 reports whether s is a valid IPv4 address in dotted-decimal notation.
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	// net.ParseIP accepts IPv4 addresses; ensure it is not IPv6 by checking
	// for the presence of a dot and the absence of a colon.
	return strings.Contains(s, ".") && !strings.Contains(s, ":") && ip.To4() != nil
}

// IsIPv6 reports whether s is a valid IPv6 address.
func IsIPv6(s string) bool {
	if !strings.Contains(s, ":") {
		return false
	}
	return net.ParseIP(s) != nil
}

// IsCreditCard reports whether s is a valid credit card number. It passes the
// Luhn checksum and matches a known card prefix (Visa, MasterCard, American
// Express, Diners Club, Discover, or JCB).
func IsCreditCard(s string) bool {
	// Strip spaces and hyphens which are common visual separators.
	cleaned := strings.NewReplacer(" ", "", "-", "").Replace(s)
	if cleaned == "" {
		return false
	}
	for _, r := range cleaned {
		if r < '0' || r > '9' {
			return false
		}
	}
	if !knownCardRe.MatchString(cleaned) {
		return false
	}
	return luhn(cleaned)
}

var knownCardRe = regexp.MustCompile(
	`^(?:4[0-9]{12}(?:[0-9]{3})?` + // Visa
		`|(?:5[1-5][0-9]{14}|2(?:22[1-9]|2[3-9][0-9]|[3-6][0-9]{2}|7[01][0-9]|720)[0-9]{12})` + // MasterCard
		`|3[47][0-9]{13}` + // American Express
		`|3(?:0[0-5]|[68][0-9])[0-9]{11}` + // Diners Club
		`|6(?:011|5[0-9]{2})[0-9]{12}` + // Discover
		`|(?:2131|1800|35\d{3})\d{11})$`) // JCB

func luhn(number string) bool {
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')
		if alt {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alt = !alt
	}
	return sum%10 == 0
}

// IsJSON reports whether s is a valid JSON document.
func IsJSON(s string) bool {
	if strings.TrimSpace(s) == "" {
		return false
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// IsHexColor reports whether s is a valid hexadecimal color, requiring a
// leading "#" and either three or six hexadecimal digits.
func IsHexColor(s string) bool {
	return hexColorRe.MatchString(s)
}

// IsBase64 reports whether s is a valid standard (RFC 4648) base64 string.
func IsBase64(s string) bool {
	if s == "" || len(s)%4 != 0 {
		return false
	}
	return base64Re.MatchString(s)
}

// IsAlpha reports whether s is non-empty and contains only ASCII letters
// (a-z, A-Z).
func IsAlpha(s string) bool {
	return alphaRe.MatchString(s)
}

// IsAlphanumeric reports whether s is non-empty and contains only ASCII
// letters and digits.
func IsAlphanumeric(s string) bool {
	return alphanumericRe.MatchString(s)
}

// IsNumeric reports whether s consists only of digits with an optional leading
// sign.
func IsNumeric(s string) bool {
	return numericRe.MatchString(s)
}

// IsInt reports whether s is a valid integer with an optional leading sign and
// no superfluous leading zeros.
func IsInt(s string) bool {
	return intRe.MatchString(s)
}

// IsFloat reports whether s is a valid floating-point number, allowing an
// optional sign, a decimal point, and scientific notation.
func IsFloat(s string) bool {
	return floatRe.MatchString(s)
}

// IsMobilePhone reports whether s looks like a mobile phone number: an
// optional leading "+" followed by 7 to 15 digits.
func IsMobilePhone(s string) bool {
	return mobilePhoneRe.MatchString(s)
}

// IsSlug reports whether s is a valid slug: lowercase alphanumeric words
// separated by single hyphens, with no leading or trailing hyphen.
func IsSlug(s string) bool {
	return slugRe.MatchString(s)
}

// IsStrongPassword reports whether s is a strong password: at least eight
// characters including at least one lowercase letter, one uppercase letter,
// one digit, and one symbol.
func IsStrongPassword(s string) bool {
	if utf8.RuneCountInString(s) < 8 {
		return false
	}
	var hasLower, hasUpper, hasNumber, hasSymbol bool
	for _, r := range s {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	return hasLower && hasUpper && hasNumber && hasSymbol
}

// IsMongoId reports whether s is a valid MongoDB ObjectId: exactly 24
// hexadecimal characters.
func IsMongoId(s string) bool {
	return mongoIdRe.MatchString(s)
}

// IsJWT reports whether s is a well-formed JSON Web Token: three non-empty
// base64url-encoded segments separated by dots.
func IsJWT(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" || !base64URLSegmentRe.MatchString(p) {
			return false
		}
	}
	return true
}

// IsSemVer reports whether s is a valid Semantic Versioning 2.0.0 version
// string as defined at https://semver.org.
func IsSemVer(s string) bool {
	return semVerRe.MatchString(s)
}

// Contains reports whether s contains the substring substr.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// IsLength reports whether the number of runes in s is between min and max
// inclusive. A negative max means there is no upper bound.
func IsLength(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	if n < min {
		return false
	}
	if max >= 0 && n > max {
		return false
	}
	return true
}
