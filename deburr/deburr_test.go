package deburr

import "testing"

func TestDeburr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain ascii", "hello world", "hello world"},
		{"deja precomposed", "déjà", "deja"},
		{"creme brulee", "Crème Brûlée", "Creme Brulee"},
		{"grave A", "À", "A"},
		{"e acute", "é", "e"},
		{"n tilde", "ñ", "n"},
		{"u umlaut", "ü", "u"},
		{"c cedilla", "ç", "c"},
		{"o slash", "ø", "o"},
		{"ae ligature lower", "æ", "ae"},
		{"AE ligature upper", "Æ", "Ae"},
		{"sharp s", "ß", "ss"},
		{"oe ligature", "œ", "oe"},
		{"OE ligature", "Œ", "Oe"},
		{"ij ligature", "ĳ", "ij"},
		{"thorn upper", "Þ", "Th"},
		{"thorn lower", "þ", "th"},
		{"eth", "ð", "d"},
		{"polish l", "Łódź", "Lodz"},
		{"vowels mixed", "àáâãäåèéêëìíîïòóôõöùúûý", "aaaaaaeeeeiiiiooooouuuy"},
		{"caps vowels", "ÀÁÂÃÄÅ", "AAAAAA"},
		{"extended a", "ĀăĄćĈ", "AaAcC"},
		{"combining marks decomposed", "é", "e"},
		{"combining a with ring", "å", "a"},
		{"nonlatin untouched", "日本語", "日本語"},
		{"mixed sentence", "Zoë García São Paulo", "Zoe Garcia Sao Paulo"},
		{"apostrophe n", "ŉ", "'n"},
		{"long s", "ſ", "s"},
		{"decomposed e acute", "é", "e"},
		{"decomposed a ring", "å", "a"},
		{"decomposed o tilde word", "São Paulo", "Sao Paulo"},
		{"y with dieresis", "ÿ", "y"},
		{"digits and punctuation preserved", "café-123!", "cafe-123!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Deburr(tt.in); got != tt.want {
				t.Errorf("Deburr(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
