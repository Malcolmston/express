// Package nodepath is a standard-library-only Go port of Node.js's POSIX "path"
// module (https://nodejs.org/api/path.html), the utility Express and virtually
// every Node program uses to manipulate filesystem path strings. It reproduces
// the exact string semantics of node:path/posix — Basename, Dirname, Extname,
// Join, Normalize, Resolve, Relative, IsAbsolute, Parse and Format — so path
// logic can be ported from JavaScript to Go without behavioural surprises.
//
// The functions are chosen because Go's own path and path/filepath packages,
// while similar, differ from Node in edge cases that matter when porting code:
// Node's Extname treats a leading dot as part of the name (Extname(".bashrc")
// is ""), Node's Parse reports an empty Dir for a bare filename (whereas Dirname
// returns "."), Normalize preserves a trailing slash, and Join returns "." for
// empty input. Matching those rules exactly is the point of this package.
//
// Paths are treated as forward-slash separated POSIX paths regardless of host
// operating system, mirroring node:path/posix. Basename, Dirname, Extname,
// Normalize, Join, IsAbsolute, Parse and Format operate purely on their string
// arguments and are fully deterministic. Resolve and Relative fall back to the
// process working directory only when they must anchor a relative path to an
// absolute one; when every argument is already absolute they too are
// deterministic. Sep and Delimiter hold the POSIX separator and path-list
// delimiter. The implementation depends only on os (for the working directory)
// and strings.
package nodepath

import (
	"os"
	"strings"
)

// Sep is the POSIX path segment separator, "/".
const Sep = "/"

// Delimiter is the POSIX path-list delimiter (as used in PATH), ":".
const Delimiter = ":"

const npSlash = '/'
const npDot = '.'

// IsAbsolute reports whether p is an absolute POSIX path (begins with "/").
func IsAbsolute(p string) bool {
	return len(p) > 0 && p[0] == npSlash
}

// normalizeString is a faithful port of Node's internal normalizeStringPosix:
// it resolves "." and ".." segments and collapses repeated separators over a
// separatorless-relative view of the path. When allowAboveRoot is true, leading
// ".." segments that cannot be resolved are preserved.
func npNormalizeString(path string, allowAboveRoot bool) string {
	var res strings.Builder
	lastSegmentLength := 0
	lastSlash := -1
	dots := 0
	var code byte
	for i := 0; i <= len(path); i++ {
		if i < len(path) {
			code = path[i]
		} else if code == npSlash {
			break
		} else {
			code = npSlash
		}
		if code == npSlash {
			if lastSlash == i-1 || dots == 1 {
				// no-op
			} else if dots == 2 {
				s := res.String()
				if len(s) < 2 || lastSegmentLength != 2 ||
					s[len(s)-1] != npDot || s[len(s)-2] != npDot {
					if len(s) > 2 {
						lastSlashIndex := strings.LastIndexByte(s, npSlash)
						if lastSlashIndex == -1 {
							res.Reset()
							lastSegmentLength = 0
						} else {
							ns := s[:lastSlashIndex]
							res.Reset()
							res.WriteString(ns)
							lastSegmentLength = len(ns) - 1 - strings.LastIndexByte(ns, npSlash)
						}
						lastSlash = i
						dots = 0
						continue
					} else if len(s) != 0 {
						res.Reset()
						lastSegmentLength = 0
						lastSlash = i
						dots = 0
						continue
					}
				}
				if allowAboveRoot {
					if res.Len() > 0 {
						res.WriteString("/..")
					} else {
						res.WriteString("..")
					}
					lastSegmentLength = 2
				}
			} else {
				if res.Len() > 0 {
					res.WriteByte(npSlash)
					res.WriteString(path[lastSlash+1 : i])
				} else {
					res.WriteString(path[lastSlash+1 : i])
				}
				lastSegmentLength = i - lastSlash - 1
			}
			lastSlash = i
			dots = 0
		} else if code == npDot && dots != -1 {
			dots++
		} else {
			dots = -1
		}
	}
	return res.String()
}

// Normalize normalizes a POSIX path, resolving ".." and "." segments and
// collapsing repeated separators. A trailing separator is preserved, and an
// empty path returns ".".
func Normalize(p string) string {
	if len(p) == 0 {
		return "."
	}
	isAbs := p[0] == npSlash
	trailing := p[len(p)-1] == npSlash
	q := npNormalizeString(p, !isAbs)
	if len(q) == 0 {
		if isAbs {
			return "/"
		}
		if trailing {
			return "./"
		}
		return "."
	}
	if trailing {
		q += "/"
	}
	if isAbs {
		return "/" + q
	}
	return q
}

