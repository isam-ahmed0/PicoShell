// Package screens builds the DrawLists for PicoShell's surfaces. A screen is a
// pure function of state: same inputs, same pixels, on every platform.
package screens

import (
	"strconv"
	"unicode"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

// PanelHeight is the bar's height in dp.
const PanelHeight = 48

const (
	itemWidth  = 132
	iconSize   = 26
	clockWidth = 96
)

type PanelItem struct {
	Title  string
	Active bool
}

type PanelState struct {
	Items  []PanelItem
	Clock  string
	Unread int
}

// Panel draws the bar that sits along the bottom of the screen, which is where
// PicoShell gets its name.
func Panel(bounds ui.Rect, t theme.Theme, st PanelState) *ui.DrawList {
	d := ui.NewDrawList(bounds)
	bar := ui.R(bounds.X0, bounds.Y0, bounds.X1, bounds.Y0+PanelHeight)

	d.Rect(bar, t.BarBg)
	d.Rect(ui.R(bar.X0, bar.Y0, bar.X1, bar.Y0+1), t.Border)

	d.PushClip(bar)
	x := bar.X0 + t.Pad
	for _, item := range st.Items {
		cell := ui.R(x, bar.Y0+t.Pad, x+itemWidth, bar.Y1-t.Pad)
		if cell.X1 > bar.X1-clockWidth {
			break
		}
		x = cell.X1 + t.Gap

		bg := t.ItemBg
		if item.Active {
			bg = t.ItemBgOn
		}
		d.PushClip(cell)
		d.RoundedRect(cell, t.Radius, bg)

		icon := ui.R(cell.X0+t.Pad, cell.Y0+t.Pad, cell.X0+t.Pad+iconSize, cell.Y1-t.Pad)
		iconBg := t.Accent
		if !item.Active {
			iconBg = t.Border
		}
		d.RoundedRect(icon, t.Radius/2, iconBg)
		d.Text(initial(item.Title), t.Font(12, ui.Bold), icon, t.Text, ui.AlignCenter)

		label := ui.R(icon.X1+t.Pad, cell.Y0, cell.X1-t.Pad, cell.Y1)
		d.Text(item.Title, t.Font(13, ui.Regular), label, t.Text, ui.AlignStart)
		d.PopClip()
	}

	clock := ui.R(bar.X1-clockWidth, bar.Y0, bar.X1-t.Pad, bar.Y1)
	if st.Unread > 0 {
		d.RoundedRect(ui.R(clock.X0, bar.Y0+t.Pad, clock.X0+iconSize, bar.Y1-t.Pad), t.Radius/2, t.Accent)
		d.Text(strconv.Itoa(min(st.Unread, 99)), t.Font(12, ui.Bold),
			ui.R(clock.X0, bar.Y0+t.Pad, clock.X0+iconSize, bar.Y1-t.Pad), t.Text, ui.AlignCenter)
		clock = clock.Offset(iconSize+t.Gap, 0)
	}
	d.Text(st.Clock, t.Font(13, ui.Medium), clock, t.Text, ui.AlignEnd)
	d.PopClip()

	return d
}

func initial(s string) string {
	for _, r := range s {
		return string(unicode.ToUpper(r))
	}
	return ""
}
