// Package filesize converts a number of bytes into a human readable string.
//
// It is a faithful port of the npm package "filesize". By default it uses
// base 10 (SI) units such as kB and MB. Base 2 rendering is available via the
// Options, producing IEC units (KiB, MiB, ...) or JEDEC units (KB, MB, ...).
package filesize

import (
	"math"
	"strconv"
	"strings"
)

var siUnits = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var iecUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
var jedecUnits = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

// Options configures how a byte count is rendered.
type Options struct {
	// Base is the numeric base: 10 (divide by 1000) or 2 (divide by 1024).
	// Zero means the default of 10.
	Base int
	// Round is the number of decimal places to keep. When nil the default of
	// 2 is used.
	Round *int
	// Standard selects the unit family: "si" (kB), "iec" (KiB) or "jedec"
	// (KB). When empty it is derived from Base: base 2 uses "iec" and base 10
	// uses "si".
	Standard string
}

// FileSize converts a number of bytes into a human readable string using the
// default options (base 10, two decimals). For example FileSize(1337) returns
// "1.34 kB".
func FileSize(bytes float64) string {
	return FileSizeOpts(bytes, Options{})
}

// FileSizeOpts converts a number of bytes into a human readable string honouring
// the supplied Options.
func FileSizeOpts(bytes float64, opts Options) string {
	base := opts.Base
	if base == 0 {
		base = 10
	}
	round := 2
	if opts.Round != nil {
		round = *opts.Round
	}
	standard := opts.Standard
	if standard == "" {
		if base == 2 {
			standard = "iec"
		} else {
			standard = "si"
		}
	}

	var divisor float64
	var units []string
	switch standard {
	case "iec":
		divisor, units = 1024, iecUnits
	case "jedec":
		divisor, units = 1024, jedecUnits
	default: // "si"
		divisor, units = 1000, siUnits
	}

	negative := bytes < 0
	if negative {
		bytes = -bytes
	}
	prefix := ""
	if negative {
		prefix = "-"
	}

	if bytes == 0 {
		return prefix + "0 " + units[0]
	}

	exponent := int(math.Floor(math.Log(bytes) / math.Log(divisor)))
	if exponent < 0 {
		exponent = 0
	}
	if exponent > len(units)-1 {
		exponent = len(units) - 1
	}
	// Correct for floating point error in the log estimate.
	for exponent < len(units)-1 && bytes >= math.Pow(divisor, float64(exponent+1)) {
		exponent++
	}
	for exponent > 0 && bytes < math.Pow(divisor, float64(exponent)) {
		exponent--
	}

	val := bytes / math.Pow(divisor, float64(exponent))
	s := stripTrailingZeros(strconv.FormatFloat(val, 'f', round, 64))
	return prefix + s + " " + units[exponent]
}

// stripTrailingZeros removes trailing zeros (and a dangling decimal point).
func stripTrailingZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}
