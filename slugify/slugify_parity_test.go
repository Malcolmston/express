package slugify

// Upstream-parity vectors for simov/slugify.
//
// Every input -> expected pair below is taken verbatim from the ORIGINAL npm
// library's own Mocha test suite and its shipped charmap, fetched from:
//
//   https://raw.githubusercontent.com/simov/slugify/master/test/slugify.js
//   https://raw.githubusercontent.com/simov/slugify/master/test/locales.js
//   https://raw.githubusercontent.com/simov/slugify/master/config/charmap.json
//   https://raw.githubusercontent.com/simov/slugify/master/slugify.js
//
// The Go Options fields map to upstream options as follows:
//   Separator <-> replacement, Lower <-> lower, Strict <-> strict,
//   Trim <-> trim, Remove <-> remove.
//
// Vectors exercising upstream features the Go port does not implement are
// intentionally omitted (see the file-level notes returned by the sync task):
//   - options.replacement: '' (empty replacement): the Go zero/empty Separator
//     falls back to "-", so an explicit empty replacement cannot be expressed.
//   - the `normalize` vector, which relies on NFC normalization of decomposed
//     input (not available in the pure standard library).
//   - locale-specific transliteration (options.locale) and the stateful
//     extend() API, which the Go port does not expose.

import (
	"regexp"
	"testing"
)

func checkParity(t *testing.T, in, want string, opts ...Options) {
	t.Helper()
	got := Slugify(in, opts...)
	if got != want {
		t.Errorf("Slugify(%q, %+v) = %q, want %q", in, opts, got, want)
	}
}

// Source: test/slugify.js — "replace whitespaces with replacement",
// "remove duplicates of the replacement character", "remove trailing space if
// any", "remove not allowed chars".
func TestParityWhitespaceAndRemoval(t *testing.T) {
	checkParity(t, "foo bar baz", "foo-bar-baz")
	checkParity(t, "foo bar baz", "foo_bar_baz", Options{Separator: "_"})
	checkParity(t, "foo , bar", "foo-bar")
	checkParity(t, " foo bar baz ", "foo-bar-baz")
	checkParity(t, "foo, bar baz", "foo-bar-baz")
	checkParity(t, "foo- bar baz", "foo-bar-baz")
	checkParity(t, "foo] bar baz", "foo-bar-baz")
	checkParity(t, "foo  bar--baz", "foo-bar-baz")
}

// Source: test/slugify.js — "leave allowed chars". The default remove pattern
// keeps these punctuation marks.
func TestParityLeaveAllowedChars(t *testing.T) {
	allowed := []string{"*", "+", "~", ".", "(", ")", "'", "\"", "!", ":", "@"}
	for _, sym := range allowed {
		checkParity(t, "foo "+sym+" bar baz", "foo-"+sym+"-bar-baz")
	}
}

// Source: test/slugify.js — "options.replacement", "options.lower",
// "options.strict", "options.strict - remove duplicates ...",
// "options.replacement and options.strict".
func TestParityOptions(t *testing.T) {
	checkParity(t, "foo bar baz", "foo_bar_baz", Options{Separator: "_"})
	checkParity(t, "Foo bAr baZ", "foo-bar-baz", Options{Lower: true})
	checkParity(t, "foo_bar. -@-baz!", "foobar-baz", Options{Strict: true})
	checkParity(t, "foo @ bar", "foo-bar", Options{Strict: true})
	checkParity(t, "foo_@_bar-baz!", "foo_barbaz", Options{Separator: "_", Strict: true})
}

// Source: test/slugify.js — "options.remove" and
// "options.remove regex without g flag".
func TestParityOptionsRemove(t *testing.T) {
	checkParity(t, "foo *+~.() bar '\"!:@ baz", "foo-bar-baz",
		Options{Remove: regexp.MustCompile(`[$*_+~.()'"!\-:@]`)})
	checkParity(t, "foo bar, bar foo, foo bar", "foo-bar-bar-foo-foo-bar",
		Options{Remove: regexp.MustCompile(`[^a-zA-Z0-9 -]`)})
}

// Source: test/slugify.js — "replaces leading and trailing replacement chars",
// "replaces leading and trailing replacement chars in strict mode", and
// "should preserve leading/trailing replacement characters if option set".
func TestParityLeadingTrailing(t *testing.T) {
	checkParity(t, "-Come on, fhqwhgads-", "Come-on-fhqwhgads")
	// Upstream defaults trim=true even when other options are supplied; the Go
	// port only defaults Trim in the no-Options form, so the faithful
	// equivalent of upstream's { strict: true } here is Options{Strict, Trim}.
	checkParity(t, "! Come on, fhqwhgads !", "Come-on-fhqwhgads", Options{Strict: true, Trim: true})
	checkParity(t, " foo bar baz ", "-foo-bar-baz-", Options{Trim: false, Separator: "-"})
}

