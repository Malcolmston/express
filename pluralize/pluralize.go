// Package pluralize pluralizes and singularizes English words. It is a faithful
// port of the npm package "pluralize", reproducing its rule set, irregular and
// uncountable word lists, and case-preserving behaviour using only the Go
// standard library (regexp, strconv and strings).
//
// English is used everywhere in software — UI labels, log messages, generated
// API and ORM names — and constantly needs the correct grammatical number: "1
// file" but "3 files", "1 person" but "2 people". Hand-writing that logic is
// error-prone because English pluralization is deeply irregular. This package
// centralizes the rules so callers can convert in either direction (Plural and
// Singular) or merely ask which form a word is already in (IsPlural and
// IsSingular).
//
// Conversion proceeds through three tiers, checked in order. First, uncountable
// words and patterns (fish, sheep, series, information, and patterns such as
// "-ese" nationalities) are returned unchanged in both directions. Second,
// explicit irregular pairs (person/people, mouse/mice, foot/feet, ox/oxen,
// die/dice, ...) are looked up directly. Third, when no special case applies, an
// ordered list of regular regular-expression rules is scanned from the end toward
// the beginning and the first matching rule is applied — this is how the common
// suffix transformations (-s, -es, -ies, -ves, and many Latin/Greek endings like
// -us/-i, -a/-ae, -ix/-ices) are handled.
//
// Case of the input is preserved wherever it is natural. If the input is all
// lower case the result is lower case; all upper case yields upper case; and a
// leading capital ("Bus") produces a capitalized result ("Buses"). This mirrors
// the reference library's restoreCase logic so that pluralizing "Person" gives
// "People" rather than "people". Matching against the rule and word lists is
// always done case-insensitively on a lower-cased token, with casing reapplied to
// the final output.
//
// Edge cases and Node parity: an empty string is returned unchanged. Operations
// are idempotent in the sense that pluralizing an already-plural word, or
// singularizing an already-singular word, leaves it unchanged (Plural("people")
// is "people"). IsPlural and IsSingular are exact complements for countable
// words and both return true for uncountable words, since those forms are
// identical. The rules and word lists are ported verbatim from the JavaScript
// source, so results should agree with pluralize for the shared vocabulary;
// custom-rule registration functions from the Node API are internal helpers here
// rather than an exported extension surface.
package pluralize

import (
	"regexp"
	"strconv"
	"strings"
)

// rule pairs a compiled pattern with its replacement template.
type rule struct {
	re   *regexp.Regexp
	repl string
}

var (
	pluralRules      []rule
	singularRules    []rule
	uncountables     = map[string]bool{}
	irregularSingles = map[string]string{} // singular -> plural
	irregularPlurals = map[string]string{} // plural -> singular
)

var dollarRe = regexp.MustCompile(`\$(\d{1,2})`)

func mk(pattern, repl string) rule {
	return rule{re: regexp.MustCompile("(?i)" + pattern), repl: repl}
}

func addIrregular(single, plural string) {
	irregularSingles[single] = plural
	irregularPlurals[plural] = single
}

// addUncountableWord registers a word that has the same singular and plural.
func addUncountableWord(word string) {
	uncountables[strings.ToLower(word)] = true
}

// addUncountablePattern registers a pattern whose matches are uncountable.
func addUncountablePattern(pattern string) {
	pluralRules = append(pluralRules, mk(pattern, "$0"))
	singularRules = append(singularRules, mk(pattern, "$0"))
}

func init() {
	initIrregulars()
	initPluralRules()
	initSingularRules()
	initUncountables()
}

// restoreCase adapts token to the casing of word.
func restoreCase(word, token string) string {
	if token == "" {
		return token
	}
	if word == token {
		return token
	}
	if word == strings.ToLower(word) {
		return strings.ToLower(token)
	}
	if word == strings.ToUpper(word) {
		return strings.ToUpper(token)
	}
	if word != "" && word[:1] == strings.ToUpper(word[:1]) {
		return strings.ToUpper(token[:1]) + strings.ToLower(token[1:])
	}
	return strings.ToLower(token)
}

// interpolate expands $N references in repl using the submatch indices loc
// against word.
func interpolate(repl, word string, loc []int) string {
	return dollarRe.ReplaceAllStringFunc(repl, func(m string) string {
		n, _ := strconv.Atoi(m[1:])
		if 2*n+1 < len(loc) && loc[2*n] >= 0 {
			return word[loc[2*n]:loc[2*n+1]]
		}
		return ""
	})
}

