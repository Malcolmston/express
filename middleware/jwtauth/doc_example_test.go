package jwtauth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/jwtauth"
)

// signHS256 mints an HS256 JWT for the example, mirroring how a real signer (or
// the package's own tests) would build a token: base64url-encode a
// {"alg":"HS256","typ":"JWT"} header and the claims, join them with a dot, and
// append the base64url-encoded HMAC-SHA256 signature over "header.payload".
func signHS256(secret []byte, claims map[string]any) string {
	enc := func(v any) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	signing := enc(map[string]any{"alg": "HS256", "typ": "JWT"}) + "." + enc(claims)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signing))
	return signing + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Example demonstrates protecting a route with the jwtauth middleware and
// reading the verified claims inside the handler. It shares a symmetric secret
// between the signer and the middleware, mints a valid HS256 token for the
// claims {"sub":"alice"}, and mounts jwtauth.New(Options{Secret: secret}) with
// app.Use so the route is guarded. A request is driven through
// net/http/httptest carrying the token in the "Authorization: Bearer <token>"
// header; the middleware verifies it, stores the claims under the default
// "claims" key, and the handler retrieves them with req.Value before echoing
// them as JSON. The status code and body are deterministic, so an Output block
// is included.
func Example() {
	secret := []byte("example-secret")

	app := express.New()
	app.Use(jwtauth.New(jwtauth.Options{Secret: secret}))
	app.Get("/me", func(req *express.Request, res *express.Response, next express.Next) {
		claims, _ := req.Value("claims")
		res.JSON(claims)
	})

	token := signHS256(secret, map[string]any{"sub": "alice"})
	r := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, r)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// {"sub":"alice"}
}
