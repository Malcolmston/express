// Package titlecase converts strings to Title Case, matching the behavior of
// the npm "title-case" package from the change-case family
// (blakeembrey/change-case). It is a faithful port of that library's algorithm
// (packages/title-case/src/index.ts), preserving the original characters of the
// input and only upper-casing the first letter of each word, with English
// "small word" handling, sentence detection, and special-casing of URLs,
// acronyms, and manually cased words such as "iPhone".
//
// Unlike a naive title-caser, this package never lower-cases letters: acronyms
// like "NASA" and manually cased words like "camelCase" are preserved as-is,
// punctuation and whitespace (including newlines and repeated spaces) are kept
// intact, and small function words such as "a", "of", and "the" are left in
// lower case except at the start or end of a sentence.
//
// The transformation is deterministic and depends only on the input and the
// optional Options.
package titlecase

import "unicode"

// Options configures TitleCase. The zero value produces the default behavior:
// full title casing using the built-in set of small words.
type Options struct {
	// SentenceCase, when true, only capitalizes the first word of each
	// sentence rather than every significant word.
	SentenceCase bool
	// SmallWords overrides the set of words that are kept in lower case mid
	// sentence. A nil map uses the default set (DefaultSmallWords); a non-nil
	// (possibly empty) map replaces it entirely.
	SmallWords map[string]bool
}

// sentenceTerminators marks the end of a sentence. In sentence-case mode this
// is the active terminator set.
var sentenceTerminators = map[rune]bool{
	'.': true, '!': true, '?': true, '\n': true, '\r': true,
}

// titleTerminators is the active terminator set in title-case (default) mode.
// It extends the sentence terminators with a handful of characters that also
// begin a new "title" segment.
var titleTerminators = map[rune]bool{
	'.': true, '!': true, '?': true, '\n': true, '\r': true,
	':': true, '"': true, '\'': true, '”': true, // ”
}

// wordSeparators are the characters that separate the sub-words of a compound
// token such as "step-by-step" or "foo/bar".
var wordSeparators = map[rune]bool{
	'—': true, // em dash
	'–': true, // en dash
	'-': true,
	'―': true, // horizontal bar
	'/': true,
}

// DefaultSmallWords is the set of English function words that are kept in lower
// case mid-sentence. It mirrors the SMALL_WORDS set in the upstream library.
var DefaultSmallWords = map[string]bool{
	"a": true, "an": true, "and": true, "as": true, "at": true,
	"because": true, "but": true, "by": true, "en": true, "for": true,
	"if": true, "in": true, "neither": true, "nor": true, "of": true,
	"on": true, "only": true, "or": true, "over": true, "per": true,
	"so": true, "some": true, "than": true, "that": true, "the": true,
	"to": true, "up": true, "upon": true, "v": true, "versus": true,
	"via": true, "vs": true, "when": true, "with": true, "without": true,
	"yet": true,
}

// TitleCase converts s to Title Case following the upstream change-case
// algorithm. An optional Options value tweaks the behavior; if omitted the
// default title-case behavior is used.
//
// For example:
//
//	TitleCase("one two")                  == "One Two"
//	TitleCase("a small word starts")      == "A Small Word Starts"
//	TitleCase("we keep NASA capitalized") == "We Keep NASA Capitalized"
//	TitleCase("pass camelCase through")   == "Pass camelCase Through"
//	TitleCase("this sub-phrase is nice")  == "This Sub-Phrase Is Nice"
func TitleCase(s string, opts ...Options) string {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	smallWords := opt.SmallWords
	if smallWords == nil {
		smallWords = DefaultSmallWords
	}
	terminators := titleTerminators
	if opt.SentenceCase {
		terminators = sentenceTerminators
	}

	runes := []rune(s)
	total := len(runes)
	var result []rune
	isNewSentence := true

	i := 0
	for i < total {
		r := runes[i]

		// Whitespace is emitted one character at a time and preserved.
		if isSpace(r) {
			result = append(result, r)
			if terminators[r] {
				isNewSentence = true
			}
			i++
			continue
		}

		// Collect a maximal run of non-space characters (a token).
		start := i
		for i < total && !isSpace(runes[i]) {
			i++
		}
		token := runes[start:i]
		tokenEnd := i

		if isSpecialCase(token) {
			if ok, prefixLen, suffixChar, hasSuffix := parseAcronym(token); ok {
				// The period ending an acronym is not a new sentence, but we
				// still upper-case the first letter (e.g. "i.e." -> "I.e.").
				if opt.SentenceCase && !isNewSentence {
					result = append(result, token...)
				} else {
					result = append(result, upperAt(token, prefixLen)...)
				}
				isNewSentence = hasSuffix && terminators[suffixChar]
				continue
			}
			// URLs, emails, "foo-bar.com", "#tag", etc. pass through untouched.
			result = append(result, token...)
			isNewSentence = terminators[token[len(token)-1]]
			continue
		}

		value := make([]rune, len(token))
		copy(value, token)
		words := alnumWords(token)
		isSentenceEnd := false

		for wi, w := range words {
			wordIndex := w.start
			wordEnd := w.start + w.length
			var nextChar rune
			hasNext := wordEnd < len(token)
			if hasNext {
				nextChar = token[wordEnd]
			}
			isSentenceEnd = hasNext && terminators[nextChar]

			skip := false
			switch {
			case isNewSentence:
				// Always capitalize the first word of a sentence.
				isNewSentence = false
			case opt.SentenceCase || isManualCase(token[w.start:wordEnd]):
				// Skip everything but sentence starts in sentence-case mode,
				// and never touch manually cased words such as "iPhone".
				skip = true
			case len(words) == 1:
				// Simple single-word token: keep small words lower-case unless
				// they end a sentence or the whole input.
				if smallWords[string(token[w.start:wordEnd])] {
					isFinalToken := tokenEnd == total
					if !isFinalToken && !isSentenceEnd {
						skip = true
					}
				}
			case wi > 0:
				// Subsequent sub-words of a compound token.
				prevChar := token[wordIndex-1]
				if !wordSeparators[prevChar] {
					// e.g. "apple's" or "test(ing)".
					skip = true
				} else if smallWords[string(token[w.start:wordEnd])] && hasNext && wordSeparators[nextChar] {
					// Small words in the middle of a hyphenated compound.
					skip = true
				}
			}

			if !skip {
				value[wordIndex] = unicode.ToUpper(value[wordIndex])
			}
		}

		result = append(result, value...)
		isNewSentence = isSentenceEnd || terminators[token[len(token)-1]]
	}

	return string(result)
}

