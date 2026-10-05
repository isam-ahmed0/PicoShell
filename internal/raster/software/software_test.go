package software

import (
	"image/color"
	"testing"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
)

var (
	bg      = color.NRGBA{0, 0, 0, 0xff}
	fill    = color.NRGBA{0xff, 0xff, 0xff, 0xff}
	inkFont = ui.Font{Family: "Test", Size: 14}
)

func surface(w, h float32) *Surface {
	return NewSurface(ui.Panel, "test", ui.R(0, 0, w, h), 1, bg)
}

func at(t *testing.T, s *Surface, x, y int) color.NRGBA {
	t.Helper()
	r, g, b, a := s.Image.At(x, y).RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func equal(a, b color.NRGBA) bool { return a.R == b.R && a.G == b.G && a.B == b.B && a.A == b.A }

// inkColumns returns the x range that received ink, so alignment can be
// asserted without pinning exact glyph pixels.
func inkColumns(t *testing.T, s *Surface) (int, int) {
	t.Helper()
	lo, hi := -1, -1
	for x := 0; x < s.Image.Bounds().Dx(); x++ {
		for y := 0; y < s.Image.Bounds().Dy(); y++ {
			if !equal(at(t, s, x, y), bg) {
				if lo < 0 {
					lo = x
				}
				hi = x
				break
			}
		}
	}
	return lo, hi
}

func TestRectFill(t *testing.T) {
	s := surface(16, 16)
	d := ui.NewDrawList(ui.R(0, 0, 16, 16))
	d.Rect(ui.R(2, 2, 6, 6), fill)
	New(1).Draw(s, d)

	if got := at(t, s, 3, 3); !equal(got, fill) {
		t.Errorf("inside rect = %v, want %v", got, fill)
	}
	if got := at(t, s, 1, 1); !equal(got, bg) {
		t.Errorf("outside rect = %v, want background %v", got, bg)
	}
}

func TestClipLimitsFill(t *testing.T) {
	s := surface(16, 16)
	d := ui.NewDrawList(ui.R(0, 0, 16, 16))
	d.PushClip(ui.R(0, 0, 4, 4))
	d.Rect(ui.R(0, 0, 16, 16), fill)
	d.PopClip()
	New(1).Draw(s, d)

	if got := at(t, s, 3, 3); !equal(got, fill) {
		t.Errorf("clipped area = %v, want %v", got, fill)
	}
	if got := at(t, s, 5, 5); !equal(got, bg) {
		t.Errorf("outside clip = %v, want background %v", got, bg)
	}
}

func TestRoundedRectLeavesCorners(t *testing.T) {
	s := surface(32, 32)
	d := ui.NewDrawList(ui.R(0, 0, 32, 32))
	d.RoundedRect(ui.R(4, 4, 28, 28), 8, fill)
	New(1).Draw(s, d)

	if got := at(t, s, 16, 16); !equal(got, fill) {
		t.Errorf("centre = %v, want %v", got, fill)
	}
	if got := at(t, s, 4, 4); !equal(got, bg) {
		t.Errorf("corner = %v, want background %v", got, bg)
	}
}

func TestTextStaysInBox(t *testing.T) {
	const width = 200
	s := surface(width, 24)
	d := ui.NewDrawList(ui.R(0, 0, width, 24))
	d.Text("PicoShell", inkFont, ui.R(4, 2, 60, 22), fill, ui.AlignStart)
	New(1).Draw(s, d)

	lo, hi := inkColumns(t, s)
	if lo < 0 {
		t.Fatal("no text pixels drawn")
	}
	// One pixel of slack for the left side bearing of the first glyph.
	if lo < 4 || hi >= 60 {
		t.Errorf("ink spans x %d..%d, want inside the 4..60 box", lo, hi)
	}
}

func TestTextAlignment(t *testing.T) {
	const width, box = 200, 100
	draw := func(a ui.Align) (int, int) {
		s := surface(width, 24)
		d := ui.NewDrawList(ui.R(0, 0, width, 24))
		d.Text("PicoShell", inkFont, ui.R(4, 2, box, 22), fill, a)
		New(1).Draw(s, d)
		return inkColumns(t, s)
	}

	startLo, startHi := draw(ui.AlignStart)
	centerLo, centerHi := draw(ui.AlignCenter)
	endLo, endHi := draw(ui.AlignEnd)

	if startLo < 4 || startLo > 5 {
		t.Errorf("AlignStart ink starts at %d, want ~4", startLo)
	}
	if centerLo <= startLo || centerHi >= endHi {
		t.Errorf("AlignStart %d..%d, AlignCenter %d..%d, AlignEnd %d..%d: want centred between start and end",
			startLo, startHi, centerLo, centerHi, endLo, endHi)
	}
	if endHi >= box {
		t.Errorf("AlignEnd ink ends at %d, want left of the box edge %d", endHi, box)
	}
}

func TestMeasureAgreesWithDraw(t *testing.T) {
	const width = 200
	s := surface(width, 24)
	d := ui.NewDrawList(ui.R(0, 0, width, 24))
	d.Text("PicoShell", inkFont, ui.R(0, 0, width, 24), fill, ui.AlignStart)
	New(1).Draw(s, d)

	r := New(1)
	m := r.Measure(inkFont, "PicoShell")
	lo, hi := inkColumns(t, s)
	// Advance includes side bearings, so the ink is a little narrower.
	if ink := float32(hi - lo); ink > m.W || m.W-ink > 6 {
		t.Errorf("measured width %v does not match ink width %v", m.W, ink)
	}
	if r.LineHeight(inkFont) <= 0 {
		t.Errorf("LineHeight = %v, want positive", r.LineHeight(inkFont))
	}
}