// replaceRule applies a single rule to word, replacing the first match.
func replaceRule(word string, r rule) string {
	loc := r.re.FindStringSubmatchIndex(word)
	if loc == nil {
		return word
	}
	match := word[loc[0]:loc[1]]
	result := interpolate(r.repl, word, loc)
	var restored string
	if match == "" {
		var ch string
		if loc[0]-1 >= 0 {
			ch = word[loc[0]-1 : loc[0]]
		}
		restored = restoreCase(ch, result)
	} else {
		restored = restoreCase(match, result)
	}
	return word[:loc[0]] + restored + word[loc[1]:]
}

// sanitizeWord applies the first matching rule (searched from the end) to word.
func sanitizeWord(token, word string, rules []rule) string {
	if len(token) == 0 || uncountables[token] {
		return word
	}
	for i := len(rules) - 1; i >= 0; i-- {
		if rules[i].re.MatchString(word) {
			return replaceRule(word, rules[i])
		}
	}
	return word
}

func replaceWord(word string, replaceMap, keepMap map[string]string, rules []rule) string {
	token := strings.ToLower(word)
	if _, ok := keepMap[token]; ok {
		return restoreCase(word, token)
	}
	if v, ok := replaceMap[token]; ok {
		return restoreCase(word, v)
	}
	return sanitizeWord(token, word, rules)
}

func checkWord(word string, replaceMap, keepMap map[string]string, rules []rule) bool {
	token := strings.ToLower(word)
	if _, ok := keepMap[token]; ok {
		return true
	}
	if _, ok := replaceMap[token]; ok {
		return false
	}
	return sanitizeWord(token, token, rules) == token
}

// Plural returns the plural form of the given word.
func Plural(word string) string {
	return replaceWord(word, irregularSingles, irregularPlurals, pluralRules)
}

// Singular returns the singular form of the given word.
func Singular(word string) string {
	return replaceWord(word, irregularPlurals, irregularSingles, singularRules)
}

// IsPlural reports whether the given word is already in plural form.
func IsPlural(word string) bool {
	return checkWord(word, irregularSingles, irregularPlurals, pluralRules)
}

// IsSingular reports whether the given word is already in singular form.
func IsSingular(word string) bool {
	return checkWord(word, irregularPlurals, irregularSingles, singularRules)
}

func initIrregulars() {
	pairs := [][2]string{
		{"i", "we"},
		{"me", "us"},
		{"he", "they"},
		{"she", "they"},
		{"them", "them"},
		{"myself", "ourselves"},
		{"yourself", "yourselves"},
		{"itself", "themselves"},
		{"herself", "themselves"},
		{"himself", "themselves"},
		{"themself", "themselves"},
		{"is", "are"},
		{"was", "were"},
		{"has", "have"},
		{"this", "these"},
		{"that", "those"},
		{"my", "our"},
		{"echo", "echoes"},
		{"dingo", "dingoes"},
		{"volcano", "volcanoes"},
		{"tornado", "tornadoes"},
		{"torpedo", "torpedoes"},
		{"genus", "genera"},
		{"viscus", "viscera"},
		{"stigma", "stigmata"},
		{"stoma", "stomata"},
		{"dogma", "dogmata"},
		{"lemma", "lemmata"},
		{"schema", "schemata"},
		{"anathema", "anathemata"},
		{"ox", "oxen"},
		{"axe", "axes"},
		{"die", "dice"},
		{"yes", "yeses"},
		{"foot", "feet"},
		{"eave", "eaves"},
		{"goose", "geese"},
		{"tooth", "teeth"},
		{"quiz", "quizzes"},
		{"human", "humans"},
		{"proof", "proofs"},
		{"carve", "carves"},
		{"valve", "valves"},
		{"looey", "looies"},
		{"thief", "thieves"},
		{"groove", "grooves"},
		{"pickaxe", "pickaxes"},
		{"passerby", "passersby"},
	}
	for _, p := range pairs {
		addIrregular(p[0], p[1])
	}
}

