# stringdistance — standard-library-only Go collection of the string similarity and edit-distance

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/stringdistance.svg)](https://pkg.go.dev/github.com/malcolmston/express/stringdistance)

Package stringdistance is a standard-library-only Go collection of the string
similarity and edit-distance algorithms distributed on npm as leven,
fast-levenshtein, string-similarity, natural and talisman, which Express/Node
apps use for fuzzy matching, spell correction and deduplication. It provides
Levenshtein and Damerau-Levenshtein edit distances, Hamming distance, the Jaro
and Jaro-Winkler similarities, the Sorensen-Dice bigram coefficient, the
length of the longest common subsequence, the longest common substring, and
the ClosestMatch/FindBestMatch conveniences that rank candidates by Dice
similarity (or, with ClosestMatchLevenshtein, by edit distance).

Distances count edits (insertions, deletions, substitutions, and for
Damerau-Levenshtein adjacent transpositions) and are therefore non-negative
integers; LevenshteinRatio normalises Levenshtein to a [0,1] similarity.
Similarities (Jaro, JaroWinkler, DiceCoefficient) are floats in [0,1] where 1
means identical. All functions compare by Unicode code point (runes), so
multibyte text is handled correctly, and Hamming returns an error when the
inputs differ in length. Everything is deterministic and depends only on the
standard library.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/stringdistance"
```

## Usage

This is the package's own `ExampleClosestMatch`, so it compiles and its output is
asserted on every `go test ./stringdistance/`.

```go
name, score := stringdistance.ClosestMatch("healthz", []string{"/health", "/metrics", "/readyz"})
	fmt.Printf("%s %.4f\n", name, score)
```

```
/health 0.8333
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func ClosestMatch(target string, candidates []string) (string, float64)` | ClosestMatch returns the candidate most similar to target by Dice coefficient, along with its similarity score. |
| `func ClosestMatchLevenshtein(target string, candidates []string) (string, int)` | ClosestMatchLevenshtein returns the candidate closest to target by Levenshtein edit distance, along with that distance. |
| `func DamerauLevenshtein(a, b string) int` | DamerauLevenshtein returns the optimal string alignment (restricted Damerau-Levenshtein) distance between a and b, which additionally counts a swap… |
| `func DiceCoefficient(a, b string) float64` | DiceCoefficient returns the Sorensen-Dice coefficient of a and b based on shared adjacent character bigrams, a value in [0,1]. |
| `func Hamming(a, b string) (int, error)` | Hamming returns the Hamming distance between two equal-length strings: the number of positions at which the corresponding runes differ. |
| `func JaroSimilarity(a, b string) float64` | JaroSimilarity returns the Jaro similarity of a and b, a value in [0,1] based on the number of matching characters and transpositions. |
| `func JaroWinkler(a, b string) float64` | JaroWinkler returns the Jaro-Winkler similarity of a and b, which boosts the Jaro score for strings sharing a common prefix (up to four characters,… |
| `func Levenshtein(a, b string) int` | Levenshtein returns the Levenshtein edit distance between a and b: the minimum number of single-character insertions, deletions or substitutions… |
| `func LevenshteinRatio(a, b string) float64` | LevenshteinRatio returns a similarity in [0,1] derived from the Levenshtein distance: 1 - distance/max(len(a), len(b)). |
| `func LongestCommonSubsequence(a, b string) int` | LongestCommonSubsequence returns the length of the longest subsequence common to a and b (characters in the same relative order, not necessarily… |
| `func LongestCommonSubstring(a, b string) string` | LongestCommonSubstring returns the longest run of characters occurring contiguously in both a and b. |

### Types

| Type | What it is |
| --- | --- |
| `Match` | Match pairs a candidate string with its similarity to a target. |

<details>
<summary><code>Match</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func FindBestMatch(target string, candidates []string) (ratings []Match, bestIndex int)` | FindBestMatch scores every candidate against target with DiceCoefficient and returns the scores in candidate order together with the index of the… |

</details>

### Variables

`ErrLengthMismatch`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/stringdistance).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
