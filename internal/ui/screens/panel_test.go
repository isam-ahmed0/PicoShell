package screens

import (
	"testing"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

// Each item contributes exactly one start-aligned label, which is what the
// tests below count: the icon initial is centred and the clock is end-aligned.
func itemCount(list *ui.DrawList) int {
	n := 0
	for _, o := range list.Ops() {
		if o.Kind == ui.OpText && o.Text.Align == ui.AlignStart {
			n++
		}
	}
	return n
}

func items(n int) []PanelItem {
	out := make([]PanelItem, n)
	for i := range out {
		out[i] = PanelItem{Title: "App"}
	}
	return out
}

func TestPanelDropsItemsThatWouldReachTheClock(t *testing.T) {
	tm := theme.Dark()
	state := PanelState{Items: items(12), Clock: "09:41", Unread: 3}

	// The clock column is reserved on the right; a narrow bar must show fewer
	// items rather than draw them underneath it.
	if got := itemCount(Panel(ui.R(0, 0, 320, PanelHeight), tm, state)); got != 1 {
		t.Errorf("320dp bar drew %d items, want 1", got)
	}
	wide := itemCount(Panel(ui.R(0, 0, 1280, PanelHeight), tm, state))
	if wide <= 1 || wide > len(state.Items) {
		t.Errorf("1280dp bar drew %d items, want between 2 and %d", wide, len(state.Items))
	}
	if evenWider := itemCount(Panel(ui.R(0, 0, 4096, PanelHeight), tm, state)); evenWider != len(state.Items) {
		t.Errorf("4096dp bar drew %d items, want all %d", evenWider, len(state.Items))
	}
}

func TestPanelItemsStayClearOfClockColumn(t *testing.T) {
	tm := theme.Dark()
	for _, width := range []float32{320, 640, 1280, 2560} {
		bounds := ui.R(0, 0, width, PanelHeight)
		list := Panel(bounds, tm, PanelState{Items: items(8), Clock: "09:41", Unread: 3})
		clock := ui.R(width-clockWidth, 0, width, PanelHeight)

		for _, o := range list.Ops() {
			if o.Kind != ui.OpText || o.Text.Align != ui.AlignStart {
				continue
			}
			if !o.Rect.Intersect(clock).Empty() {
				t.Errorf("width %.0f: item %q at %v overlaps clock column %v", width, o.Text.Text, o.Rect, clock)
			}
		}
	}
}

func TestPanelStaysInsideBounds(t *testing.T) {
	tm := theme.Dark()
	bounds := ui.R(0, 0, 1280, PanelHeight)
	list := Panel(bounds, tm, PanelState{Items: items(4), Clock: "09:41"})

	for i, o := range list.Ops() {
		if o.Rect.Intersect(bounds) != o.Rect {
			t.Errorf("op %d (%v) rect %v escapes bounds %v", i, o.Kind, o.Rect, bounds)
		}
	}
}

func TestPanelWithoutUnreadHasNoBadge(t *testing.T) {
	tm := theme.Dark()
	list := Panel(ui.R(0, 0, 1280, PanelHeight), tm, PanelState{Items: items(2), Clock: "09:41"})
	if got := itemCount(list); got != 2 {
		t.Errorf("drew %d items, want 2", got)
	}
}
