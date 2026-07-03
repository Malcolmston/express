// Package abtest provides middleware that assigns each visitor a stable A/B
// test bucket. The assignment is persisted in a cookie so a returning visitor
// stays in the same bucket across requests.
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
