// Package str ports the "String" category of the npm lodash library to Go. It
// provides the familiar lodash string helpers: the case converters (CamelCase,
// KebabCase, SnakeCase, StartCase, LowerCase, UpperCase, Capitalize,
// UpperFirst, LowerFirst), the word splitter Words, the padding and repeating
// helpers (Pad, PadStart, PadEnd, Repeat), the trimming family (Trim,
// TrimStart, TrimEnd), predicates (StartsWith, EndsWith), HTML entity handling
// (Escape, Unescape), diacritic folding (Deburr), plus Truncate, Replace and a
// JavaScript-style ParseInt. Every function operates on plain Go strings and
// depends only on the standard library.
//
// Use this package when you are porting front-end JavaScript that relied on
// lodash string utilities, or whenever you want their concise, well-defined
// behavior in Go: turning arbitrary identifiers or user input into a consistent
// case (slugs, config keys, display labels), normalizing accented text for
// search or comparison, padding and truncating strings for fixed-width output,
// or escaping text for safe inclusion in HTML. The helpers are Unicode-aware
// where it matters — lengths, slicing and padding are measured in runes rather
// than bytes — so multi-byte input is handled the way lodash measures strings
// by code point.
//
// The heart of the package is Words, and understanding it explains most of the
// rest. Words scans the input rune by rune and emits maximal runs of letters or
// digits, treating any non-alphanumeric rune as a separator; it also splits on
// case boundaries so that camelCase, PascalCase, snake_case, kebab-case and
// spaced text all decompose into the same word list. A run of upper-case
// letters is kept together as an acronym, except that a trailing upper-case
// letter immediately followed by a lower-case letter starts the next word, so
// "XMLHttpTest" splits into "XML", "Http", "Test". The case converters are all
// built on a shared compounder that first runs Deburr and strips apostrophes,
// then splits with Words, then rejoins the words with a per-converter combiner
// (CamelCase lower-cases each word and upper-cases the first letter of all but
// the first; KebabCase and SnakeCase lower-case and join with "-" or "_";
// StartCase upper-cases each word's first letter and joins with spaces).
//
// Edge cases follow lodash. Empty input yields an empty result from every
// function, and the case converters collapse leading, trailing and repeated
// separators (so "--foo-bar--" and "__FOO_BAR__" both reduce cleanly). The pad
// helpers return the string unchanged when it already meets or exceeds the
// requested length, default to a single space when the chars argument is empty,
// and truncate the padding pattern to fit exactly. The trim helpers strip
// Unicode whitespace by default, or exactly the set of runes in chars when one
// is given. StartsWith and EndsWith interpret their position argument in runes,
// with EndsWith treating a negative position as the end of the string. ParseInt
// mirrors JavaScript's parseInt: it honours leading whitespace and a sign,
// auto-detects a "0x" hex prefix when the radix is 0, and stops at the first
// character that is not a valid digit for the radix (returning 0 when nothing
// parses). Deburr covers the Latin-1 Supplement and the more common Latin
// Extended-A letters and drops combining diacritical marks (U+0300..U+036F).
//
// Parity with Node's lodash is high for the covered functions, with a few
// deliberate boundaries. Deburr's mapping table is a representative subset of
// lodash's full Latin transliteration rather than an exhaustive copy, so exotic
// code points may pass through unchanged. Escape and Unescape handle the same
// five HTML characters lodash does (& < > " '); they are not general-purpose
// HTML sanitizers. This package covers the pure string utilities and omits
// lodash template compilation and the various pluralization/inflection helpers
// that were never part of lodash's string category.
package str

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// runeSlice returns the sub-slice of s covering rune indices [start, end).
// Working in runes keeps the case/pad helpers Unicode-aware, mirroring the way
// lodash measures string length by code point.
func runeSlice(s string, start, end int) string {
	r := []rune(s)
	if start < 0 {
		start = 0
	}
	if end > len(r) {
		end = len(r)
	}
	if start >= end {
		return ""
	}
	return string(r[start:end])
}

// UpperFirst converts the first character of string to upper case, leaving the
// remainder untouched.
//
//	UpperFirst("fred") => "Fred"
//	UpperFirst("FRED") => "FRED"
func UpperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// LowerFirst converts the first character of string to lower case, leaving the
// remainder untouched.
//
//	LowerFirst("Fred") => "fred"
//	LowerFirst("FRED") => "fRED"
func LowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

