package color

// This file extends the color package with three common theming helpers built
// on the base conversions: ContrastColor picks black or white for legible text
// on a background, and Tint and Shade blend a colour toward white or black by a
// given amount. All operate on and return "#rrggbb" hex strings and are
// deterministic.

// ContrastColor returns "#000000" or "#ffffff", whichever has the higher WCAG
// contrast ratio against the given background hex colour — the standard way to
// choose legible foreground text. It returns an error if hex is malformed.
func ContrastColor(hex string) (string, error) {
	c, err := HexToRGB(hex)
	if err != nil {
		return "", err
	}
	if Luminance(c) > 0.179 {
		return "#000000", nil
	}
	return "#ffffff", nil
}

// Tint blends a hex colour toward white by amount (a fraction in [0,1]) and
// returns the resulting hex colour.
func Tint(hex string, amount float64) (string, error) {
	return Mix(hex, "#ffffff", amount)
}

// Shade blends a hex colour toward black by amount (a fraction in [0,1]) and
// returns the resulting hex colour.
func Shade(hex string, amount float64) (string, error) {
	return Mix(hex, "#000000", amount)
}
