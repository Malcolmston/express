package useragent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func parse(t *testing.T, uaStr string) UserAgent {
	t.Helper()
	app := express.New()
	app.Use(New())
	var got UserAgent
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		got, _ = From(req)
		res.Send("ok")
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("User-Agent", uaStr)
	app.ServeHTTP(httptest.NewRecorder(), r)
	return got
}

func TestChromeWindows(t *testing.T) {
	ua := parse(t, "Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	if ua.Browser != "Chrome" {
		t.Fatalf("browser = %q", ua.Browser)
	}
	if ua.OS != "Windows" {
		t.Fatalf("os = %q", ua.OS)
	}
	if ua.Mobile {
		t.Fatalf("should not be mobile")
	}
}

func TestFirefox(t *testing.T) {
	ua := parse(t, "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0")
	if ua.Browser != "Firefox" || ua.OS != "Linux" {
		t.Fatalf("got %+v", ua)
	}
}

func TestMobileAndroid(t *testing.T) {
	ua := parse(t, "Mozilla/5.0 (Linux; Android 13; Pixel) AppleWebKit/537.36 Chrome/120.0 Mobile Safari/537.36")
	if !ua.Mobile {
		t.Fatalf("expected mobile")
	}
	if ua.OS != "Android" {
		t.Fatalf("os = %q", ua.OS)
	}
}

func TestUnknown(t *testing.T) {
	ua := parse(t, "curl/8.0")
	if ua.Browser != "Unknown" || ua.OS != "Unknown" {
		t.Fatalf("got %+v", ua)
	}
}
