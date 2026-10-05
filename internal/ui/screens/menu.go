// Package screens builds the DrawLists for PicoShell's surfaces. A screen is a
// pure function of state: same inputs, same pixels, on every platform.
package screens

import (
	"strings"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

// The menu is a floating window centred over the desktop, so it is sized
// relative to the output rather than filling it.
const (
	MenuWidth  = 460
	MenuMaxH   = 420
	MenuRowH   = 40
	MenuSearch = 44
)

type MenuItem struct {
	Title string
	Hint  string
}

type MenuState struct {
	Query    string
	Items    []MenuItem
	Selected int
	// Scroll is the row offset at the top of the list.
	Scroll int
}

// Matches filters items by a case-insensitive substring of the title or hint,
// which is enough for the app list and stays predictable across platforms.
func (st MenuState) Matches() []int {
	q := strings.ToLower(strings.TrimSpace(st.Query))
	if q == "" {
		idx := make([]int, len(st.Items))
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	var idx []int
	for i, it := range st.Items {
		if strings.Contains(strings.ToLower(it.Title), q) || strings.Contains(strings.ToLower(it.Hint), q) {
			idx = append(idx, i)
		}
	}
	return idx
}

// Menu draws the launcher: a search field over a filtered, scrollable list.
func Menu(bounds ui.Rect, t theme.Theme, st MenuState, r ui.Rasterizer) *ui.DrawList {
	d := ui.NewDrawList(bounds)

	box := menuBox(bounds, t)
	d.PushClip(box)
	d.RoundedRect(box, t.Radius*1.5, t.MenuBg)
	d.RoundedRect(box.Inset(0.5, 0.5), t.Radius*1.5, t.Border)
	inner := box.Inset(t.Pad+1, t.Pad+1)

	// Heights come from the box rather than being fixed, so a small output
	// shrinks the search field and rows instead of drawing past the edges.
	searchH := min(MenuSearch, max(0, inner.H()-MenuRowH))
	rowH := min(MenuRowH, max(MenuRowH/2, inner.H()-searchH-t.Gap))

	search := ui.R(inner.X0, inner.Y0, inner.X1, inner.Y0+searchH)
	d.RoundedRect(search, t.Radius, t.ItemBg)
	caret := ui.R(search.X0+t.Pad/2, search.Y0+t.Pad/2, search.X0+t.Pad/2+2, search.Y1-t.Pad/2)
	d.RoundedRect(caret, 1, t.Accent)
	queryBox := ui.R(search.X0+t.Pad*2, search.Y0, search.X1-t.Pad, search.Y1)
	if st.Query == "" {
		d.Text("Search", t.Font(14, ui.Regular), queryBox, t.TextMuted, ui.AlignStart)
	} else {
		d.Text(ui.Fit(r, t.Font(14, ui.Regular), st.Query, queryBox.W()),
			t.Font(14, ui.Regular), queryBox, t.Text, ui.AlignStart)
	}

	rows := st.Matches()
	list := ui.R(inner.X0, search.Y1+t.Gap, inner.X1, inner.Y1)
	if len(rows) == 0 || rowH <= 0 {
		d.Text("No matches", t.Font(13, ui.Regular), list, t.TextMuted, ui.AlignCenter)
		d.PopClip()
		return d
	}

	visible := int(list.H() / rowH)
	if visible < 1 {
		visible = 1
	}
	scroll := clamp(st.Scroll, 0, max(0, len(rows)-visible))
	for i := 0; i < visible && scroll+i < len(rows); i++ {
		row := ui.R(list.X0, list.Y0+float32(i)*rowH, list.X1, list.Y0+float32(i+1)*rowH)
		row = row.Intersect(list)
		if row.Empty() {
			break
		}
		it := st.Items[rows[scroll+i]]

		if rows[scroll+i] == st.Selected {
			d.RoundedRect(row.Inset(0, 2), t.Radius, t.ItemBgOn)
		}
		hint := row.Inset(0, 0)
		hint.X0 = max(row.X0, row.X1-t.Pad*7)
		title := ui.R(row.X0+t.Pad, row.Y0, hint.X0-t.Pad/2, row.Y1)
		if it.Hint != "" {
			d.Text(ui.Fit(r, t.Font(12, ui.Regular), it.Hint, hint.W()),
				t.Font(12, ui.Regular), hint, t.TextMuted, ui.AlignEnd)
		}
		d.Text(ui.Fit(r, t.Font(14, ui.Medium), it.Title, title.W()),
			t.Font(14, ui.Medium), title, t.Text, ui.AlignStart)
	}

	d.PopClip()
	return d
}

// menuBox centres the menu and caps its height so it never covers the whole
// desktop, which would leave no visual connection to the panel.
func menuBox(bounds ui.Rect, t theme.Theme) ui.Rect {
	w := min(MenuWidth, bounds.W()-t.Pad*4)
	h := min(MenuMaxH, bounds.H()-t.Pad*4)
	x0 := bounds.X0 + (bounds.W()-w)/2
	y0 := bounds.Y0 + (bounds.H()-h)/2
	return ui.R(x0, y0, x0+w, y0+h)
}

func clamp(v, lo, hi int) int { return max(lo, min(v, hi)) }