// Capitalize converts the first character of string to upper case and the
// remaining to lower case.
//
//	Capitalize("FRED") => "Fred"
func Capitalize(s string) string {
	return UpperFirst(strings.ToLower(s))
}

// Words splits string into an array of its words. It understands camelCase,
// snake_case, kebab-case, spaced text, digits and upper-case acronyms.
//
//	Words("fooBar")      => ["foo", "Bar"]
//	Words("XMLHttpTest") => ["XML", "Http", "Test"]
//	Words("foo_bar-baz") => ["foo", "bar", "baz"]
func Words(s string) []string {
	isLower := func(r rune) bool { return r >= 'a' && r <= 'z' }
	isUpper := func(r rune) bool { return r >= 'A' && r <= 'Z' }
	isDigit := func(r rune) bool { return r >= '0' && r <= '9' }
	isAlpha := func(r rune) bool { return isLower(r) || isUpper(r) }

	runes := []rune(s)
	n := len(runes)
	var words []string
	i := 0
	for i < n {
		r := runes[i]
		if !isAlpha(r) && !isDigit(r) {
			i++
			continue
		}
		start := i
		switch {
		case isUpper(r):
			i++
			for i < n && isUpper(runes[i]) {
				i++
			}
			// An acronym directly followed by a lower-case letter yields two
			// words: the trailing upper-case letter begins the next word
			// (e.g. "XMLHttp" => "XML", "Http").
			if i-start > 1 && i < n && isLower(runes[i]) {
				i--
			}
			for i < n && isLower(runes[i]) {
				i++
			}
		case isLower(r):
			i++
			for i < n && isLower(runes[i]) {
				i++
			}
		default: // digit
			i++
			for i < n && isDigit(runes[i]) {
				i++
			}
		}
		words = append(words, string(runes[start:i]))
	}
	return words
}

var reApostrophe = regexp.MustCompile(`['\x{2019}]`)

// compound splits string (after deburring and dropping apostrophes) into words
// and reduces them with the supplied combiner, mirroring lodash's internal
// createCompounder helper used by the case converters.
func compound(s string, combine func(result, word string, index int) string) string {
	cleaned := reApostrophe.ReplaceAllString(Deburr(s), "")
	result := ""
	for i, w := range Words(cleaned) {
		result = combine(result, w, i)
	}
	return result
}

// CamelCase converts string to camel case.
//
//	CamelCase("Foo Bar")   => "fooBar"
//	CamelCase("--foo-bar--") => "fooBar"
//	CamelCase("__FOO_BAR__") => "fooBar"
func CamelCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		word = strings.ToLower(word)
		if index > 0 {
			word = UpperFirst(word)
		}
		return result + word
	})
}

// KebabCase converts string to kebab case.
//
//	KebabCase("Foo Bar")   => "foo-bar"
//	KebabCase("fooBar")    => "foo-bar"
//	KebabCase("__FOO_BAR__") => "foo-bar"
func KebabCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		if index > 0 {
			result += "-"
		}
		return result + strings.ToLower(word)
	})
}

// SnakeCase converts string to snake case.
//
//	SnakeCase("Foo Bar")   => "foo_bar"
//	SnakeCase("fooBar")    => "foo_bar"
//	SnakeCase("--FOO-BAR--") => "foo_bar"
func SnakeCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		if index > 0 {
			result += "_"
		}
		return result + strings.ToLower(word)
	})
}

// StartCase converts string to start case, capitalizing the first letter of
// every word while preserving the remaining letters of each word.
//
//	StartCase("--foo-bar--") => "Foo Bar"
//	StartCase("fooBar")      => "Foo Bar"
//	StartCase("XMLHttp")     => "XML Http"
func StartCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		if index > 0 {
			result += " "
		}
		return result + UpperFirst(word)
	})
}

// LowerCase converts string, as space separated words, to lower case.
//
//	LowerCase("--Foo-Bar--") => "foo bar"
//	LowerCase("fooBar")      => "foo bar"
func LowerCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		if index > 0 {
			result += " "
		}
		return result + strings.ToLower(word)
	})
}

