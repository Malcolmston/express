// Package useragent provides middleware that performs lightweight parsing of
// the User-Agent request header into a small struct and stores it on the
// request for downstream handlers. It is a spiritual, dependency-free port of
// the Node "express-useragent" middleware and the "useragent" npm package,
// but deliberately tiny: rather than shipping a large table of regular
// expressions covering hundreds of browsers and devices, it recognises only a
// handful of common browser families and operating systems using plain
// substring matching. The goal is to give downstream handlers a fast, coarse
// signal ("this looks like mobile Chrome on Android") without pulling in any
// third-party parsing library.
//
// Reach for this middleware when you want to branch on broad client
// categories — for example, serving a lighter template to mobile clients,
// recording coarse browser statistics, or logging an OS breakdown — and can
// tolerate approximate results. It is not suitable as a security control or
// for precise device fingerprinting; the User-Agent header is client-supplied
// and easily spoofed, and the detection here is intentionally shallow. For
// blocking specific agents outright see the sibling useragentblock package,
// which this package does not itself perform.
//
// Chain position is early: mount New with app.Use before any handler that
// needs the parsed result. On each request the middleware reads the
// "User-Agent" request header via req.Get, calls Parse on it, and stores the
// resulting UserAgent value on the request with req.Set(Key, ua) where Key is
// the constant "useragent". It never writes response headers, never
// short-circuits, and always calls next() so the request continues down the
// chain. Downstream handlers retrieve the value with From(req), which returns
// the parsed UserAgent and a boolean reporting whether the middleware ran.
//
// The parsing semantics live in Parse and are worth understanding. Detection
// is lower-cased and ordered so that more specific tokens win: "edg" is
// checked before "chrome" (Edge advertises Chrome), "crios" counts as Chrome
// and Opera's "opr"/"opera" is matched before the generic engines. Unknown
// browsers and operating systems fall back to the literal string "Unknown"
// rather than an empty value, and Raw always preserves the original header.
// The Mobile flag is a simple OR over the "mobi", "android", "iphone" and
// "ipad" tokens, so an Android tablet or an unusual UA string may be
// classified only approximately.
//
// Compared with the Node originals this port is far narrower in scope: it does
// not expose version numbers, bot/crawler detection, device model names, the
// long list of per-vendor booleans (isChrome, isIE, isBot, and so on), or
// platform sub-classification beyond the coarse OS field. Only Edge, Opera,
// Firefox, Chrome and Safari are recognised as browsers, and only Android,
// iOS, Windows, macOS and Linux as operating systems; everything else is
// reported as "Unknown". Treat the output as a hint, not a contract, and if
// you need richer data parse req.Get("User-Agent") yourself.
package useragent

import (
	"strings"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the parsed UserAgent is stored.
const Key = "useragent"

// UserAgent is the result of basic substring-based User-Agent detection.
type UserAgent struct {
	// Browser is a coarse browser family name (e.g. "Chrome", "Firefox",
	// "Safari") or "Unknown".
	Browser string

	// OS is a coarse operating-system name (e.g. "Windows", "macOS",
	// "Android") or "Unknown".
	OS string

	// Mobile reports whether the agent appears to be a mobile device.
	Mobile bool

	// Raw is the original User-Agent header value.
	Raw string
}

// New returns middleware that parses the User-Agent header and stores the
// resulting UserAgent via req.Set(Key, ua).
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		req.Set(Key, Parse(req.Get("User-Agent")))
		next()
	}
}

// From retrieves the parsed UserAgent stored by this middleware. The second
// return value reports whether a value was present.
func From(req *express.Request) (UserAgent, bool) {
	v, ok := req.Value(Key)
	if !ok {
		return UserAgent{}, false
	}
	ua, ok := v.(UserAgent)
	return ua, ok
}

// Parse performs basic substring detection on a User-Agent string. Order
// matters: more specific tokens are checked before more general ones.
func Parse(raw string) UserAgent {
	ua := UserAgent{Browser: "Unknown", OS: "Unknown", Raw: raw}
	s := strings.ToLower(raw)

	switch {
	case strings.Contains(s, "edg"):
		ua.Browser = "Edge"
	case strings.Contains(s, "opr") || strings.Contains(s, "opera"):
		ua.Browser = "Opera"
	case strings.Contains(s, "firefox"):
		ua.Browser = "Firefox"
	case strings.Contains(s, "chrome") || strings.Contains(s, "crios"):
		ua.Browser = "Chrome"
	case strings.Contains(s, "safari"):
		ua.Browser = "Safari"
	}

	switch {
	case strings.Contains(s, "android"):
		ua.OS = "Android"
	case strings.Contains(s, "iphone") || strings.Contains(s, "ipad") || strings.Contains(s, "ios"):
		ua.OS = "iOS"
	case strings.Contains(s, "windows"):
		ua.OS = "Windows"
	case strings.Contains(s, "mac os") || strings.Contains(s, "macintosh"):
		ua.OS = "macOS"
	case strings.Contains(s, "linux"):
		ua.OS = "Linux"
	}

	ua.Mobile = strings.Contains(s, "mobi") ||
		strings.Contains(s, "android") ||
		strings.Contains(s, "iphone") ||
		strings.Contains(s, "ipad")

	return ua
}
