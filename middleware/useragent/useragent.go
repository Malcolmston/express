// Package useragent provides middleware that performs lightweight parsing of
// the User-Agent request header into a small struct and stores it on the
// request for downstream handlers under the key "useragent".
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
