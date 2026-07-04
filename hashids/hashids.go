// Package hashids is a standard-library port of the Hashids algorithm
// (https://hashids.org). It encodes slices of non-negative integers into short,
// reversible, non-sequential strings using a salt, a configurable minimum
// length and a custom alphabet.
package hashids

import (
	"errors"
	"math"
)

const (
	// DefaultAlphabet is the alphabet used when none is supplied.
	DefaultAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	defaultSeps = "cfhistuCFHISTU"
	sepDiv      = 3.5
	guardDiv    = 12.0
	minAlphabet = 16
)

// HashID encodes and decodes integer slices.
type HashID struct {
	salt      []rune
	minLength int
	alphabet  []rune
	seps      []rune
	guards    []rune
}

// New returns a HashID using the default alphabet.
func New(salt string, minLength int) (*HashID, error) {
	return NewWithAlphabet(salt, minLength, DefaultAlphabet)
}

// NewWithAlphabet returns a HashID using a custom alphabet.
func NewWithAlphabet(salt string, minLength int, alphabet string) (*HashID, error) {
	if minLength < 0 {
		minLength = 0
	}
	saltRunes := []rune(salt)

	// Deduplicate the alphabet, preserving order.
	seen := make(map[rune]bool)
	uniq := make([]rune, 0, len(alphabet))
	for _, r := range alphabet {
		if r == ' ' {
			return nil, errors.New("hashids: alphabet must not contain spaces")
		}
		if !seen[r] {
			seen[r] = true
			uniq = append(uniq, r)
		}
	}
	if len(uniq) < minAlphabet {
		return nil, errors.New("hashids: alphabet must contain at least 16 unique characters")
	}
	alph := uniq

	// Separators are the default seps intersected with the alphabet.
	seps := []rune(defaultSeps)
	filtered := seps[:0:0]
	for _, s := range seps {
		if runeIndex(alph, s) != -1 {
			filtered = append(filtered, s)
		}
	}
	seps = filtered
	alph = removeRunes(alph, seps)

	seps = consistentShuffle(seps, saltRunes)

	if len(seps) == 0 || float64(len(alph))/float64(len(seps)) > sepDiv {
		sepsLength := int(math.Ceil(float64(len(alph)) / sepDiv))
		if sepsLength == 1 {
			sepsLength = 2
		}
		if sepsLength > len(seps) {
			diff := sepsLength - len(seps)
			seps = append(seps, alph[:diff]...)
			alph = alph[diff:]
		} else {
			seps = seps[:sepsLength]
		}
	}

	alph = consistentShuffle(alph, saltRunes)
	guardCount := int(math.Ceil(float64(len(alph)) / guardDiv))

	var guards []rune
	if len(alph) < 3 {
		guards = cloneRunes(seps[:guardCount])
		seps = seps[guardCount:]
	} else {
		guards = cloneRunes(alph[:guardCount])
		alph = alph[guardCount:]
	}

	return &HashID{
		salt:      saltRunes,
		minLength: minLength,
		alphabet:  cloneRunes(alph),
		seps:      cloneRunes(seps),
		guards:    guards,
	}, nil
}

// Encode encodes the given non-negative integers.
func (h *HashID) Encode(nums ...int64) (string, error) {
	if len(nums) == 0 {
		return "", nil
	}
	for _, n := range nums {
		if n < 0 {
			return "", errors.New("hashids: negative numbers are not supported")
		}
	}

	alphabet := cloneRunes(h.alphabet)

	var numbersIDInt int64
	for i, n := range nums {
		numbersIDInt += n % int64(i+100)
	}

	ret := make([]rune, 0, len(nums)*4)
	lottery := alphabet[numbersIDInt%int64(len(alphabet))]
	ret = append(ret, lottery)

	for i, num := range nums {
		buffer := make([]rune, 0, 1+len(h.salt)+len(alphabet))
		buffer = append(buffer, lottery)
		buffer = append(buffer, h.salt...)
		buffer = append(buffer, alphabet...)
		alphabet = consistentShuffle(alphabet, buffer[:len(alphabet)])

		last := hashNum(num, alphabet)
		ret = append(ret, last...)

		if i+1 < len(nums) {
			num %= int64(last[0]) + int64(i)
			sepsIndex := num % int64(len(h.seps))
			ret = append(ret, h.seps[sepsIndex])
		}
	}

	if len(ret) < h.minLength {
		guardIndex := (numbersIDInt + int64(ret[0])) % int64(len(h.guards))
		guard := h.guards[guardIndex]
		ret = append([]rune{guard}, ret...)

		if len(ret) < h.minLength {
			guardIndex = (numbersIDInt + int64(ret[2])) % int64(len(h.guards))
			guard = h.guards[guardIndex]
			ret = append(ret, guard)
		}
	}

	halfLength := len(alphabet) / 2
	for len(ret) < h.minLength {
		alphabet = consistentShuffle(alphabet, alphabet)
		newRet := make([]rune, 0, len(alphabet)+len(ret))
		newRet = append(newRet, alphabet[halfLength:]...)
		newRet = append(newRet, ret...)
		newRet = append(newRet, alphabet[:halfLength]...)
		ret = newRet

		excess := len(ret) - h.minLength
		if excess > 0 {
			start := excess / 2
			ret = ret[start : start+h.minLength]
		}
	}

	return string(ret), nil
}

