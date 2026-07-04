package contentdisposition

import "testing"

func TestFormatASCII(t *testing.T) {
	got := Format("file.txt")
	want := `attachment; filename="file.txt"`
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

func TestFormatType(t *testing.T) {
	got := Format("file.txt", WithType("inline"))
	want := `inline; filename="file.txt"`
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

func TestFormatNoFilename(t *testing.T) {
	got := Format("")
	want := "attachment"
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

func TestFormatUTF8(t *testing.T) {
	got := Format("€ rates.txt")
	want := `attachment; filename="??? rates.txt"; filename*=UTF-8''%e2%82%ac%20rates.txt`
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

func TestFormatQuoteEscaping(t *testing.T) {
	got := Format(`the "best".txt`)
	want := `attachment; filename="the \"best\".txt"`
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

func TestParsePlain(t *testing.T) {
	cd, err := Parse(`attachment; filename="file.txt"`)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Type != "attachment" {
		t.Errorf("Type = %q", cd.Type)
	}
	if cd.Filename != "file.txt" {
		t.Errorf("Filename = %q", cd.Filename)
	}
}

func TestParseExtPrecedence(t *testing.T) {
	cd, err := Parse(`attachment; filename="EURO rates.txt"; filename*=UTF-8''%e2%82%ac%20rates.txt`)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Filename != "€ rates.txt" {
		t.Fatalf("Filename = %q, want %q", cd.Filename, "€ rates.txt")
	}
}

func TestParseParameters(t *testing.T) {
	cd, err := Parse(`attachment; filename="a.txt"; creation-date="today"`)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Parameters["filename"] != "a.txt" {
		t.Errorf("filename param = %q", cd.Parameters["filename"])
	}
	if cd.Parameters["creation-date"] != "today" {
		t.Errorf("creation-date param = %q", cd.Parameters["creation-date"])
	}
}

func TestParseInline(t *testing.T) {
	cd, err := Parse("inline")
	if err != nil {
		t.Fatal(err)
	}
	if cd.Type != "inline" {
		t.Errorf("Type = %q", cd.Type)
	}
	if cd.Filename != "" {
		t.Errorf("Filename = %q, want empty", cd.Filename)
	}
}

func TestRoundTripUTF8(t *testing.T) {
	orig := "€ rates.txt"
	header := Format(orig)
	cd, err := Parse(header)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Filename != orig {
		t.Fatalf("round-trip Filename = %q, want %q", cd.Filename, orig)
	}
}

func TestRoundTripASCII(t *testing.T) {
	orig := "report.pdf"
	header := Format(orig, WithType("inline"))
	cd, err := Parse(header)
	if err != nil {
		t.Fatal(err)
	}
	if cd.Type != "inline" || cd.Filename != orig {
		t.Fatalf("round-trip = %+v", cd)
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := Parse("   "); err == nil {
		t.Fatal("expected error for empty header")
	}
}
