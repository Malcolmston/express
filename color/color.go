// Package color is a standard-library-only Go port of the color primitives that
// underpin the popular npm packages color-convert, tinycolor2 and chroma-js,
// which Express/Node web apps reach for when generating themes, badges and UI
// palettes. It converts between the RGB, HSL and HSV colour models and CSS hex
// notation, and provides the everyday manipulation helpers — lighten, darken,
// saturate, desaturate, mix, invert, grayscale, complement and hue rotation —
// together with WCAG relative-luminance and contrast-ratio calculations for
// accessibility checks.
//
// Colours are 8-bit sRGB. HexToRGB accepts "#rgb", "#rrggbb", "rgb" and
// "rrggbb" (the leading '#' and letter case are optional) and returns an error
// for anything else; RGBToHex always emits lowercase "#rrggbb". The HSL and HSV
// hue is in degrees [0,360); saturation, lightness and value are fractions in
// [0,1]. Conversions round to the nearest 8-bit channel and are stable
// round-trips for the canonical inputs exercised by the tests.
//
// The manipulation helpers take an amount in [0,1]: Lighten(c, 0.2) moves the
// HSL lightness 20 percentage points toward white, Darken toward black,
// Saturate/Desaturate adjust HSL saturation, and Mix blends two colours by a
// weight (0 returns the first colour, 1 the second, 0.5 an even blend).
// Luminance returns the WCAG relative luminance of a colour and ContrastRatio
// returns the WCAG 2.1 contrast ratio between two colours (1 to 21), the basis
// for IsLight/IsDark. Everything is deterministic and depends only on the
// standard library (math, strings, fmt).
package color

import (
	"fmt"
	"math"
	"strings"
)

// RGB is an 8-bit sRGB colour with red, green and blue channels in [0,255].
type RGB struct {
	R uint8
	G uint8
	B uint8
}

// HSL is a colour in the hue-saturation-lightness model. H is in degrees
// [0,360); S and L are fractions in [0,1].
type HSL struct {
	H float64
	S float64
	L float64
}

// HSV is a colour in the hue-saturation-value model. H is in degrees [0,360);
// S and V are fractions in [0,1].
type HSV struct {
	H float64
	S float64
	V float64
}

// HexToRGB parses a CSS hex colour and returns the RGB colour. It accepts the
// three-nibble ("#rgb"), four-nibble ("#rgba"), six-digit ("#rrggbb") and
// eight-digit ("#rrggbbaa") forms, with or without the leading '#'. For the
// alpha-bearing forms the alpha channel is parsed for validation but discarded,
// since RGB has no alpha component. It returns an error for malformed input.
func HexToRGB(hex string) (RGB, error) {
	s := strings.TrimSpace(hex)
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 3, 4:
		// "#rgb" and "#rgba": each nibble is doubled. The optional fourth
		// nibble is alpha, which is validated but not represented in RGB.
		var vals [4]uint8
		ok := true
		for i := 0; i < len(s); i++ {
			n, err := colHexNibble(s[i])
			if err != nil {
				ok = false
				break
			}
			vals[i] = n*16 + n
		}
		if !ok {
			return RGB{}, fmt.Errorf("color: invalid hex %q", hex)
		}
		return RGB{R: vals[0], G: vals[1], B: vals[2]}, nil
	case 6, 8:
		// "#rrggbb" and "#rrggbbaa": the trailing byte pair is alpha, which is
		// validated but not represented in RGB.
		pairs := len(s) / 2
		var vals [4]uint8
		for i := 0; i < pairs; i++ {
			hi, err1 := colHexNibble(s[i*2])
			lo, err2 := colHexNibble(s[i*2+1])
			if err1 != nil || err2 != nil {
				return RGB{}, fmt.Errorf("color: invalid hex %q", hex)
			}
			vals[i] = hi*16 + lo
		}
		return RGB{R: vals[0], G: vals[1], B: vals[2]}, nil
	default:
		return RGB{}, fmt.Errorf("color: invalid hex length %q", hex)
	}
}

// RGBToHex returns the lowercase "#rrggbb" representation of the colour.
func RGBToHex(c RGB) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Hex is the method form of RGBToHex.
func (c RGB) Hex() string { return RGBToHex(c) }

// RGBToHSL converts an RGB colour to HSL.
func RGBToHSL(c RGB) HSL {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	l := (max + min) / 2
	if max == min {
		return HSL{H: 0, S: 0, L: l}
	}
	d := max - min
	var s float64
	if l > 0.5 {
		s = d / (2 - max - min)
	} else {
		s = d / (max + min)
	}
	h := colHue(r, g, b, max, d)
	return HSL{H: h, S: s, L: l}
}

// HSLToRGB converts an HSL colour to RGB.
func HSLToRGB(c HSL) RGB {
	h := colWrapHue(c.H) / 360
	s := colClamp01(c.S)
	l := colClamp01(c.L)
	if s == 0 {
		v := colRound255(l)
		return RGB{R: v, G: v, B: v}
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	r := colHueToRGB(p, q, h+1.0/3.0)
	g := colHueToRGB(p, q, h)
	b := colHueToRGB(p, q, h-1.0/3.0)
	return RGB{R: colRound255(r), G: colRound255(g), B: colRound255(b)}
}

// RGBToHSV converts an RGB colour to HSV.
func RGBToHSV(c RGB) HSV {
	r := float64(c.R) / 255
	g := float64(c.G) / 255
	b := float64(c.B) / 255
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	d := max - min
	var s float64
	if max != 0 {
		s = d / max
	}
	h := 0.0
	if d != 0 {
		h = colHue(r, g, b, max, d)
	}
	return HSV{H: h, S: s, V: max}
}

// HSVToRGB converts an HSV colour to RGB.
func HSVToRGB(c HSV) RGB {
	h := colWrapHue(c.H) / 60
	s := colClamp01(c.S)
	v := colClamp01(c.V)
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))
	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return RGB{R: colRound255(r), G: colRound255(g), B: colRound255(b)}
}

