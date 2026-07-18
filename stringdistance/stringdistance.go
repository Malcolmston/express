// Package stringdistance is a standard-library-only Go collection of the string
// similarity and edit-distance algorithms distributed on npm as leven,
// fast-levenshtein, string-similarity, natural and talisman, which Express/Node
// apps use for fuzzy matching, spell correction and deduplication. It provides
// Levenshtein and Damerau-Levenshtein edit distances, Hamming distance, the
// Jaro and Jaro-Winkler similarities, the Sorensen-Dice bigram coefficient, the
// length of the longest common subsequence, and a ClosestMatch convenience that
// ranks candidates by Dice similarity.
//
// Distances count edits (insertions, deletions, substitutions, and for
// Damerau-Levenshtein adjacent transpositions) and are therefore non-negative
// integers; LevenshteinRatio normalises Levenshtein to a [0,1] similarity.
// Similarities (Jaro, JaroWinkler, DiceCoefficient) are floats in [0,1] where 1
// means identical. All functions compare by Unicode code point (runes), so
// multibyte text is handled correctly, and Hamming returns an error when the
// inputs differ in length. Everything is deterministic and depends only on the
// standard library.
package stringdistance

import (
	"errors"
	"unicode/utf8"
)

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Levenshtein returns the Levenshtein edit distance between a and b: the minimum
// number of single-character insertions, deletions or substitutions needed to
// transform a into b.
func Levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = minInt(minInt(curr[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

// LevenshteinRatio returns a similarity in [0,1] derived from the Levenshtein
// distance: 1 - distance/max(len(a), len(b)). Two empty strings are considered
// identical (ratio 1).
func LevenshteinRatio(a, b string) float64 {
	maxLen := maxInt(utf8.RuneCountInString(a), utf8.RuneCountInString(b))
	if maxLen == 0 {
		return 1
	}
	return 1 - float64(Levenshtein(a, b))/float64(maxLen)
}

// DamerauLevenshtein returns the optimal string alignment (restricted
// Damerau-Levenshtein) distance between a and b, which additionally counts a
// swap of two adjacent characters as a single edit.
func DamerauLevenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = minInt(minInt(d[i-1][j]+1, d[i][j-1]+1), d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				d[i][j] = minInt(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[la][lb]
}

// ErrLengthMismatch is returned by Hamming when the inputs differ in length.
var ErrLengthMismatch = errors.New("stringdistance: strings must be of equal length")

// Hamming returns the Hamming distance between two equal-length strings: the
// number of positions at which the corresponding runes differ. It returns
// ErrLengthMismatch when the lengths differ.
func Hamming(a, b string) (int, error) {
	ra, rb := []rune(a), []rune(b)
	if len(ra) != len(rb) {
		return 0, ErrLengthMismatch
	}
	dist := 0
	for i := range ra {
		if ra[i] != rb[i] {
			dist++
		}
	}
	return dist, nil
}

// JaroSimilarity returns the Jaro similarity of a and b, a value in [0,1] based
// on the number of matching characters and transpositions.
func JaroSimilarity(a, b string) float64 {
	if a == b {
		return 1
	}
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 || lb == 0 {
		return 0
	}
	matchDist := maxInt(la, lb)/2 - 1
	if matchDist < 0 {
		matchDist = 0
	}
	aMatch := make([]bool, la)
	bMatch := make([]bool, lb)
	matches := 0
	for i := 0; i < la; i++ {
		start := maxInt(0, i-matchDist)
		end := minInt(i+matchDist+1, lb)
		for j := start; j < end; j++ {
			if bMatch[j] || ra[i] != rb[j] {
				continue
			}
			aMatch[i] = true
			bMatch[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}
	transpositions := 0
	k := 0
	for i := 0; i < la; i++ {
		if !aMatch[i] {
			continue
		}
		for !bMatch[k] {
			k++
		}
		if ra[i] != rb[k] {
			transpositions++
		}
		k++
	}
	t := float64(transpositions) / 2
	m := float64(matches)
	return (m/float64(la) + m/float64(lb) + (m-t)/m) / 3
}

// JaroWinkler returns the Jaro-Winkler similarity of a and b, which boosts the
// Jaro score for strings sharing a common prefix (up to four characters, scaling
// factor 0.1), applied only when the Jaro similarity exceeds 0.7.
func JaroWinkler(a, b string) float64 {
	jaro := JaroSimilarity(a, b)
	if jaro <= 0.7 {
		return jaro
	}
	ra, rb := []rune(a), []rune(b)
	prefix := 0
	for prefix < 4 && prefix < len(ra) && prefix < len(rb) && ra[prefix] == rb[prefix] {
		prefix++
	}
	return jaro + float64(prefix)*0.1*(1-jaro)
}

// DiceCoefficient returns the Sorensen-Dice coefficient of a and b based on
// shared adjacent character bigrams, a value in [0,1]. Identical strings score
// 1; strings shorter than two characters score 1 when equal and 0 otherwise.
func DiceCoefficient(a, b string) float64 {
	if a == b {
		return 1
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) < 2 || len(rb) < 2 {
		return 0
	}
	bigrams := make(map[string]int)
	for i := 0; i < len(ra)-1; i++ {
		bigrams[string(ra[i:i+2])]++
	}
	intersection := 0
	for i := 0; i < len(rb)-1; i++ {
		key := string(rb[i : i+2])
		if bigrams[key] > 0 {
			bigrams[key]--
			intersection++
		}
	}
	return 2 * float64(intersection) / float64(len(ra)-1+len(rb)-1)
}

// LongestCommonSubsequence returns the length of the longest subsequence common
// to a and b (characters in the same relative order, not necessarily adjacent).
func LongestCommonSubsequence(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 || lb == 0 {
		return 0
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if ra[i-1] == rb[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = maxInt(prev[j], curr[j-1])
			}
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// ClosestMatch returns the candidate most similar to target by Dice coefficient,
// along with its similarity score. It returns ("", 0) when candidates is empty.
func ClosestMatch(target string, candidates []string) (string, float64) {
	best := ""
	bestScore := -1.0
	for _, c := range candidates {
		s := DiceCoefficient(target, c)
		if s > bestScore {
			bestScore = s
			best = c
		}
	}
	if bestScore < 0 {
		return "", 0
	}
	return best, bestScore
}