// Decode decodes a hash back into the integers it encodes. It returns an empty
// slice for input that does not decode consistently.
func (h *HashID) Decode(hash string) ([]int64, error) {
	ret := []int64{}
	if len(hash) == 0 {
		return ret, nil
	}
	hashRunes := []rune(hash)

	alphabet := cloneRunes(h.alphabet)

	// Split on guard characters.
	guardSet := runeSet(h.guards)
	breakdown := make([]rune, len(hashRunes))
	for i, r := range hashRunes {
		if guardSet[r] {
			breakdown[i] = ' '
		} else {
			breakdown[i] = r
		}
	}
	hashArray := splitRunes(breakdown, ' ')

	i := 0
	if len(hashArray) == 3 || len(hashArray) == 2 {
		i = 1
	}
	part := hashArray[i]
	if len(part) == 0 {
		return ret, nil
	}

	lottery := part[0]
	part = part[1:]

	// Split on separator characters.
	sepSet := runeSet(h.seps)
	bd := make([]rune, len(part))
	for k, r := range part {
		if sepSet[r] {
			bd[k] = ' '
		} else {
			bd[k] = r
		}
	}
	subArray := splitRunes(bd, ' ')

	for _, sub := range subArray {
		buffer := make([]rune, 0, 1+len(h.salt)+len(alphabet))
		buffer = append(buffer, lottery)
		buffer = append(buffer, h.salt...)
		buffer = append(buffer, alphabet...)
		alphabet = consistentShuffle(alphabet, buffer[:len(alphabet)])
		ret = append(ret, unhash(sub, alphabet))
	}

	// Validate by re-encoding.
	check, err := h.Encode(ret...)
	if err != nil || check != hash {
		return []int64{}, nil
	}
	return ret, nil
}

// consistentShuffle deterministically shuffles alphabet based on salt.
func consistentShuffle(alphabet, salt []rune) []rune {
	if len(salt) == 0 {
		return cloneRunes(alphabet)
	}
	result := cloneRunes(alphabet)
	for i, v, p := len(result)-1, 0, 0; i > 0; i, v = i-1, v+1 {
		v %= len(salt)
		intVal := int(salt[v])
		p += intVal
		j := (intVal + v + p) % i
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// hashNum converts input to its representation in the given alphabet base.
func hashNum(input int64, alphabet []rune) []rune {
	n := int64(len(alphabet))
	var result []rune
	for {
		result = append([]rune{alphabet[input%n]}, result...)
		input /= n
		if input == 0 {
			break
		}
	}
	return result
}

// unhash converts an alphabet-encoded string back to its integer value.
func unhash(input, alphabet []rune) int64 {
	var number int64
	n := len(alphabet)
	for i := 0; i < len(input); i++ {
		pos := runeIndex(alphabet, input[i])
		number += int64(pos) * intPow(int64(n), len(input)-i-1)
	}
	return number
}

func intPow(base int64, exp int) int64 {
	r := int64(1)
	for i := 0; i < exp; i++ {
		r *= base
	}
	return r
}

func runeIndex(s []rune, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}

func removeRunes(s, toRemove []rune) []rune {
	rem := runeSet(toRemove)
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if !rem[r] {
			out = append(out, r)
		}
	}
	return out
}

func runeSet(s []rune) map[rune]bool {
	m := make(map[rune]bool, len(s))
	for _, r := range s {
		m[r] = true
	}
	return m
}

func cloneRunes(s []rune) []rune {
	out := make([]rune, len(s))
	copy(out, s)
	return out
}

// splitRunes splits s on sep, preserving empty tokens (matching JS String.split).
func splitRunes(s []rune, sep rune) [][]rune {
	result := [][]rune{}
	cur := []rune{}
	for _, r := range s {
		if r == sep {
			result = append(result, cur)
			cur = []rune{}
		} else {
			cur = append(cur, r)
		}
	}
	result = append(result, cur)
	return result
}
