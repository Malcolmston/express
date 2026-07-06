package validatorjs_test

import (
	"fmt"

	"github.com/malcolmston/express/validatorjs"
)

// ExampleIsEmail shows the e-mail predicate accepting a well-formed address and
// rejecting a malformed one. IsEmail requires a local part, a single "@", and a
// domain with at least one dot and an alphabetic top-level domain of two or more
// letters. The first input satisfies all of those requirements and returns
// true. The second input has no domain dot and no "@" structure at all, so it
// returns false. The Output block records both boolean results.
func ExampleIsEmail() {
	fmt.Println(validatorjs.IsEmail("ada@example.com"))
	fmt.Println(validatorjs.IsEmail("not-an-email"))
	// Output:
	// true
	// false
}

// ExampleIsURL demonstrates URL validation across an accepted scheme and a
// rejected one. IsURL parses the string with net/url and requires one of the
// http, https, ftp, or ftps schemes together with a non-empty host. The HTTPS
// address with a dotted domain and a path is accepted. A "javascript:" pseudo
// URL has an unsupported scheme and is rejected. The two results are printed in
// order.
func ExampleIsURL() {
	fmt.Println(validatorjs.IsURL("https://example.com/path?q=1"))
	fmt.Println(validatorjs.IsURL("javascript:alert(1)"))
	// Output:
	// true
	// false
}

// ExampleIsUUID validates a canonical UUID against a truncated string. IsUUID
// accepts any version in the standard 8-4-4-4-12 hexadecimal layout without
// checking the version or variant nibbles. The first value is a properly shaped
// UUID and returns true. The second value is too short to match the pattern and
// returns false. The example prints both outcomes so the behavior is explicit.
func ExampleIsUUID() {
	fmt.Println(validatorjs.IsUUID("123e4567-e89b-12d3-a456-426614174000"))
	fmt.Println(validatorjs.IsUUID("123e4567-e89b"))
	// Output:
	// true
	// false
}

// ExampleIsCreditCard exercises the Luhn-plus-prefix credit-card check. The
// function strips spaces and hyphens, confirms the digits match a known card
// prefix (Visa, MasterCard, Amex, Diners, Discover, or JCB), and then verifies
// the Luhn checksum. The first number is a classic Visa test number that both
// matches the prefix and passes the checksum. The second alters a digit so the
// checksum fails. Both boolean results are shown below.
func ExampleIsCreditCard() {
	fmt.Println(validatorjs.IsCreditCard("4111 1111 1111 1111"))
	fmt.Println(validatorjs.IsCreditCard("4111 1111 1111 1112"))
	// Output:
	// true
	// false
}

// ExampleIsInt highlights the strict integer rule and its treatment of leading
// zeros. IsInt allows an optional leading sign but forbids superfluous leading
// zeros, so "42" and "-7" are valid while "007" is not. This is deliberately
// stricter than IsNumeric, which would accept the zero-padded form. The example
// checks three inputs to make the distinction visible. The Output block lists
// the results in order.
func ExampleIsInt() {
	fmt.Println(validatorjs.IsInt("42"))
	fmt.Println(validatorjs.IsInt("-7"))
	fmt.Println(validatorjs.IsInt("007"))
	// Output:
	// true
	// false? no
}

// ExampleIsStrongPassword checks the composite strength rule. IsStrongPassword
// requires at least eight characters and demands the presence of a lowercase
// letter, an uppercase letter, a digit, and a symbol. The first password meets
// every requirement and returns true. The second is all lowercase letters and
// therefore fails several of the character-class checks. Both results are
// printed so the passing and failing cases are clear.
func ExampleIsStrongPassword() {
	fmt.Println(validatorjs.IsStrongPassword("Passw0rd!"))
	fmt.Println(validatorjs.IsStrongPassword("password"))
	// Output:
	// true
	// false
}

// ExampleIsLength measures string length in runes rather than bytes and honors
// an inclusive minimum and maximum. Passing a negative max removes the upper
// bound entirely. Here a five-rune string is checked against a 1-to-10 window
// and against a lower bound of 6 with no upper bound. The first check passes and
// the second fails because five is below the minimum. The example prints both
// boolean answers.
func ExampleIsLength() {
	fmt.Println(validatorjs.IsLength("hello", 1, 10))
	fmt.Println(validatorjs.IsLength("hello", 6, -1))
	// Output:
	// true
	// false
}
