package numberformat

import "testing"

func TestRoundTo(t *testing.T) {
	tests := []struct {
		f    float64
		d    int
		want float64
	}{
		{2.5, 0, 3},
		{2.345, 2, 2.35},
		{-2.345, 2, -2.35},
		{1234, -2, 1200},
	}
	for _, tt := range tests {
		if got := RoundTo(tt.f, tt.d); got != tt.want {
			t.Errorf("RoundTo(%g,%d) = %g, want %g", tt.f, tt.d, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		f              float64
		d              int
		dec, thousands string
		want           string
	}{
		{1234567.891, 2, ".", ",", "1,234,567.89"},
		{0, 2, ".", ",", "0.00"},
		{-1234.5, 1, ".", ",", "-1,234.5"},
		{1000, 0, ".", " ", "1 000"},
		{12.345, 0, ".", ",", "12"},
	}
	for _, tt := range tests {
		if got := FormatFloat(tt.f, tt.d, tt.dec, tt.thousands); got != tt.want {
			t.Errorf("FormatFloat(%g,%d) = %q, want %q", tt.f, tt.d, got, tt.want)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{1234567, "1,234,567"},
		{-1000, "-1,000"},
		{999, "999"},
		{0, "0"},
	}
	for _, tt := range tests {
		if got := Comma(tt.n); got != tt.want {
			t.Errorf("Comma(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestCommaFloat(t *testing.T) {
	if got := CommaFloat(1234.5, 2); got != "1,234.50" {
		t.Errorf("CommaFloat = %q", got)
	}
}

func TestCurrency(t *testing.T) {
	if got := FormatCurrency(1234.5, "$", 2); got != "$1,234.50" {
		t.Errorf("FormatCurrency = %q", got)
	}
	if got := FormatCurrency(-1234.5, "$", 2); got != "-$1,234.50" {
		t.Errorf("FormatCurrency neg = %q", got)
	}
}

func TestPercent(t *testing.T) {
	if got := FormatPercent(0.1234, 2); got != "12.34%" {
		t.Errorf("FormatPercent = %q", got)
	}
	if got := FormatPercent(1, 0); got != "100%" {
		t.Errorf("FormatPercent 1 = %q", got)
	}
}

func TestOrdinal(t *testing.T) {
	tests := map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th", 21: "21st", 22: "22nd", 113: "113th", 101: "101st"}
	for in, want := range tests {
		if got := Ordinal(in); got != want {
			t.Errorf("Ordinal(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestUnformat(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"$1,234.56", 1234.56},
		{"1,000", 1000},
		{"-$2,500.00", -2500},
		{"12.34%", 12.34},
		{"42", 42},
	}
	for _, tt := range tests {
		got, err := Unformat(tt.in)
		if err != nil {
			t.Fatalf("Unformat(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("Unformat(%q) = %g, want %g", tt.in, got, tt.want)
		}
	}
	if _, err := Unformat("no digits"); err == nil {
		t.Error("expected error")
	}
}

func TestAbbreviate(t *testing.T) {
	tests := []struct {
		f    float64
		d    int
		want string
	}{
		{1500, 1, "1.5k"},
		{2300000, 2, "2.30M"},
		{1234567890, 1, "1.2B"},
		{5, 0, "5"},
		{-1500, 1, "-1.5k"},
	}
	for _, tt := range tests {
		if got := Abbreviate(tt.f, tt.d); got != tt.want {
			t.Errorf("Abbreviate(%g,%d) = %q, want %q", tt.f, tt.d, got, tt.want)
		}
	}
}

func BenchmarkFormatFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatFloat(1234567.891, 2, ".", ",")
	}
}
