package color

import (
	"math"
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		in      string
		r, g, b uint8
	}{
		{"#ffffff", 255, 255, 255},
		{"000000", 0, 0, 0},
		{"#ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#fff", 255, 255, 255},
		{"#abc", 0xaa, 0xbb, 0xcc},
		{"#336699", 0x33, 0x66, 0x99},
	}
	for _, tt := range tests {
		c, err := HexToRGB(tt.in)
		if err != nil {
			t.Fatalf("HexToRGB(%q): %v", tt.in, err)
		}
		if c.R != tt.r || c.G != tt.g || c.B != tt.b {
			t.Errorf("HexToRGB(%q) = %v, want %d,%d,%d", tt.in, c, tt.r, tt.g, tt.b)
		}
	}
	for _, bad := range []string{"", "#12", "#gggggg", "12345", "#1234567"} {
		if _, err := HexToRGB(bad); err == nil {
			t.Errorf("HexToRGB(%q) expected error", bad)
		}
	}
}

func TestRGBToHex(t *testing.T) {
	if got := RGBToHex(RGB{0x33, 0x66, 0x99}); got != "#336699" {
		t.Errorf("RGBToHex = %q", got)
	}
	if got := (RGB{255, 0, 0}).Hex(); got != "#ff0000" {
		t.Errorf("Hex = %q", got)
	}
}

func approx(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

func TestRGBToHSL(t *testing.T) {
	tests := []struct {
		c       RGB
		h, s, l float64
	}{
		{RGB{255, 0, 0}, 0, 1, 0.5},
		{RGB{0, 255, 0}, 120, 1, 0.5},
		{RGB{0, 0, 255}, 240, 1, 0.5},
		{RGB{255, 255, 255}, 0, 0, 1},
		{RGB{0, 0, 0}, 0, 0, 0},
		{RGB{128, 128, 128}, 0, 0, 0.502},
	}
	for _, tt := range tests {
		h := RGBToHSL(tt.c)
		if !approx(h.H, tt.h, 0.5) || !approx(h.S, tt.s, 0.01) || !approx(h.L, tt.l, 0.01) {
			t.Errorf("RGBToHSL(%v) = %+v, want %g,%g,%g", tt.c, h, tt.h, tt.s, tt.l)
		}
	}
}

func TestHSLRoundTrip(t *testing.T) {
	for _, c := range []RGB{{255, 0, 0}, {0, 255, 0}, {0, 0, 255}, {123, 45, 67}, {200, 100, 50}} {
		got := HSLToRGB(RGBToHSL(c))
		if got != c {
			t.Errorf("HSL round trip %v -> %v", c, got)
		}
	}
}

func TestHSVRoundTrip(t *testing.T) {
	for _, c := range []RGB{{255, 0, 0}, {0, 255, 0}, {0, 0, 255}, {123, 45, 67}, {10, 200, 150}} {
		got := HSVToRGB(RGBToHSV(c))
		if got != c {
			t.Errorf("HSV round trip %v -> %v", c, got)
		}
	}
}

func TestLightenDarken(t *testing.T) {
	// #ff0000 is L=0.5; lighten by 0.5 -> white; darken by 0.5 -> black.
	if got, _ := Lighten("#ff0000", 0.5); got != "#ffffff" {
		t.Errorf("Lighten = %q", got)
	}
	if got, _ := Darken("#ff0000", 0.5); got != "#000000" {
		t.Errorf("Darken = %q", got)
	}
}

func TestGrayscaleInvert(t *testing.T) {
	if got, _ := Grayscale("#ff0000"); got != "#808080" {
		t.Errorf("Grayscale = %q", got)
	}
	if got, _ := Invert("#000000"); got != "#ffffff" {
		t.Errorf("Invert = %q", got)
	}
	if got, _ := Invert("#ff0000"); got != "#00ffff" {
		t.Errorf("Invert = %q", got)
	}
}

func TestComplementRotate(t *testing.T) {
	// complement of red (hue 0) is cyan (hue 180)
	if got, _ := Complement("#ff0000"); got != "#00ffff" {
		t.Errorf("Complement = %q", got)
	}
	if got, _ := RotateHue("#ff0000", 120); got != "#00ff00" {
		t.Errorf("RotateHue = %q", got)
	}
}

func TestMix(t *testing.T) {
	if got, _ := Mix("#000000", "#ffffff", 0.5); got != "#808080" {
		t.Errorf("Mix = %q", got)
	}
	if got, _ := Mix("#000000", "#ffffff", 0); got != "#000000" {
		t.Errorf("Mix 0 = %q", got)
	}
	if got, _ := Mix("#000000", "#ffffff", 1); got != "#ffffff" {
		t.Errorf("Mix 1 = %q", got)
	}
}

func TestLuminanceContrast(t *testing.T) {
	if l := Luminance(RGB{255, 255, 255}); !approx(l, 1, 0.001) {
		t.Errorf("white luminance = %g", l)
	}
	if l := Luminance(RGB{0, 0, 0}); !approx(l, 0, 0.001) {
		t.Errorf("black luminance = %g", l)
	}
	r, err := ContrastRatio("#000000", "#ffffff")
	if err != nil {
		t.Fatal(err)
	}
	if !approx(r, 21, 0.01) {
		t.Errorf("contrast = %g, want 21", r)
	}
	if r2, _ := ContrastRatio("#ffffff", "#ffffff"); !approx(r2, 1, 0.001) {
		t.Errorf("same contrast = %g", r2)
	}
}

func TestIsLightDark(t *testing.T) {
	if ok, _ := IsLight("#ffffff"); !ok {
		t.Error("white should be light")
	}
	if ok, _ := IsDark("#000000"); !ok {
		t.Error("black should be dark")
	}
}

func BenchmarkRGBToHSL(b *testing.B) {
	c := RGB{123, 45, 67}
	for i := 0; i < b.N; i++ {
		_ = RGBToHSL(c)
	}
}
