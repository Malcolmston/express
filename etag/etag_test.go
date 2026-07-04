package etag

import (
	"testing"
	"time"
)

func TestEmptyContent(t *testing.T) {
	got := Generate([]byte(""), false)
	want := `"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestEmptyContentWeak(t *testing.T) {
	got := Generate([]byte(""), true)
	want := `W/"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk"`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestKnownContent(t *testing.T) {
	// "Hello World" -> sha1 base64 truncated, length 0xb = "b"
	got := Generate([]byte("Hello World"), false)
	want := `"b-Ck1VqNd45QIvq3AZd8XYQLvEhtA"`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestWeakPrefix(t *testing.T) {
	got := Generate([]byte("abc"), true)
	if got[:2] != "W/" {
		t.Fatalf("weak tag should start with W/, got %s", got)
	}
	strong := Generate([]byte("abc"), false)
	if got != "W/"+strong {
		t.Fatalf("weak tag mismatch: %s vs W/%s", got, strong)
	}
}

func TestGenerateStat(t *testing.T) {
	// size 1024 = 0x400, mtime 1000ms = 0x3e8
	mt := time.Unix(1, 0).UTC() // 1000 ms
	got := GenerateStat(1024, mt, false)
	want := `"400-3e8"`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestGenerateStatWeak(t *testing.T) {
	mt := time.Unix(0, 0).UTC()
	got := GenerateStat(0, mt, true)
	want := `W/"0-0"`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