// HexToHSL parses a hex colour and returns it in the HSL model.
func HexToHSL(hex string) (HSL, error) {
	c, err := HexToRGB(hex)
	if err != nil {
		return HSL{}, err
	}
	return RGBToHSL(c), nil
}

// HSLToHex converts an HSL colour to a "#rrggbb" hex string.
func HSLToHex(c HSL) string { return RGBToHex(HSLToRGB(c)) }

// Lighten increases the HSL lightness of a hex colour by amount (a fraction in
// [0,1], clamped) and returns the new hex colour.
func Lighten(hex string, amount float64) (string, error) {
	h, err := HexToHSL(hex)
	if err != nil {
		return "", err
	}
	h.L = colClamp01(h.L + amount)
	return HSLToHex(h), nil
}

// Darken decreases the HSL lightness of a hex colour by amount and returns the
// new hex colour.
func Darken(hex string, amount float64) (string, error) {
	return Lighten(hex, -amount)
}

// Saturate increases the HSL saturation of a hex colour by amount and returns
// the new hex colour.
func Saturate(hex string, amount float64) (string, error) {
	h, err := HexToHSL(hex)
	if err != nil {
		return "", err
	}
	h.S = colClamp01(h.S + amount)
	return HSLToHex(h), nil
}

// Desaturate decreases the HSL saturation of a hex colour by amount and returns
// the new hex colour.
func Desaturate(hex string, amount float64) (string, error) {
	return Saturate(hex, -amount)
}

// Grayscale removes all saturation from a hex colour, returning its grey
// equivalent (preserving lightness).
func Grayscale(hex string) (string, error) {
	return Desaturate(hex, 1)
}

// Invert returns the per-channel inverse (255-c) of a hex colour.
func Invert(hex string) (string, error) {
	c, err := HexToRGB(hex)
	if err != nil {
		return "", err
	}
	return RGBToHex(RGB{R: 255 - c.R, G: 255 - c.G, B: 255 - c.B}), nil
}

// RotateHue rotates the HSL hue of a hex colour by deg degrees (wrapping into
// [0,360)) and returns the new hex colour.
func RotateHue(hex string, deg float64) (string, error) {
	h, err := HexToHSL(hex)
	if err != nil {
		return "", err
	}
	h.H = colWrapHue(h.H + deg)
	return HSLToHex(h), nil
}

// Complement returns the complementary colour (hue rotated 180 degrees).
func Complement(hex string) (string, error) { return RotateHue(hex, 180) }

// Mix blends two hex colours in sRGB space by weight (0 returns c1, 1 returns
// c2, 0.5 an even mix) and returns the resulting hex colour.
func Mix(c1, c2 string, weight float64) (string, error) {
	a, err := HexToRGB(c1)
	if err != nil {
		return "", err
	}
	b, err := HexToRGB(c2)
	if err != nil {
		return "", err
	}
	w := colClamp01(weight)
	mix := func(x, y uint8) uint8 {
		return colRound8(float64(x)*(1-w) + float64(y)*w)
	}
	return RGBToHex(RGB{R: mix(a.R, b.R), G: mix(a.G, b.G), B: mix(a.B, b.B)}), nil
}

// Luminance returns the WCAG 2.1 relative luminance of an RGB colour, a value in
// [0,1] where 0 is black and 1 is white.
func Luminance(c RGB) float64 {
	lin := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// ContrastRatio returns the WCAG 2.1 contrast ratio between two hex colours, a
// value from 1 (identical) to 21 (black on white).
func ContrastRatio(c1, c2 string) (float64, error) {
	a, err := HexToRGB(c1)
	if err != nil {
		return 0, err
	}
	b, err := HexToRGB(c2)
	if err != nil {
		return 0, err
	}
	l1 := Luminance(a)
	l2 := Luminance(b)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05), nil
}

// IsLight reports whether a hex colour has a relative luminance above 0.5,
// meaning dark text reads well on it.
func IsLight(hex string) (bool, error) {
	c, err := HexToRGB(hex)
	if err != nil {
		return false, err
	}
	return Luminance(c) > 0.5, nil
}

// IsDark reports whether a hex colour is not light (see IsLight).
func IsDark(hex string) (bool, error) {
	light, err := IsLight(hex)
	if err != nil {
		return false, err
	}
	return !light, nil
}

// --- helpers ----------------------------------------------------------------

func colHue(r, g, b, max, d float64) float64 {
	var h float64
	switch max {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h * 60
}

func colHueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6
	default:
		return p
	}
}

func colHexNibble(c byte) (uint8, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	default:
		return 0, fmt.Errorf("color: bad hex digit %q", string(c))
	}
}

func colClamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func colWrapHue(h float64) float64 {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}

func colRound255(v float64) uint8 { return colRound8(v * 255) }

func colRound8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(math.Round(v))
}
