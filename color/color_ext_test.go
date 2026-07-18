package color

import "testing"

func TestContrastColor(t *testing.T) {
	if got, _ := ContrastColor("#ffffff"); got != "#000000" {
		t.Errorf("ContrastColor white bg = %q", got)
	}
	if got, _ := ContrastColor("#000000"); got != "#ffffff" {
		t.Errorf("ContrastColor black bg = %q", got)
	}
	if got, _ := ContrastColor("#0000ff"); got != "#ffffff" {
		t.Errorf("ContrastColor blue bg = %q", got)
	}
	if _, err := ContrastColor("xyz"); err == nil {
		t.Error("expected error")
	}
}

func TestTintShade(t *testing.T) {
	if got, _ := Tint("#000000", 0.5); got != "#808080" {
		t.Errorf("Tint = %q", got)
	}
	if got, _ := Shade("#ffffff", 0.5); got != "#808080" {
		t.Errorf("Shade = %q", got)
	}
	if got, _ := Tint("#ff0000", 0); got != "#ff0000" {
		t.Errorf("Tint 0 = %q", got)
	}
}
