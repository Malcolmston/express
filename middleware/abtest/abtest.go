// Package abtest provides middleware that assigns each visitor a stable A/B
// test bucket and persists the assignment in a cookie so a returning visitor
// stays in the same bucket across requests. It is the Go analogue of the
// cookie-based split-testing middleware common in the Node ecosystem, such as
// express-ab, which sticks a user to one experiment variant for the lifetime
// of the cookie. It is packaged as a drop-in express.Handler rather than a
// route-splitting router: it labels the request and lets downstream handlers
// decide what to render for each bucket.
//
// Use this middleware whenever you want to divide traffic into deterministic
// cohorts for experiments, gradual feature rollouts, or multivariate testing,
// and you need the same user to see a consistent experience on every visit. A
// random-per-request choice would let a visitor flip between variants and
// pollute your metrics; persisting the choice in a cookie keeps the cohort
// stable so conversion can be attributed to a single variant. Because the
// number and names of buckets are configurable, it also serves canary
// rollouts (for example {"stable", "canary"}) as easily as a plain A/B split.
//
// Operationally the middleware belongs early in the chain, before any handler
// that varies its behavior by bucket. On each request it reads the configured
// cookie (Options.CookieName, default "abtest") via req.Cookie. If the cookie
// is present and its value is one of the configured Buckets, that assignment
// is reused; otherwise a fresh bucket is chosen and written back with
// res.Cookie using Path "/", HTTPOnly, SameSite=Lax, and the configured
// MaxAge. Either way the resolved bucket is stored on the request under an
// internal key, and next() is always called so the request continues down the
// chain. Downstream code retrieves the label with the exported Bucket helper.
//
// A new bucket is selected with crypto/rand for an unbiased, uniformly random
// draw across the bucket slice; if the random source fails, the first bucket
// is chosen so assignment can never panic or block. The defaults are applied
// once when the handler is built: an empty CookieName becomes "abtest" and an
// empty Buckets slice becomes {"A", "B"}. A MaxAge of 0 yields a session
// cookie that expires when the browser closes, while a positive MaxAge pins
// the cohort for that many seconds. Any incoming cookie value that is not a
// currently configured bucket is treated as invalid and triggers a fresh
// assignment, so shrinking or renaming the bucket set safely re-buckets stale
// visitors rather than leaking a label the application no longer understands.
//
// Compared with the Node originals, this port deliberately does not route or
// render anything itself: it neither swaps handlers per variant nor tracks
// conversions, leaving both concerns to the application. It also does not
// weight buckets — every bucket is equally likely — so a 90/10 split must be
// expressed by listing a label multiple times or by branching in your own
// handler. What it does match is the core promise of cookie-based split
// testing: a stable, cheap, dependency-free bucket that the rest of the
// application can read through a single accessor.
package abtest

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"github.com/malcolmston/express"
)

const contextKey = "bucket"

// Options configures the A/B test middleware.
type Options struct {
	// CookieName is the cookie that persists the assignment (default "abtest").
	CookieName string
	// Buckets is the set of bucket labels a visitor may be assigned to
	// (default {"A", "B"}).
	Buckets []string
	// MaxAge sets the cookie Max-Age in seconds (0 = session cookie).
	MaxAge int
}

func (o *Options) applyDefaults() {
	if o.CookieName == "" {
		o.CookieName = "abtest"
	}
	if len(o.Buckets) == 0 {
		o.Buckets = []string{"A", "B"}
	}
}

// New returns middleware that assigns and persists a stable bucket for each
// visitor. The assigned bucket is stored on the request; retrieve it with
// Bucket.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		bucket := req.Cookie(o.CookieName)
		if !contains(o.Buckets, bucket) {
			bucket = o.Buckets[randIndex(len(o.Buckets))]
			res.Cookie(o.CookieName, bucket, &express.CookieOptions{
				Path:     "/",
				MaxAge:   o.MaxAge,
				HTTPOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}
		req.Set(contextKey, bucket)
		next()
	}
}

// Bucket returns the A/B bucket assigned to the request, or "" if the
// middleware did not run.
func Bucket(req *express.Request) string {
	if v, ok := req.Value(contextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}

func randIndex(n int) int {
	if n <= 1 {
		return 0
	}
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(idx.Int64())
}