// UpperCase converts string, as space separated words, to upper case.
//
//	UpperCase("--foo-bar--") => "FOO BAR"
//	UpperCase("fooBar")      => "FOO BAR"
func UpperCase(s string) string {
	return compound(s, func(result, word string, index int) string {
		if index > 0 {
			result += " "
		}
		return result + strings.ToUpper(word)
	})
}

// createPadding builds padding of the given length (measured in runes) by
// repeating chars and truncating to fit.
func createPadding(length int, chars string) string {
	if length <= 0 || chars == "" {
		return ""
	}
	charsRunes := []rune(chars)
	out := make([]rune, 0, length)
	for len(out) < length {
		out = append(out, charsRunes...)
	}
	return string(out[:length])
}

// Pad pads string on the left and right sides if it is shorter than length.
// Padding characters are truncated if they cannot be evenly divided by length.
//
//	Pad("abc", 8, "_-") => "_-abc_-_"
//	Pad("abc", 3, " ")  => "abc"
func Pad(s string, length int, chars string) string {
	if chars == "" {
		chars = " "
	}
	strLen := utf8.RuneCountInString(s)
	if length <= strLen {
		return s
	}
	mid := length - strLen
	left := mid / 2
	right := mid - left
	return createPadding(left, chars) + s + createPadding(right, chars)
}

// PadStart pads string on the left side if it is shorter than length.
//
//	PadStart("abc", 6, " ")  => "   abc"
//	PadStart("abc", 6, "_-") => "_-_abc"
func PadStart(s string, length int, chars string) string {
	if chars == "" {
		chars = " "
	}
	strLen := utf8.RuneCountInString(s)
	if length <= strLen {
		return s
	}
	return createPadding(length-strLen, chars) + s
}

// PadEnd pads string on the right side if it is shorter than length.
//
//	PadEnd("abc", 6, " ")  => "abc   "
//	PadEnd("abc", 6, "_-") => "abc_-_"
func PadEnd(s string, length int, chars string) string {
	if chars == "" {
		chars = " "
	}
	strLen := utf8.RuneCountInString(s)
	if length <= strLen {
		return s
	}
	return s + createPadding(length-strLen, chars)
}

// Repeat repeats the given string n times. A non-positive n yields "".
//
//	Repeat("*", 3) => "***"
//	Repeat("abc", 0) => ""
func Repeat(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	return strings.Repeat(s, n)
}

// trimSetFunc returns a predicate reporting whether a rune should be trimmed.
// When chars is empty the predicate matches Unicode whitespace, matching the
// default behaviour of lodash's trim family.
func trimSetFunc(chars string) func(rune) bool {
	if chars == "" {
		return unicode.IsSpace
	}
	set := make(map[rune]struct{}, len(chars))
	for _, r := range chars {
		set[r] = struct{}{}
	}
	return func(r rune) bool {
		_, ok := set[r]
		return ok
	}
}

// Trim removes leading and trailing characters (whitespace by default, or any
// of the runes in chars when provided) from string.
//
//	Trim("  abc  ", "")     => "abc"
//	Trim("-_-abc-_-", "_-") => "abc"
func Trim(s string, chars string) string {
	f := trimSetFunc(chars)
	return strings.TrimFunc(s, f)
}

// TrimStart removes leading characters (whitespace by default, or any of the
// runes in chars when provided) from string.
//
//	TrimStart("  abc  ", "")     => "abc  "
//	TrimStart("-_-abc-_-", "_-") => "abc-_-"
func TrimStart(s string, chars string) string {
	f := trimSetFunc(chars)
	return strings.TrimLeftFunc(s, f)
}

// TrimEnd removes trailing characters (whitespace by default, or any of the
// runes in chars when provided) from string.
//
//	TrimEnd("  abc  ", "")     => "  abc"
//	TrimEnd("-_-abc-_-", "_-") => "-_-abc"
func TrimEnd(s string, chars string) string {
	f := trimSetFunc(chars)
	return strings.TrimRightFunc(s, f)
}

// StartsWith reports whether string begins with the given target, testing from
// the supplied position (in runes).
//
//	StartsWith("abc", "a", 0) => true
//	StartsWith("abc", "b", 1) => true
func StartsWith(s, target string, position int) bool {
	if position < 0 {
		position = 0
	}
	r := []rune(s)
	if position > len(r) {
		return false
	}
	return strings.HasPrefix(string(r[position:]), target)
}

