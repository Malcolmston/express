package shortid

import "testing"

func TestGenerateNonEmptyAndValidChars(t *testing.T) {
	for i := 0; i < 100; i++ {
		id, err := Generate()
		if err != nil {
			t.Fatal(err)
		}
		if id == "" {
			t.Fatal("Generate produced empty id")
		}
		if !IsValid(id) {
			t.Fatalf("generated id %q is not valid", id)
		}
	}
}

func TestGenerateLengthRange(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	n := len([]rune(id))
	if n < 7 || n > 14 {
		t.Fatalf("id length %d out of expected range 7..14 (%q)", n, id)
	}
}

func TestUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 5000; i++ {
		id, err := Generate()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = true
	}
}

func TestIsValid(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !IsValid(id) {
		t.Fatalf("IsValid(%q) = false, want true", id)
	}
	if IsValid("abc!def") {
		t.Fatal("IsValid returned true for string with char outside alphabet")
	}
	if IsValid("") {
		t.Fatal("IsValid returned true for empty string")
	}
}

func TestSetAlphabetRejectsWrongLength(t *testing.T) {
	if err := SetAlphabet("tooshort"); err == nil {
		t.Fatal("expected error for wrong-length alphabet")
	}
	// 64 chars but with a duplicate.
	dup := "AABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"
	if len([]rune(dup)) != 64 {
		t.Fatalf("test setup: dup length %d", len([]rune(dup)))
	}
	if err := SetAlphabet(dup); err == nil {
		t.Fatal("expected error for duplicate characters")
	}
}
