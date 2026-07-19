package pluralize

import "testing"

func TestPlural(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// Regular rules.
		{"test", "tests"},
		{"apple", "apples"},
		{"box", "boxes"},
		{"church", "churches"},
		{"bus", "buses"},
		{"lady", "ladies"},
		{"baby", "babies"},
		{"city", "cities"},
		{"day", "days"},
		{"knife", "knives"},
		{"leaf", "leaves"},
		{"wolf", "wolves"},
		{"hero", "heroes"},
		{"potato", "potatoes"},
		// Irregulars.
		{"person", "people"},
		{"man", "men"},
		{"woman", "women"},
		{"child", "children"},
		{"mouse", "mice"},
		{"tooth", "teeth"},
		{"foot", "feet"},
		{"goose", "geese"},
		{"ox", "oxen"},
		{"die", "dice"},
		{"quiz", "quizzes"},
		// Case preservation.
		{"Bus", "Buses"},
		{"BUS", "BUSES"},
		{"Person", "People"},
		// Uncountables.
		{"fish", "fish"},
		{"sheep", "sheep"},
		{"series", "series"},
		{"deer", "deer"},
		{"information", "information"},
	}
	for _, tt := range tests {
		if got := Plural(tt.in); got != tt.want {
			t.Errorf("Plural(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSingular(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		// Regular rules.
		{"tests", "test"},
		{"apples", "apple"},
		{"boxes", "box"},
		{"churches", "church"},
		{"buses", "bus"},
		{"ladies", "lady"},
		{"babies", "baby"},
		{"cities", "city"},
		{"knives", "knife"},
		{"leaves", "leaf"},
		{"wolves", "wolf"},
		{"heroes", "hero"},
		{"potatoes", "potato"},
		// Irregulars.
		{"people", "person"},
		{"men", "man"},
		{"women", "woman"},
		{"children", "child"},
		{"mice", "mouse"},
		{"teeth", "tooth"},
		{"feet", "foot"},
		{"geese", "goose"},
		{"oxen", "ox"},
		{"dice", "die"},
		{"quizzes", "quiz"},
		// Case preservation.
		{"Buses", "Bus"},
		{"People", "Person"},
		// Uncountables.
		{"fish", "fish"},
		{"sheep", "sheep"},
		{"series", "series"},
		{"deer", "deer"},
		{"information", "information"},
	}
	for _, tt := range tests {
		if got := Singular(tt.in); got != tt.want {
			t.Errorf("Singular(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsPlural(t *testing.T) {
	plurals := []string{"tests", "people", "men", "children", "mice", "boxes", "buses", "ladies"}
	for _, w := range plurals {
		if !IsPlural(w) {
			t.Errorf("IsPlural(%q) = false, want true", w)
		}
	}
	singulars := []string{"test", "person", "man", "child", "mouse", "box", "bus", "lady"}
	for _, w := range singulars {
		if IsPlural(w) {
			t.Errorf("IsPlural(%q) = true, want false", w)
		}
	}
}

func TestIsSingular(t *testing.T) {
	singulars := []string{"test", "person", "man", "child", "mouse", "box", "bus", "lady"}
	for _, w := range singulars {
		if !IsSingular(w) {
			t.Errorf("IsSingular(%q) = false, want true", w)
		}
	}
	plurals := []string{"tests", "people", "men", "children", "mice", "boxes", "buses", "ladies"}
	for _, w := range plurals {
		if IsSingular(w) {
			t.Errorf("IsSingular(%q) = true, want false", w)
		}
	}
}

func TestUncountablesBothWays(t *testing.T) {
	words := []string{"fish", "sheep", "series", "deer", "information"}
	for _, w := range words {
		if !IsPlural(w) {
			t.Errorf("IsPlural(%q) = false, want true (uncountable)", w)
		}
		if !IsSingular(w) {
			t.Errorf("IsSingular(%q) = false, want true (uncountable)", w)
		}
	}
}

func TestIdempotent(t *testing.T) {
	// Pluralizing a plural leaves it unchanged.
	if got := Plural("people"); got != "people" {
		t.Errorf("Plural(%q) = %q, want %q", "people", got, "people")
	}
	if got := Singular("person"); got != "person" {
		t.Errorf("Singular(%q) = %q, want %q", "person", got, "person")
	}
}