// Source: test/slugify.js — "replace custom characters" and issue #144, both
// asserting DEFAULT behavior after the extend() cache is reset.
func TestParityDefaultSymbols(t *testing.T) {
	checkParity(t, "unicode ♥ is ☢", "unicode-love-is") // ♥ -> love, ☢ unmapped -> removed
	checkParity(t, "day + night", "day-+-night")        // + is an allowed char
}

// Source: test/slugify.js — "replace latin chars". charMap verbatim from the
// upstream test's expected values.
func TestParityLatin(t *testing.T) {
	m := map[string]string{
		"À": "A", "Á": "A", "Â": "A", "Ã": "A", "Ä": "A", "Å": "A", "Æ": "AE",
		"Ç": "C", "È": "E", "É": "E", "Ê": "E", "Ë": "E", "Ì": "I", "Í": "I",
		"Î": "I", "Ï": "I", "Ð": "D", "Ñ": "N", "Ò": "O", "Ó": "O", "Ô": "O",
		"Õ": "O", "Ö": "O", "Ő": "O", "Ø": "O", "Ù": "U", "Ú": "U", "Û": "U",
		"Ü": "U", "Ű": "U", "Ý": "Y", "Þ": "TH", "ß": "ss", "à": "a", "á": "a",
		"â": "a", "ã": "a", "ä": "a", "å": "a", "æ": "ae", "ç": "c", "è": "e",
		"é": "e", "ê": "e", "ë": "e", "ì": "i", "í": "i", "î": "i", "ï": "i",
		"ð": "d", "ñ": "n", "ò": "o", "ó": "o", "ô": "o", "õ": "o", "ö": "o",
		"ő": "o", "ø": "o", "ù": "u", "ú": "u", "û": "u", "ü": "u", "ű": "u",
		"ý": "y", "þ": "th", "ÿ": "y", "ẞ": "SS",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace greek chars".
func TestParityGreek(t *testing.T) {
	m := map[string]string{
		"α": "a", "β": "b", "γ": "g", "δ": "d", "ε": "e", "ζ": "z", "η": "h", "θ": "8",
		"ι": "i", "κ": "k", "λ": "l", "μ": "m", "ν": "n", "ξ": "3", "ο": "o", "π": "p",
		"ρ": "r", "σ": "s", "τ": "t", "υ": "y", "φ": "f", "χ": "x", "ψ": "ps", "ω": "w",
		"ά": "a", "έ": "e", "ί": "i", "ό": "o", "ύ": "y", "ή": "h", "ώ": "w", "ς": "s",
		"ϊ": "i", "ΰ": "y", "ϋ": "y", "ΐ": "i",
		"Α": "A", "Β": "B", "Γ": "G", "Δ": "D", "Ε": "E", "Ζ": "Z", "Η": "H", "Θ": "8",
		"Ι": "I", "Κ": "K", "Λ": "L", "Μ": "M", "Ν": "N", "Ξ": "3", "Ο": "O", "Π": "P",
		"Ρ": "R", "Σ": "S", "Τ": "T", "Υ": "Y", "Φ": "F", "Χ": "X", "Ψ": "PS", "Ω": "W",
		"Ά": "A", "Έ": "E", "Ί": "I", "Ό": "O", "Ύ": "Y", "Ή": "H", "Ώ": "W", "Ϊ": "I",
		"Ϋ": "Y",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace turkish chars".
func TestParityTurkish(t *testing.T) {
	m := map[string]string{
		"ş": "s", "Ş": "S", "ı": "i", "İ": "I", "ç": "c", "Ç": "C", "ü": "u", "Ü": "U",
		"ö": "o", "Ö": "O", "ğ": "g", "Ğ": "G",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace cyrillic chars". Empty-mapped chars (ь/Ь)
// collapse away, per the upstream test's special-casing.
func TestParityCyrillic(t *testing.T) {
	m := map[string]string{
		"а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "е": "e", "ё": "yo", "ж": "zh",
		"з": "z", "и": "i", "й": "j", "к": "k", "л": "l", "м": "m", "н": "n", "о": "o",
		"п": "p", "р": "r", "с": "s", "т": "t", "у": "u", "ф": "f", "х": "h", "ц": "c",
		"ч": "ch", "ш": "sh", "щ": "sh", "ъ": "u", "ы": "y", "ь": "", "э": "e", "ю": "yu",
		"я": "ya",
		"А": "A", "Б": "B", "В": "V", "Г": "G", "Д": "D", "Е": "E", "Ё": "Yo", "Ж": "Zh",
		"З": "Z", "И": "I", "Й": "J", "К": "K", "Л": "L", "М": "M", "Н": "N", "О": "O",
		"П": "P", "Р": "R", "С": "S", "Т": "T", "У": "U", "Ф": "F", "Х": "H", "Ц": "C",
		"Ч": "Ch", "Ш": "Sh", "Щ": "Sh", "Ъ": "U", "Ы": "Y", "Ь": "", "Э": "E", "Ю": "Yu",
		"Я": "Ya", "Є": "Ye", "І": "I", "Ї": "Yi", "Ґ": "G", "є": "ye", "і": "i",
		"ї": "yi", "ґ": "g",
	}
	for ch, want := range m {
		expected := "foo-" + want + "-bar-baz"
		if want == "" {
			expected = "foo-bar-baz"
		}
		checkParity(t, "foo "+ch+" bar baz", expected)
	}
}

// Source: test/slugify.js — "replace kazakh cyrillic chars".
func TestParityKazakh(t *testing.T) {
	m := map[string]string{
		"Ә": "AE", "ә": "ae", "Ғ": "GH", "ғ": "gh", "Қ": "KH", "қ": "kh", "Ң": "NG", "ң": "ng",
		"Ү": "UE", "ү": "ue", "Ұ": "U", "ұ": "u", "Һ": "H", "һ": "h", "Ө": "OE", "ө": "oe",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace czech chars".
func TestParityCzech(t *testing.T) {
	m := map[string]string{
		"č": "c", "ď": "d", "ě": "e", "ň": "n", "ř": "r", "š": "s", "ť": "t", "ů": "u",
		"ž": "z", "Č": "C", "Ď": "D", "Ě": "E", "Ň": "N", "Ř": "R", "Š": "S", "Ť": "T",
		"Ů": "U", "Ž": "Z",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace polish chars".
func TestParityPolish(t *testing.T) {
	m := map[string]string{
		"ą": "a", "ć": "c", "ę": "e", "ł": "l", "ń": "n", "ó": "o", "ś": "s", "ź": "z",
		"ż": "z", "Ą": "A", "Ć": "C", "Ę": "E", "Ł": "L", "Ń": "N", "Ś": "S",
		"Ź": "Z", "Ż": "Z",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace latvian chars".
func TestParityLatvian(t *testing.T) {
	m := map[string]string{
		"ā": "a", "č": "c", "ē": "e", "ģ": "g", "ī": "i", "ķ": "k", "ļ": "l", "ņ": "n",
		"š": "s", "ū": "u", "ž": "z", "Ā": "A", "Č": "C", "Ē": "E", "Ģ": "G", "Ī": "I",
		"Ķ": "K", "Ļ": "L", "Ņ": "N", "Š": "S", "Ū": "U", "Ž": "Z",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace serbian chars".
func TestParitySerbian(t *testing.T) {
	m := map[string]string{
		"đ": "dj", "ǌ": "nj", "ǉ": "lj", "Đ": "DJ", "ǋ": "NJ", "ǈ": "LJ", "ђ": "dj", "ј": "j",
		"љ": "lj", "њ": "nj", "ћ": "c", "џ": "dz", "Ђ": "DJ", "Ј": "J", "Љ": "LJ", "Њ": "NJ",
		"Ћ": "C", "Џ": "DZ",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace currencies". Spaces in the expected value
// become the separator (the upstream test replaces the first space only, but
// each expected value here contains at most one space).
func TestParityCurrencies(t *testing.T) {
	m := map[string]string{
		"€": "euro", "₢": "cruzeiro", "₣": "french-franc", "£": "pound",
		"₤": "lira", "₥": "mill", "₦": "naira", "₧": "peseta", "₨": "rupee",
		"₩": "won", "₪": "new-shequel", "₫": "dong", "₭": "kip", "₮": "tugrik", "₸": "kazakhstani-tenge",
		"₯": "drachma", "₰": "penny", "₱": "peso", "₲": "guarani", "₳": "austral",
		"₴": "hryvnia", "₵": "cedi", "¢": "cent", "¥": "yen", "元": "yuan",
		"円": "yen", "﷼": "rial", "₠": "ecu", "¤": "currency", "฿": "baht",
		"$": "dollar", "₽": "russian-ruble", "₿": "bitcoin", "₺": "turkish-lira",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}

// Source: test/slugify.js — "replace symbols".
func TestParitySymbols(t *testing.T) {
	m := map[string]string{
		"©": "(c)", "œ": "oe", "Œ": "OE", "∑": "sum", "®": "(r)", "†": "+",
		"“": "\"", "”": "\"", "‘": "'", "’": "'", "∂": "d", "ƒ": "f", "™": "tm",
		"℠": "sm", "…": "...", "˚": "o", "º": "o", "ª": "a", "•": "*",
		"∆": "delta", "∞": "infinity", "♥": "love", "&": "and", "|": "or",
		"<": "less", ">": "greater",
	}
	for ch, want := range m {
		checkParity(t, "foo "+ch+" bar baz", "foo-"+want+"-bar-baz")
	}
}
