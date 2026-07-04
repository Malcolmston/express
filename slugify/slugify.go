// Package slugify converts strings into URL-safe slugs, mirroring the behavior
// of the npm "slugify" library: accented Latin characters are transliterated to
// ASCII, whitespace is collapsed to a separator, and disallowed characters are
// removed.
package slugify

import (
	"regexp"
	"strings"
)

// Options configures Slugify. The zero value matches the npm defaults except
// that Trim defaults to true, so callers that want the default behavior should
// use Options{Trim: true} or rely on the single-argument default in Slugify.
type Options struct {
	// Lower lowercases the resulting slug. npm slugify does not lowercase by
	// default.
	Lower bool
	// Separator replaces whitespace runs. Defaults to "-".
	Separator string
	// Strict removes every character that is not alphanumeric or whitespace
	// (before whitespace is turned into separators).
	Strict bool
	// Trim trims leading and trailing whitespace before joining. Defaults to
	// true when using the no-Options form of Slugify.
	Trim bool
}

var (
	removeRe = regexp.MustCompile(`[^\w\s]`)
	strictRe = regexp.MustCompile(`[^A-Za-z0-9\s]`)
	spaceRe  = regexp.MustCompile(`\s+`)
)

// charMap transliterates common accented Latin characters and a handful of
// symbols to ASCII equivalents.
var charMap = map[rune]string{
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Ä': "A", 'Å': "A", 'Æ': "AE",
	'Ç': "C", 'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E",
	'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'Ð': "D", 'Ñ': "N",
	'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ö': "O", 'Ø': "O",
	'Ù': "U", 'Ú': "U", 'Û': "U", 'Ü': "U",
	'Ý': "Y", 'Þ': "TH", 'ß': "ss",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae",
	'ç': "c", 'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'ð': "d", 'ñ': "n",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
	'ý': "y", 'þ': "th", 'ÿ': "y",
	'œ': "oe", 'Œ': "OE",
	// Symbol replacements.
	'&': "and", '@': "at", '#': "number", '%': "percent",
	'<': "less", '>': "greater", '|': "or", '$': "dollar",
	'€': "euro", '£': "pound", '©': "c", '®': "r",
}

// Slugify converts s into a slug. Called with no Options it uses the library
// defaults (separator "-", trimming enabled, no lowercasing). Called with an
// explicit Options, the provided value is used verbatim except that an empty
// Separator falls back to "-".
func Slugify(s string, opts ...Options) string {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	} else {
		o.Trim = true
	}
	if o.Separator == "" {
		o.Separator = "-"
	}

	var b strings.Builder
	for _, ch := range s {
		var appended string
		if mapped, ok := charMap[ch]; ok {
			appended = mapped
		} else {
			appended = string(ch)
		}
		// If the mapped char is the separator, treat it as a space so that
		// separator runs collapse naturally.
		if appended == o.Separator {
			appended = " "
		}
		b.WriteString(removeRe.ReplaceAllString(appended, ""))
	}
	slug := b.String()

	if o.Strict {
		slug = strictRe.ReplaceAllString(slug, "")
	}
	if o.Trim {
		slug = strings.TrimSpace(slug)
	}
	slug = spaceRe.ReplaceAllString(slug, o.Separator)

	// Collapse repeated separators into a single one.
	if o.Separator != "" {
		dup := o.Separator + o.Separator
		for strings.Contains(slug, dup) {
			slug = strings.ReplaceAll(slug, dup, o.Separator)
		}
	}

	if o.Lower {
		slug = strings.ToLower(slug)
	}
	return slug
}
