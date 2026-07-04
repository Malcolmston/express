package statuses

import "testing"

func TestMessage(t *testing.T) {
	if got := Message(404); got != "Not Found" {
		t.Fatalf("got %q, want Not Found", got)
	}
	if got := Message(999); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestCode(t *testing.T) {
	code, err := Code("Not Found")
	if err != nil || code != 404 {
		t.Fatalf("got %d, %v; want 404, nil", code, err)
	}
	if _, err := Code("nonsense"); err == nil {
		t.Fatal("expected error for unknown message")
	}
}

func TestCodeCaseInsensitive(t *testing.T) {
	code, err := Code("not found")
	if err != nil || code != 404 {
		t.Fatalf("got %d, %v; want 404", code, err)
	}
}

func TestRoundTrip(t *testing.T) {
	for _, code := range Codes() {
		msg := Message(code)
		got, err := Code(msg)
		if err != nil {
			t.Errorf("code %d message %q: %v", code, msg, err)
			continue
		}
		if got != code {
			t.Errorf("round-trip mismatch: %d -> %q -> %d", code, msg, got)
		}
	}
}

func TestIsRedirect(t *testing.T) {
	for _, c := range []int{300, 301, 302, 303, 305, 307, 308} {
		if !IsRedirect(c) {
			t.Errorf("%d should be redirect", c)
		}
	}
	for _, c := range []int{200, 304, 404} {
		if IsRedirect(c) {
			t.Errorf("%d should not be redirect", c)
		}
	}
}

func TestIsRetry(t *testing.T) {
	for _, c := range []int{502, 503, 504} {
		if !IsRetry(c) {
			t.Errorf("%d should be retry", c)
		}
	}
	for _, c := range []int{429, 500, 200} {
		if IsRetry(c) {
			t.Errorf("%d should not be retry", c)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	for _, c := range []int{204, 205, 304} {
		if !IsEmpty(c) {
			t.Errorf("%d should be empty", c)
		}
	}
	if IsEmpty(200) {
		t.Error("200 should not be empty")
	}
}

func TestCodes(t *testing.T) {
	codes := Codes()
	if len(codes) == 0 {
		t.Fatal("expected non-empty codes")
	}
	for i := 1; i < len(codes); i++ {
		if codes[i] <= codes[i-1] {
			t.Fatal("codes not sorted ascending")
		}
	}
}