// Join joins path segments with the POSIX separator and normalizes the result.
// Zero-length segments are ignored; joining nothing (or only empties) yields ".".
func Join(elem ...string) string {
	var joined string
	set := false
	for _, e := range elem {
		if len(e) > 0 {
			if !set {
				joined = e
				set = true
			} else {
				joined += "/" + e
			}
		}
	}
	if !set {
		return "."
	}
	return Normalize(joined)
}

// Basename returns the last portion of a path, with any trailing separators
// removed. Basename("/foo/bar/") is "bar".
func Basename(p string) string {
	return npBasename(p, "")
}

// BasenameExt returns the last portion of a path with the given suffix removed
// when the basename ends with it. BasenameExt("/a/index.html", ".html") is
// "index". It mirrors Node's two-argument path.basename(path, ext).
func BasenameExt(p, ext string) string {
	return npBasename(p, ext)
}

func npBasename(path, suffix string) string {
	start := 0
	end := -1
	matchedSlash := true
	if len(suffix) > 0 && len(suffix) <= len(path) {
		if suffix == path {
			return ""
		}
		extIdx := len(suffix) - 1
		firstNonSlashEnd := -1
		for i := len(path) - 1; i >= 0; i-- {
			code := path[i]
			if code == npSlash {
				if !matchedSlash {
					start = i + 1
					break
				}
			} else {
				if firstNonSlashEnd == -1 {
					matchedSlash = false
					firstNonSlashEnd = i + 1
				}
				if extIdx >= 0 {
					if code == suffix[extIdx] {
						extIdx--
						if extIdx == -1 {
							end = i
						}
					} else {
						extIdx = -1
						end = firstNonSlashEnd
					}
				}
			}
		}
		if start == end {
			end = firstNonSlashEnd
		} else if end == -1 {
			end = len(path)
		}
		if end < start {
			return ""
		}
		return path[start:end]
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == npSlash {
			if !matchedSlash {
				start = i + 1
				break
			}
		} else if end == -1 {
			matchedSlash = false
			end = i + 1
		}
	}
	if end == -1 {
		return ""
	}
	return path[start:end]
}

// Dirname returns the directory portion of a path (everything before the last
// segment). Dirname("/foo/bar/baz") is "/foo/bar", Dirname("foo") is ".", and
// Dirname("/") is "/".
func Dirname(p string) string {
	if len(p) == 0 {
		return "."
	}
	hasRoot := p[0] == npSlash
	end := -1
	matchedSlash := true
	for i := len(p) - 1; i >= 1; i-- {
		if p[i] == npSlash {
			if !matchedSlash {
				end = i
				break
			}
		} else {
			matchedSlash = false
		}
	}
	if end == -1 {
		if hasRoot {
			return "/"
		}
		return "."
	}
	if hasRoot && end == 1 {
		return "//"
	}
	return p[:end]
}

// Extname returns the extension of the last path segment, from the last "." to
// the end. A leading dot is treated as part of the name, so Extname(".bashrc")
// and Extname("index") are both "", while Extname("index.coffee.md") is ".md".
func Extname(p string) string {
	startDot := -1
	startPart := 0
	end := -1
	matchedSlash := true
	preDotState := 0
	for i := len(p) - 1; i >= 0; i-- {
		code := p[i]
		if code == npSlash {
			if !matchedSlash {
				startPart = i + 1
				break
			}
			continue
		}
		if end == -1 {
			matchedSlash = false
			end = i + 1
		}
		if code == npDot {
			if startDot == -1 {
				startDot = i
			} else if preDotState != 1 {
				preDotState = 1
			}
		} else if startDot != -1 {
			preDotState = -1
		}
	}
	if startDot == -1 || end == -1 || preDotState == 0 ||
		(preDotState == 1 && startDot == end-1 && startDot == startPart+1) {
		return ""
	}
	return p[startDot:end]
}

