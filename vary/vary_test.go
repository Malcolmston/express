package vary

import (
	"net/http"
	"testing"
)

func TestAppendToEmpty(t *testing.T) {
	got, err := Append("", "Accept")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Accept" {
		t.Fatalf("got %q", got)
	}
}

func TestAppendToExisting(t *testing.T) {
	got, err := Append("Accept-Encoding", "Accept")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Accept-Encoding, Accept" {
		t.Fatalf("got %q", got)
	}
}

func TestAppendDedupCaseInsensitive(t *testing.T) {
	got, err := Append("Accept", "accept")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Accept" {
		t.Fatalf("got %q want Accept", got)
	}
}

func TestAppendStarField(t *testing.T) {
	got, err := Append("Accept", "*")
	if err != nil {
		t.Fatal(err)
	}
	if got != "*" {
		t.Fatalf("got %q want *", got)
	}
}

func TestAppendExistingStar(t *testing.T) {
	got, err := Append("*", "Accept")
	if err != nil {
		t.Fatal(err)
	}
	if got != "*" {
		t.Fatalf("got %q want *", got)
	}
}

func TestAppendMultiple(t *testing.T) {
	got, err := Append("", "Accept", "Accept-Language")
	if err != nil {
		t.Fatal(err)
	}
	if got != "Accept, Accept-Language" {
		t.Fatalf("got %q", got)
	}
}

func TestAppendInvalid(t *testing.T) {
	_, err := Append("", "foo bar")
	if err == nil {
		t.Fatal("expected error for invalid field")
	}
}

func TestField(t *testing.T) {
	got, err := Field([]string{"Accept", "accept", "Accept-Encoding"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "Accept, Accept-Encoding" {
		t.Fatalf("got %q", got)
	}
}

func TestVaryHeaderMutation(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Accept-Encoding")
	Vary(h, "Accept")
	if got := h.Get("Vary"); got != "Accept-Encoding, Accept" {
		t.Fatalf("got %q", got)
	}
}

func TestVaryHeaderStar(t *testing.T) {
	h := http.Header{}
	Vary(h, "*")
	if got := h.Get("Vary"); got != "*" {
		t.Fatalf("got %q want *", got)
	}
}
