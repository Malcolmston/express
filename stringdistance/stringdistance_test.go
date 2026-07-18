package stringdistance

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-4 }

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"gumbo", "gambol", 2},
		{"café", "cafe", 1},
	}
	for _, tt := range tests {
		if got := Levenshtein(tt.a, tt.b); got != tt.want {
			t.Errorf("Levenshtein(%q,%q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestLevenshteinRatio(t *testing.T) {
	if !approx(LevenshteinRatio("kitten", "sitting"), 1-3.0/7.0) {
		t.Error("LevenshteinRatio")
	}
	if LevenshteinRatio("", "") != 1 {
		t.Error("LevenshteinRatio empty")
	}
}

func TestDamerau(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"ca", "ac", 1},     // one transposition
		{"abcd", "acbd", 1}, // adjacent swap
		{"kitten", "sitting", 3},
	}
	for _, tt := range tests {
		if got := DamerauLevenshtein(tt.a, tt.b); got != tt.want {
			t.Errorf("DamerauLevenshtein(%q,%q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestHamming(t *testing.T) {
	if d, err := Hamming("karolin", "kathrin"); err != nil || d != 3 {
		t.Errorf("Hamming = %d, %v", d, err)
	}
	if _, err := Hamming("abc", "ab"); err == nil {
		t.Error("expected length mismatch error")
	}
}

func TestJaro(t *testing.T) {
	if !approx(JaroSimilarity("MARTHA", "MARHTA"), 0.944444) {
		t.Errorf("Jaro MARTHA/MARHTA = %g", JaroSimilarity("MARTHA", "MARHTA"))
	}
	if JaroSimilarity("abc", "abc") != 1 {
		t.Error("Jaro identical")
	}
	if JaroSimilarity("", "abc") != 0 {
		t.Error("Jaro empty")
	}
}

func TestJaroWinkler(t *testing.T) {
	if !approx(JaroWinkler("MARTHA", "MARHTA"), 0.961111) {
		t.Errorf("JaroWinkler = %g", JaroWinkler("MARTHA", "MARHTA"))
	}
	if !approx(JaroWinkler("DWAYNE", "DUANE"), 0.84) {
		t.Errorf("JaroWinkler DWAYNE/DUANE = %g", JaroWinkler("DWAYNE", "DUANE"))
	}
}

func TestDice(t *testing.T) {
	if DiceCoefficient("night", "night") != 1 {
		t.Error("Dice identical")
	}
	if !approx(DiceCoefficient("night", "nacht"), 0.25) {
		t.Errorf("Dice night/nacht = %g", DiceCoefficient("night", "nacht"))
	}
	if DiceCoefficient("a", "b") != 0 {
		t.Error("Dice short")
	}
}

func TestLCS(t *testing.T) {
	if got := LongestCommonSubsequence("ABCBDAB", "BDCAB"); got != 4 {
		t.Errorf("LCS = %d, want 4", got)
	}
	if got := LongestCommonSubsequence("", "abc"); got != 0 {
		t.Error("LCS empty")
	}
}

func TestClosestMatch(t *testing.T) {
	got, score := ClosestMatch("healed", []string{"sealed", "help", "hulk"})
	if got != "sealed" {
		t.Errorf("ClosestMatch = %q (score %g)", got, score)
	}
	if m, s := ClosestMatch("x", nil); m != "" || s != 0 {
		t.Error("ClosestMatch empty")
	}
}

func BenchmarkLevenshtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Levenshtein("kitten", "sitting")
	}
}
