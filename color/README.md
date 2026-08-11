# color — standard-library-only Go port of the color primitives that underpin the popular npm

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/color.svg)](https://pkg.go.dev/github.com/malcolmston/express/color)

Package color is a standard-library-only Go port of the color primitives that
underpin the popular npm packages color-convert, tinycolor2 and chroma-js,
which Express/Node web apps reach for when generating themes, badges and UI
palettes. It converts between the RGB, HSL and HSV colour models and CSS hex
notation, and provides the everyday manipulation helpers — lighten, darken,
saturate, desaturate, mix, invert, grayscale, complement and hue rotation —
together with WCAG relative-luminance and contrast-ratio calculations for
accessibility checks.

Colours are 8-bit sRGB. HexToRGB accepts "#rgb", "#rrggbb", "rgb" and "rrggbb"
(the leading '#' and letter case are optional) and returns an error for
anything else; RGBToHex always emits lowercase "#rrggbb". The HSL and HSV hue
is in degrees [0,360); saturation, lightness and value are fractions in [0,1].
Conversions round to the nearest 8-bit channel and are stable round-trips for
the canonical inputs exercised by the tests.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/color"
```

## Usage

This is the package's own `ExampleContrastRatio`, so it compiles and its output is
asserted on every `go test ./color/`.

```go
ratio, err := color.ContrastRatio("#000000", "#ffffff")
	fmt.Printf("%.0f %v\n", ratio, err)

	fg, _ := color.ContrastColor("#ffff00")
	fmt.Println(fg)

	fg, _ = color.ContrastColor("#003366")
	fmt.Println(fg)
```

```
21 <nil>
#000000
#ffffff
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Complement(hex string) (string, error)` | Complement returns the complementary colour (hue rotated 180 degrees). |
| `func ContrastColor(hex string) (string, error)` | ContrastColor returns "#000000" or "#ffffff", whichever has the higher WCAG contrast ratio against the given background hex colour — the standard… |
| `func ContrastRatio(c1, c2 string) (float64, error)` | ContrastRatio returns the WCAG 2.1 contrast ratio between two hex colours, a value from 1 (identical) to 21 (black on white). |
| `func Darken(hex string, amount float64) (string, error)` | Darken decreases the HSL lightness of a hex colour by amount and returns the new hex colour. |
| `func Desaturate(hex string, amount float64) (string, error)` | Desaturate decreases the HSL saturation of a hex colour by amount and returns the new hex colour. |
| `func Grayscale(hex string) (string, error)` | Grayscale removes all saturation from a hex colour, returning its grey equivalent (preserving lightness). |
| `func HSLToHex(c HSL) string` | HSLToHex converts an HSL colour to a "#rrggbb" hex string. |
| `func Invert(hex string) (string, error)` | Invert returns the per-channel inverse (255-c) of a hex colour. |
| `func IsDark(hex string) (bool, error)` | IsDark reports whether a hex colour is not light (see IsLight). |
| `func IsLight(hex string) (bool, error)` | IsLight reports whether a hex colour has a relative luminance above 0.5, meaning dark text reads well on it. |
| `func Lighten(hex string, amount float64) (string, error)` | Lighten increases the HSL lightness of a hex colour by amount (a fraction in [0,1], clamped) and returns the new hex colour. |
| `func Luminance(c RGB) float64` | Luminance returns the WCAG 2.1 relative luminance of an RGB colour, a value in [0,1] where 0 is black and 1 is white. |
| `func Mix(c1, c2 string, weight float64) (string, error)` | Mix blends two hex colours in sRGB space by weight (0 returns c1, 1 returns c2, 0.5 an even mix) and returns the resulting hex colour. |
| `func RGBToHex(c RGB) string` | RGBToHex returns the lowercase "#rrggbb" representation of the colour. |
| `func RotateHue(hex string, deg float64) (string, error)` | RotateHue rotates the HSL hue of a hex colour by deg degrees (wrapping into [0,360)) and returns the new hex colour. |
| `func Saturate(hex string, amount float64) (string, error)` | Saturate increases the HSL saturation of a hex colour by amount and returns the new hex colour. |
| `func Shade(hex string, amount float64) (string, error)` | Shade blends a hex colour toward black by amount (a fraction in [0,1]) and returns the resulting hex colour. |
| `func Tint(hex string, amount float64) (string, error)` | Tint blends a hex colour toward white by amount (a fraction in [0,1]) and returns the resulting hex colour. |

### Types

| Type | What it is |
| --- | --- |
| `HSL` | HSL is a colour in the hue-saturation-lightness model. |
| `HSV` | HSV is a colour in the hue-saturation-value model. |
| `RGB` | RGB is an 8-bit sRGB colour with red, green and blue channels in [0,255]. |

<details>
<summary><code>HSL</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func HexToHSL(hex string) (HSL, error)` | HexToHSL parses a hex colour and returns it in the HSL model. |
| `func RGBToHSL(c RGB) HSL` | RGBToHSL converts an RGB colour to HSL. |

</details>

<details>
<summary><code>HSV</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func RGBToHSV(c RGB) HSV` | RGBToHSV converts an RGB colour to HSV. |

</details>

<details>
<summary><code>RGB</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func HSLToRGB(c HSL) RGB` | HSLToRGB converts an HSL colour to RGB. |
| `func HSVToRGB(c HSV) RGB` | HSVToRGB converts an HSV colour to RGB. |
| `func HexToRGB(hex string) (RGB, error)` | HexToRGB parses a CSS hex colour and returns the RGB colour. |
| `func (c RGB) Hex() string` | Hex is the method form of RGBToHex. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/color).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