// EndsWith reports whether string ends with the given target, testing up to the
// supplied position (in runes). A negative position is treated as the end of
// the string.
//
//	EndsWith("abc", "c", -1) => true
//	EndsWith("abc", "b", 2)  => true
func EndsWith(s, target string, position int) bool {
	r := []rune(s)
	if position < 0 {
		position = len(r)
	}
	if position > len(r) {
		position = len(r)
	}
	return strings.HasSuffix(string(r[:position]), target)
}

var htmlEscapes = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// Escape converts the characters "&", "<", ">", '"' and "'" in string to their
// corresponding HTML entities.
//
//	Escape("fred, barney, & pebbles") => "fred, barney, &amp; pebbles"
func Escape(s string) string {
	return htmlEscapes.Replace(s)
}

var htmlUnescapes = strings.NewReplacer(
	"&amp;", "&",
	"&lt;", "<",
	"&gt;", ">",
	"&quot;", `"`,
	"&#39;", "'",
)

// Unescape is the inverse of Escape; it converts the HTML entities "&amp;",
// "&lt;", "&gt;", "&quot;" and "&#39;" in string back to their characters.
//
//	Unescape("fred, barney, &amp; pebbles") => "fred, barney, & pebbles"
func Unescape(s string) string {
	return htmlUnescapes.Replace(s)
}

// TruncateOptions configures Truncate. A zero Length falls back to the lodash
// default of 30 and an empty Omission falls back to "...". Separator, when set,
// causes truncation to happen at the last separator match within the retained
// text; SeparatorRegexp takes precedence over Separator when non-nil.
type TruncateOptions struct {
	// Length is the maximum rune length of the result including Omission. A
	// value <= 0 falls back to the lodash default of 30.
	Length int
	// Omission is the string appended to indicate truncation. An empty value
	// falls back to "...".
	Omission string
	// Separator, when non-empty, causes truncation to occur at the last
	// occurrence of this literal substring within the retained text.
	Separator string
	// SeparatorRegexp, when non-nil, causes truncation at the last match of the
	// pattern within the retained text and takes precedence over Separator.
	SeparatorRegexp *regexp.Regexp
}

// Truncate truncates string if it is longer than the requested length. The last
// characters of the truncated string are replaced with the omission string
// (default "..."). Options mirror lodash's _.truncate.
//
//	Truncate("hi-diddly-ho there, neighborino", TruncateOptions{})
//	  => "hi-diddly-ho there, neig..."
func Truncate(s string, opts TruncateOptions) string {
	length := opts.Length
	if length <= 0 {
		length = 30
	}
	omission := opts.Omission
	if omission == "" {
		omission = "..."
	}

	strLen := utf8.RuneCountInString(s)
	if strLen <= length {
		return s
	}

	omissionLen := utf8.RuneCountInString(omission)
	end := length - omissionLen
	if end < 1 {
		return omission
	}

	result := runeSlice(s, 0, end)
	rest := runeSlice(s, end, strLen)

	if opts.SeparatorRegexp != nil {
		// Skip trimming when the separator sits exactly at the cut point.
		if loc := opts.SeparatorRegexp.FindStringIndex(rest); loc == nil || loc[0] != 0 {
			if idx := lastRegexpMatch(opts.SeparatorRegexp, result); idx >= 0 {
				result = result[:idx]
			}
		}
	} else if opts.Separator != "" {
		if !strings.HasPrefix(rest, opts.Separator) {
			if idx := strings.LastIndex(result, opts.Separator); idx >= 0 {
				result = result[:idx]
			}
		}
	}

	return result + omission
}

// lastRegexpMatch returns the byte offset of the final match of re within s, or
// -1 when there is no match.
func lastRegexpMatch(re *regexp.Regexp, s string) int {
	matches := re.FindAllStringIndex(s, -1)
	if len(matches) == 0 {
		return -1
	}
	return matches[len(matches)-1][0]
}

// Replace replaces the first occurrence of old in string with replacement,
// matching JavaScript's String.prototype.replace when called with a string
// pattern.
//
//	Replace("Hi Fred", "Fred", "Barney") => "Hi Barney"
func Replace(s, old, replacement string) string {
	return strings.Replace(s, old, replacement, 1)
}

