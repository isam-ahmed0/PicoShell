package screens

import (
	"image/color"
	"testing"

	"github.com/isam-ahmed0/PicoShell/internal/raster/software"
	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

// render draws a screen through the software rasterizer so tests can assert on
// pixels rather than on op lists.
func render(t *testing.T, st MenuState, bounds ui.Rect) *software.Surface {
	t.Helper()
	tm := theme.Dark()
	r := software.New(1)
	s := software.NewSurface(ui.Menu, "test", bounds, 1, tm.MenuBg)
	r.Draw(s, Menu(bounds, tm, st, r))
	return s
}

// rowPoint returns a pixel inside row i of the list, just past the highlight's
// left edge where no text is drawn.
func rowPoint(bounds ui.Rect, t theme.Theme, i int) (int, int) {
	box := menuBox(bounds, t)
	inner := box.Inset(t.Pad+1, t.Pad+1)
	searchH := min(MenuSearch, max(0, inner.H()-MenuRowH))
	rowH := min(MenuRowH, max(MenuRowH/2, inner.H()-searchH-t.Gap))
	y := inner.Y0 + searchH + t.Gap + float32(i)*rowH + rowH/2
	return int(inner.X0 + 4), int(y)
}

func rowBg(t *testing.T, s *software.Surface, x, y int) color.NRGBA {
	t.Helper()
	r, g, b, a := s.Image.At(x, y).RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func TestSelectionIsPaintedInTheRow(t *testing.T) {
	bounds := ui.R(0, 0, 600, 600)
	tm := theme.Dark()
	items := menuItems()

	unselected := render(t, MenuState{Items: items, Selected: -1}, bounds)
	selected := render(t, MenuState{Items: items, Selected: 1}, bounds)

	x, y := rowPoint(bounds, tm, 1)

	plain := rowBg(t, unselected, x, y)
	lit := rowBg(t, selected, x, y)
	if lit == plain {
		t.Fatalf("selected row is not painted: both %v", plain)
	}

	if lit != tm.ItemBgOn {
		t.Errorf("selected row background = %v, want ItemBgOn %v", lit, tm.ItemBgOn)
	}
	if plain == tm.ItemBgOn {
		t.Errorf("unselected row is painted as selected: %v", plain)
	}
}

func TestOnlyOneRowIsHighlighted(t *testing.T) {
	bounds := ui.R(0, 0, 600, 600)
	s := render(t, MenuState{Items: menuItems(), Selected: 2}, bounds)

	lit := 0
	for row := range len(menuItems()) {
		x, y := rowPoint(bounds, theme.Dark(), row)
		if rowBg(t, s, x, y) == theme.Dark().ItemBgOn {
			lit++
		}
	}
	if lit != 1 {
		t.Errorf("%d highlighted rows, want exactly 1", lit)
	}
}

func TestQueryDrawsAccentCaret(t *testing.T) {
	bounds := ui.R(0, 0, 600, 600)
	tm := theme.Dark()

	withCaret := render(t, MenuState{Items: menuItems(), Query: "er"}, bounds)
	found := false
	for y := int(menuBox(bounds, tm).Y0); y < int(menuBox(bounds, tm).Y0+MenuSearch); y++ {
		for x := int(menuBox(bounds, tm).X0); x < int(menuBox(bounds, tm).X0)+40; x++ {
			c := rowBg(t, withCaret, x, y)
			if c == tm.Accent {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("search field caret was not drawn in the accent colour")
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	bounds := ui.R(0, 0, 600, 600)
	st := MenuState{Items: menuItems(), Query: "e", Selected: 1, Scroll: 1}
	a := render(t, st, bounds)
	b := render(t, st, bounds)

	if len(a.Image.Pix) != len(b.Image.Pix) {
		t.Fatal("renders differ in size")
	}
	for i := range a.Image.Pix {
		if a.Image.Pix[i] != b.Image.Pix[i] {
			t.Fatalf("renders differ at byte %d, so the screen is not a pure function of state", i)
		}
	}
}

func TestRenderSurvivesTinyOutputs(t *testing.T) {
	for _, bounds := range []ui.Rect{
		ui.R(0, 0, 8, 8),
		ui.R(0, 0, 60, 40),
		ui.R(0, 0, 1, 1),
		ui.R(0, 0, 600, 600),
	} {
		s := render(t, MenuState{Items: menuItems(), Query: "e", Selected: 1, Scroll: 3}, bounds)
		if s.Image.Bounds().Dx() < 1 || s.Image.Bounds().Dy() < 1 {
			t.Errorf("bounds %v produced an empty surface", bounds)
		}
	}
}