// word describes an alphanumeric run within a token by its rune offset and
// length.
type word struct {
	start  int
	length int
}

// alnumWords returns the alphanumeric runs (letters and numbers) within token.
func alnumWords(token []rune) []word {
	var words []word
	i := 0
	for i < len(token) {
		if !isAlnum(token[i]) {
			i++
			continue
		}
		start := i
		for i < len(token) && isAlnum(token[i]) {
			i++
		}
		words = append(words, word{start: start, length: i - start})
	}
	return words
}

// isSpecialCase reports whether the token contains a '.' or '#' immediately
// followed by an alphanumeric character (URLs, hashtags, acronyms, etc.).
func isSpecialCase(token []rune) bool {
	for i := 0; i+1 < len(token); i++ {
		if (token[i] == '.' || token[i] == '#') && isAlnum(token[i+1]) {
			return true
		}
	}
	return false
}

// parseAcronym reports whether token matches the acronym shape
// ^([^L])*(?:L\.){2,}([^L])*$ and, if so, returns the length of the leading
// non-letter prefix (0 or 1, matching the upstream capture semantics), the last
// trailing non-letter character, and whether such a trailing character exists.
func parseAcronym(token []rune) (ok bool, prefixLen int, suffixChar rune, hasSuffix bool) {
	k := 0
	for k < len(token) && !unicode.IsLetter(token[k]) {
		k++
	}
	leading := k

	pairs := 0
	for k+1 < len(token) && unicode.IsLetter(token[k]) && token[k+1] == '.' {
		k += 2
		pairs++
	}
	if pairs < 2 {
		return false, 0, 0, false
	}
	for j := k; j < len(token); j++ {
		if unicode.IsLetter(token[j]) {
			return false, 0, 0, false
		}
	}
	if leading > 0 {
		prefixLen = 1
	}
	if k < len(token) {
		hasSuffix = true
		suffixChar = token[len(token)-1]
	}
	return true, prefixLen, suffixChar, hasSuffix
}

// isManualCase reports whether word contains a lower-case letter immediately
// followed by an upper-case letter (e.g. "iPhone", "camelCase").
func isManualCase(word []rune) bool {
	for i := 0; i+1 < len(word); i++ {
		if unicode.IsLower(word[i]) && unicode.IsUpper(word[i+1]) {
			return true
		}
	}
	return false
}

// upperAt returns a copy of token with the rune at index upper-cased.
func upperAt(token []rune, index int) []rune {
	out := make([]rune, len(token))
	copy(out, token)
	if index >= 0 && index < len(out) {
		out[index] = unicode.ToUpper(out[index])
	}
	return out
}

// isAlnum reports whether r is a Unicode letter or number.
func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// isSpace reports whether r is whitespace, matching JavaScript's \s class.
func isSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ',
		'\u00a0', '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000', '\ufeff':
		return true
	}
	return r >= '\u2000' && r <= '\u200a'
}