// ParseInt converts string to an integer of the specified radix. A radix of 0
// selects base 16 for values prefixed with "0x"/"0X" and base 10 otherwise,
// mirroring JavaScript's parseInt. Leading whitespace and a sign are honoured;
// parsing stops at the first character that is not a valid digit.
//
//	ParseInt("08", 10)  => 8
//	ParseInt("0x1A", 0) => 26
//	ParseInt("42px", 10) => 42
func ParseInt(s string, radix int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	sign := 1
	if s[0] == '+' || s[0] == '-' {
		if s[0] == '-' {
			sign = -1
		}
		s = s[1:]
	}

	if radix == 0 {
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			radix = 16
			s = s[2:]
		} else {
			radix = 10
		}
	} else if radix == 16 && (strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X")) {
		s = s[2:]
	}

	if radix < 2 || radix > 36 {
		return 0
	}

	digitVal := func(r rune) int {
		switch {
		case r >= '0' && r <= '9':
			return int(r - '0')
		case r >= 'a' && r <= 'z':
			return int(r-'a') + 10
		case r >= 'A' && r <= 'Z':
			return int(r-'A') + 10
		}
		return -1
	}

	value := 0
	parsedAny := false
	for _, r := range s {
		d := digitVal(r)
		if d < 0 || d >= radix {
			break
		}
		value = value*radix + d
		parsedAny = true
	}
	if !parsedAny {
		return 0
	}
	return sign * value
}

// deburredLetters maps accented Latin characters to their ASCII equivalents.
// It covers the Latin-1 Supplement and the more common Latin Extended-A code
// points handled by lodash's _.deburr.
var deburredLetters = map[rune]string{
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Ä': "A", 'Å': "A",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
	'Ç': "C", 'ç': "c",
	'Ð': "D", 'ð': "d",
	'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'Ñ': "N", 'ñ': "n",
	'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ö': "O", 'Ø': "O",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'Ù': "U", 'Ú': "U", 'Û': "U", 'Ü': "U",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
	'Ý': "Y", 'ý': "y", 'ÿ': "y",
	'Æ': "Ae", 'æ': "ae",
	'Þ': "Th", 'þ': "th",
	'ß': "ss",
	// A representative slice of Latin Extended-A.
	'Ā': "A", 'ā': "a", 'Ă': "A", 'ă': "a", 'Ą': "A", 'ą': "a",
	'Ć': "C", 'ć': "c", 'Ĉ': "C", 'ĉ': "c", 'Ċ': "C", 'ċ': "c", 'Č': "C", 'č': "c",
	'Ď': "D", 'ď': "d", 'Đ': "D", 'đ': "d",
	'Ē': "E", 'ē': "e", 'Ė': "E", 'ė': "e", 'Ě': "E", 'ě': "e",
	'Ĝ': "G", 'ĝ': "g", 'Ğ': "G", 'ğ': "g",
	'Ł': "L", 'ł': "l",
	'Ń': "N", 'ń': "n", 'Ň': "N", 'ň': "n",
	'Ō': "O", 'ō': "o", 'Ő': "O", 'ő': "o",
	'Œ': "Oe", 'œ': "oe",
	'Ŕ': "R", 'ŕ': "r", 'Ř': "R", 'ř': "r",
	'Ś': "S", 'ś': "s", 'Ş': "S", 'ş': "s", 'Š': "S", 'š': "s",
	'Ť': "T", 'ť': "t",
	'Ū': "U", 'ū': "u", 'Ů': "U", 'ů': "u", 'Ű': "U", 'ű': "u",
	'Ź': "Z", 'ź': "z", 'Ż': "Z", 'ż': "z", 'Ž': "Z", 'ž': "z",
}

// Deburr converts the Latin-1 Supplement and common Latin Extended-A accented
// letters in string to their basic Latin equivalents and removes combining
// diacritical marks (U+0300..U+036F).
//
//	Deburr("déjà vu") => "deja vu"
func Deburr(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if repl, ok := deburredLetters[r]; ok {
			b.WriteString(repl)
			continue
		}
		if r >= 0x0300 && r <= 0x036f {
			// Combining diacritical mark: drop it.
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
