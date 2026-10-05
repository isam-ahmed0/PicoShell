package ui

import (
	"fmt"
	"strings"
	"testing"
)

// runeRaster measures one fixed width per rune, so tests can predict truncation
// by counting characters.
type runeRaster struct {
	unit   float32
	height float32
	caps   Caps
}

func (r runeRaster) Caps() Caps { return r.caps }

func (r runeRaster) Measure(_ Font, s string) Size {
	return Size{W: float32(len([]rune(s))) * r.unit, H: r.height}
}

func (r runeRaster) LineHeight(Font) float32 { return r.height }

func (r runeRaster) Draw(Surface, *DrawList) {}

func fitter() runeRaster { return runeRaster{unit: 10, height: 16} }

func TestFitLeavesShortTextAlone(t *testing.T) {
	if got := Fit(fitter(), Font{}, "Documents", 90); got != "Documents" {
		t.Errorf("Fit = %q, want unchanged at exact width", got)
	}
	if got := Fit(fitter(), Font{}, "Documents", 200); got != "Documents" {
		t.Errorf("Fit = %q, want unchanged when it fits", got)
	}
}

func TestFitAppendsEllipsis(t *testing.T) {
	// One unit is 10dp and the ellipsis costs one unit, so a 60dp budget holds
	// five runes plus the marker.
	for _, tc := range []struct{ maxW, want string }{
		{"60", "Docum…"},
		{"50", "Docu…"},
		{"20", "D…"},
	} {
		if got := Fit(fitter(), Font{}, "Documents", mustFloat(t, tc.maxW)); got != tc.want {
			t.Errorf("Fit(%v) = %q, want %q", tc.maxW, got, tc.want)
		}
	}
}

func TestFitNeverExceedsWidth(t *testing.T) {
	r := fitter()
	for maxW := float32(5); maxW <= 200; maxW += 5 {
		got := Fit(r, Font{}, "Documents", maxW)
		if w := r.Measure(Font{}, got).W; w > maxW {
			t.Errorf("Fit(%v) = %q measuring %v, wider than the budget", maxW, got, w)
		}
	}
}

func TestFitHandlesMultiByte(t *testing.T) {
	r := fitter()
	if got := Fit(r, Font{}, "日本語", 25); got != "日…" {
		t.Errorf("Fit = %q, want %q (no split runes)", got, "日…")
	}
	for _, s := range []string{"", "…"} {
		if got := Fit(r, Font{}, s, 100); got != s {
			t.Errorf("Fit(%q) = %q, want unchanged", s, got)
		}
	}
}

func TestFitDegenerateWidths(t *testing.T) {
	r := fitter()
	for _, maxW := range []float32{0, -5, 5} {
		if got := Fit(r, Font{}, "Documents", maxW); got != "" {
			t.Errorf("Fit(%v) = %q, want empty when the budget cannot hold the marker", maxW, got)
		}
	}
	if got := Fit(r, Font{}, "Documents", 5); got != "" {
		t.Errorf("Fit narrower than the ellipsis = %q, want empty", got)
	}
}

func TestFitLeftKeepsTail(t *testing.T) {
	if got := FitLeft(fitter(), Font{}, "/home/user/Documents", 210); got != "/home/user/Documents" {
		t.Errorf("FitLeft = %q, want unchanged at exact width", got)
	}
	if got := FitLeft(fitter(), Font{}, "/home/user/Documents", 30); got != "…ts" {
		t.Errorf("FitLeft = %q, want %q", got, "…ts")
	}
}

func TestFitLeftNeverExceedsWidthOrLosesTail(t *testing.T) {
	r := fitter()
	const s = "/home/user/Documents"
	for maxW := float32(5); maxW <= 250; maxW += 5 {
		got := FitLeft(r, Font{}, s, maxW)
		if w := r.Measure(Font{}, got).W; w > maxW {
			t.Errorf("FitLeft(%v) = %q measuring %v", maxW, got, w)
		}
		if tail := strings.TrimPrefix(got, Ellipsis); tail != "" && !strings.HasSuffix(s, tail) {
			t.Errorf("FitLeft(%v) = %q, want %q to be a tail of the input", maxW, got, tail)
		}
	}
}

func TestFittingWithoutARasterizer(t *testing.T) {
	if got := Fit(nil, Font{}, "Documents", 100); got != "" {
		t.Errorf("Fit with no rasterizer = %q, want empty", got)
	}
	if got := FitLeft(nil, Font{}, "Documents", 100); got != "" {
		t.Errorf("FitLeft with no rasterizer = %q, want empty", got)
	}
}

func mustFloat(t *testing.T, s string) float32 {
	t.Helper()
	var f float32
	if _, err := fmt.Sscanf(s, "%g", &f); err != nil {
		t.Fatalf("bad test width %q: %v", s, err)
	}
	return f
}