func initPluralRules() {
	pluralRules = []rule{
		mk(`s?$`, `s`),
		mk(`[^\x00-\x7f]$`, `$0`),
		mk(`([^aeiou]ese)$`, `$1`),
		mk(`(ax|test)is$`, `$1es`),
		mk(`(alias|[^aou]us|t[lm]as|gas|ris)$`, `$1es`),
		mk(`(e[mn]u)s?$`, `$1s`),
		mk(`([^l]ias|[aeiou]las|[ejzr]as|[iu]am)$`, `$1`),
		mk(`(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$`, `$1i`),
		mk(`(alumn|alg|vertebr)(?:a|ae)$`, `$1ae`),
		mk(`(seraph|cherub)(?:im)?$`, `$1im`),
		mk(`(her|at|gr)o$`, `$1oes`),
		mk(`(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|automat|quor)(?:a|um)$`, `$1a`),
		mk(`(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)(?:a|on)$`, `$1a`),
		mk(`sis$`, `ses`),
		mk(`(?:(kni|wi|li)fe|(ar|l|ea|eo|oa|hoo)f)$`, `$1$2ves`),
		mk(`([^aeiouy]|qu)y$`, `$1ies`),
		mk(`([^ch][ieo][ln])ey$`, `$1ies`),
		mk(`(x|ch|ss|sh|zz)$`, `$1es`),
		mk(`(matr|cod|mur|sil|vert|ind|append)(?:ix|ex)$`, `$1ices`),
		mk(`\b((?:tit)?m|l)(?:ice|ouse)$`, `$1ice`),
		mk(`(pe)(?:rson|ople)$`, `$1ople`),
		mk(`(child)(?:ren)?$`, `$1ren`),
		mk(`eaux$`, `$0`),
		mk(`m[ae]n$`, `men`),
		mk(`^thou$`, `you`),
	}
}

func initSingularRules() {
	singularRules = []rule{
		mk(`s$`, ``),
		mk(`(ss)$`, `$1`),
		mk(`(wi|kni|(?:after|half|high|low|mid|non|night|[^\w]|^)li)ves$`, `$1fe`),
		mk(`(ar|(?:wo|[ae])l|[eo][ao])ves$`, `$1f`),
		mk(`ies$`, `y`),
		mk(`(dg|ss|ois|lk|ok|wn|mb|th|ch|ec|oal|is|ck|ix|sser|ts|wb)ies$`, `$1ie`),
		mk(`\b(l|(?:neck|cross|hog|aun)?t|coll|faer|food|gen|goon|group|hipp|junk|vegg|(?:pork)?p|charl|calor|cut)ies$`, `$1ie`),
		mk(`\b(mon|smil)ies$`, `$1ey`),
		mk(`\b((?:tit)?m|l)ice$`, `$1ouse`),
		mk(`(seraph|cherub)im$`, `$1`),
		mk(`(x|ch|ss|sh|zz|tto|go|cho|alias|[^aou]us|t[lm]as|gas|(?:her|at|gr)o|[aeiou]ris)(?:es)?$`, `$1`),
		mk(`(analy|diagno|parenthe|progno|synop|the|empha|cri|ne)(?:sis|ses)$`, `$1sis`),
		mk(`(movie|twelve|abuse|e[mn]u)s$`, `$1`),
		mk(`(test)(?:is|es)$`, `$1is`),
		mk(`(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$`, `$1us`),
		mk(`(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|quor)a$`, `$1um`),
		mk(`(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)a$`, `$1on`),
		mk(`(alumn|alg|vertebr)ae$`, `$1a`),
		mk(`(cod|mur|sil|vert|ind)ices$`, `$1ex`),
		mk(`(matr|append)ices$`, `$1ix`),
		mk(`(pe)(rson|ople)$`, `$1rson`),
		mk(`(child)ren$`, `$1`),
		mk(`(eau)x?$`, `$1`),
		mk(`men$`, `man`),
	}
}

func initUncountables() {
	words := []string{
		"adulthood", "advice", "agenda", "aid", "aircraft", "alcohol", "ammo",
		"analytics", "anime", "athletics", "audio", "bison", "blood", "bream",
		"buffalo", "butter", "carp", "cash", "chassis", "chess", "clothing",
		"cod", "commerce", "cooperation", "corps", "debris", "diabetes",
		"digestion", "elk", "energy", "equipment", "excretion", "expertise",
		"firmware", "flounder", "fun", "gallows", "garbage", "graffiti",
		"hardware", "headquarters", "health", "herpes", "highjinks", "homework",
		"housework", "information", "jeans", "justice", "kudos", "labour",
		"literature", "machinery", "mackerel", "mail", "media", "mews", "money",
		"moose", "music", "mud", "manga", "news", "only", "personnel", "pike",
		"plankton", "pliers", "police", "pollution", "premises", "rain",
		"research", "rice", "salmon", "scissors", "series", "sewage", "shambles",
		"shrimp", "software", "species", "staff", "swine", "tennis", "traffic",
		"transportation", "trout", "tuna", "wealth", "welfare", "whiting",
		"wildebeest", "wildlife", "you",
	}
	for _, w := range words {
		addUncountableWord(w)
	}
	patterns := []string{
		`pok[eé]mon$`,
		`[^aeiou]ese$`,
		`deer$`,
		`fish$`,
		`measles$`,
		`o[iu]s$`,
		`pox$`,
		`sheep$`,
	}
	for _, p := range patterns {
		addUncountablePattern(p)
	}
}
