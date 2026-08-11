# numberformat — standard-library-only Go port of the number presentation helpers popularised by the npm

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/numberformat.svg)](https://pkg.go.dev/github.com/malcolmston/express/numberformat)

Package numberformat is a standard-library-only Go port of the number
presentation helpers popularised by the npm packages numeral.js and
accounting.js, which Express/Node apps use to render numbers for humans:
thousands grouping, fixed-decimal formatting, currency and percentage
formatting, ordinal suffixes and compact "1.2k / 3.4M" abbreviation.

The workhorse is FormatFloat, which rounds to a fixed number of decimal places
(half away from zero), groups the integer part with a configurable thousands
separator and joins it to the fraction with a configurable decimal separator.
Comma and CommaFloat are the common English-locale shortcuts (comma grouping,
dot decimal). FormatCurrency prefixes a symbol and FormatPercent multiplies by
100 and appends "%". Ordinal returns the English ordinal string for an integer
("1st", "22nd", "113th"), and Abbreviate renders large magnitudes compactly
with k/M/B/T suffixes. RoundTo exposes the underlying decimal rounding.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/numberformat"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Abbreviate(f float64, decimals int) string` | Abbreviate renders f compactly with a magnitude suffix (k, M, B, T) and the given number of decimal places, e.g. |
| `func Comma(n int64) string` | Comma formats an integer with comma thousands grouping (English locale). |
| `func CommaFloat(f float64, decimals int) string` | CommaFloat formats a float with comma thousands grouping, a dot decimal separator and the given number of decimal places. |
| `func FormatCurrency(f float64, symbol string, decimals int) string` | FormatCurrency formats f as a currency amount, prefixing symbol to the comma-grouped, fixed-decimal number. |
| `func FormatFloat(f float64, decimals int, decSep, thousandsSep string) string` | FormatFloat formats f with exactly decimals fractional digits, grouping the integer part with thousandsSep and separating the fraction with decSep. |
| `func FormatInt(n int64, thousandsSep string) string` | FormatInt formats an integer with the given thousands separator, e.g. |
| `func FormatPercent(f float64, decimals int) string` | FormatPercent formats f as a percentage by multiplying by 100 and appending "%", with the given number of decimal places: FormatPercent(0.1234, 2) is… |
| `func Ordinal(n int) string` | Ordinal returns the English ordinal representation of n, e.g. |
| `func RoundTo(f float64, decimals int) float64` | RoundTo rounds f to the given number of decimal places using half-away-from-zero rounding. |
| `func Unformat(s string) (float64, error)` | Unformat parses a human-formatted number string back to a float64, ignoring any grouping separators, currency symbols, percent signs and surrounding… |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/numberformat).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
