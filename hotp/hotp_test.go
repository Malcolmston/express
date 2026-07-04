package hotp

import "testing"

func TestGenerateRFC4226(t *testing.T) {
	secret := []byte("12345678901234567890")
	want := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	for counter, expected := range want {
		got := Generate(secret, uint64(counter), 6)
		if got != expected {
			t.Errorf("counter %d: got %q, want %q", counter, got, expected)
		}
	}
}

func TestVerify(t *testing.T) {
	secret := []byte("12345678901234567890")
	if !Verify(secret, 0, "755224", 6) {
		t.Error("expected verify to succeed for counter 0")
	}
	if Verify(secret, 0, "000000", 6) {
		t.Error("expected verify to fail for wrong code")
	}
	if Verify(secret, 1, "755224", 6) {
		t.Error("expected verify to fail for wrong counter")
	}
}