// Resolve resolves a sequence of path segments into an absolute path, processing
// from right to left until an absolute path is built; if the segments never
// reach an absolute path they are anchored to the process working directory.
func Resolve(elem ...string) string {
	var resolvedPath string
	resolvedAbsolute := false
	for i := len(elem) - 1; i >= -1 && !resolvedAbsolute; i-- {
		var path string
		if i >= 0 {
			path = elem[i]
		} else {
			path, _ = os.Getwd()
		}
		if len(path) == 0 {
			continue
		}
		resolvedPath = path + "/" + resolvedPath
		resolvedAbsolute = path[0] == npSlash
	}
	resolvedPath = npNormalizeString(resolvedPath, !resolvedAbsolute)
	if resolvedAbsolute {
		return "/" + resolvedPath
	}
	if len(resolvedPath) > 0 {
		return resolvedPath
	}
	return "."
}

// Relative returns the relative path from from to to. Both operands are first
// resolved to absolute paths (using the working directory only for relative
// inputs), so passing absolute paths keeps the result deterministic.
func Relative(from, to string) string {
	if from == to {
		return ""
	}
	from = Resolve(from)
	to = Resolve(to)
	if from == to {
		return ""
	}
	fromStart := 1
	fromEnd := len(from)
	fromLen := fromEnd - fromStart
	toStart := 1
	toLen := len(to) - toStart
	length := fromLen
	if toLen < length {
		length = toLen
	}
	lastCommonSep := -1
	i := 0
	for ; i < length; i++ {
		fromCode := from[fromStart+i]
		if fromCode != to[toStart+i] {
			break
		} else if fromCode == npSlash {
			lastCommonSep = i
		}
	}
	if i == length {
		if toLen > length {
			if to[toStart+i] == npSlash {
				return to[toStart+i+1:]
			}
			if i == 0 {
				return to[toStart+i:]
			}
		} else if fromLen > length {
			if from[fromStart+i] == npSlash {
				lastCommonSep = i
			} else if i == 0 {
				lastCommonSep = 0
			}
		}
	}
	var out strings.Builder
	for i = fromStart + lastCommonSep + 1; i <= fromEnd; i++ {
		if i == fromEnd || from[i] == npSlash {
			if out.Len() == 0 {
				out.WriteString("..")
			} else {
				out.WriteString("/..")
			}
		}
	}
	return out.String() + to[toStart+lastCommonSep:]
}

// ParsedPath is the structured decomposition of a path returned by Parse:
// Root is the filesystem root ("/" or ""), Dir is the full directory,
// Base is the final segment, Name is Base without its extension, and Ext is the
// extension including the leading dot.
type ParsedPath struct {
	Root string
	Dir  string
	Base string
	Name string
	Ext  string
}

// Parse decomposes a path into a ParsedPath. Unlike Dirname, Parse reports an
// empty Dir for a bare filename ("foo.txt"), matching Node's path.parse.
func Parse(p string) ParsedPath {
	var ret ParsedPath
	if len(p) == 0 {
		return ret
	}
	isAbs := p[0] == npSlash
	start := 0
	if isAbs {
		ret.Root = "/"
		start = 1
	}
	startDot := -1
	startPart := 0
	end := -1
	matchedSlash := true
	preDotState := 0
	i := len(p) - 1
	for ; i >= start; i-- {
		code := p[i]
		if code == npSlash {
			if !matchedSlash {
				startPart = i + 1
				break
			}
			continue
		}
		if end == -1 {
			matchedSlash = false
			end = i + 1
		}
		if code == npDot {
			if startDot == -1 {
				startDot = i
			} else if preDotState != 1 {
				preDotState = 1
			}
		} else if startDot != -1 {
			preDotState = -1
		}
	}
	if end != -1 {
		start2 := startPart
		if startPart == 0 && isAbs {
			start2 = 1
		}
		if startDot == -1 || preDotState == 0 ||
			(preDotState == 1 && startDot == end-1 && startDot == startPart+1) {
			ret.Base = p[start2:end]
			ret.Name = ret.Base
		} else {
			ret.Name = p[start2:startDot]
			ret.Base = p[start2:end]
			ret.Ext = p[startDot:end]
		}
	}
	if startPart > 0 {
		ret.Dir = p[:startPart-1]
	} else if isAbs {
		ret.Dir = "/"
	}
	return ret
}

// Format builds a path string from a ParsedPath, the inverse of Parse. Dir (or
// Root when Dir is empty) is joined to Base (or Name+Ext when Base is empty).
func Format(pp ParsedPath) string {
	dir := pp.Dir
	if dir == "" {
		dir = pp.Root
	}
	base := pp.Base
	if base == "" {
		base = pp.Name + pp.Ext
	}
	if dir == "" {
		return base
	}
	if dir == pp.Root {
		return dir + base
	}
	return dir + "/" + base
}
