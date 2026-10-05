package screens

import (
	"strings"
	"testing"

	"github.com/isam-ahmed0/PicoShell/internal/raster/software"
	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

func menuItems() []MenuItem {
	return []MenuItem{
		{Title: "Files", Hint: "Folders"},
		{Title: "Browser", Hint: "Web"},
		{Title: "Terminal", Hint: "Shell"},
		{Title: "A Very Long Application Name Indeed", Hint: "Trimmed"},
	}
}

func TestMatchesIsCaseInsensitiveAndCoversHint(t *testing.T) {
	all := MenuState{Items: menuItems()}
	if got := len(all.Matches()); got != len(all.Items) {
		t.Errorf("empty query matched %d of %d", got, len(all.Items))
	}

	title := MenuState{Items: menuItems(), Query: "TERMIN"}
	got := title.Matches()
	if len(got) != 1 || title.Items[got[0]].Title != "Terminal" {
		t.Errorf("query %q = %v, want just Terminal", title.Query, got)
	}

	hint := MenuState{Items: menuItems(), Query: "BROWS"}
	got = hint.Matches()
	if len(got) != 1 || hint.Items[got[0]].Title != "Browser" {
		t.Errorf("hint query %q = %v, want just Browser (matched on its hint)", hint.Query, got)
	}
	if got := len(MenuState{Items: menuItems(), Query: "  "}.Matches()); got != len(menuItems()) {
		t.Errorf("whitespace query matched %d, want all", got)
	}
	none := MenuState{Items: menuItems(), Query: "zzz"}.Matches()
	if len(none) != 0 {
		t.Errorf("unmatched query = %v, want none", none)
	}
}

func TestMenuShowsPlaceholderOnlyWhenEmpty(t *testing.T) {
	tm := theme.Dark()
	bounds := ui.R(0, 0, 1280, 720)
	r := software.New(1)

	hasText := func(list *ui.DrawList, want string) bool {
		for _, o := range list.Ops() {
			if o.Kind == ui.OpText && o.Text.Text == want {
				return true
			}
		}
		return false
	}

	empty := Menu(bounds, tm, MenuState{Items: menuItems()}, r)
	if !hasText(empty, "Search") {
		t.Error("empty menu is missing the search placeholder")
	}
	filled := Menu(bounds, tm, MenuState{Items: menuItems(), Query: "Bro"}, r)
	if hasText(filled, "Search") {
		t.Error("menu with a query still draws the placeholder")
	}
	if !hasText(filled, "Bro") {
		t.Error("menu is missing the query text")
	}
}

func TestMenuReportsEmptyResult(t *testing.T) {
	tm := theme.Dark()
	list := Menu(ui.R(0, 0, 1280, 720), tm, MenuState{Items: menuItems(), Query: "zzz"}, software.New(1))
	found := false
	for _, o := range list.Ops() {
		if o.Kind == ui.OpText && o.Text.Text == "No matches" {
			found = true
		}
	}
	if !found {
		t.Error("filtered-to-nothing menu does not say so")
	}
}

func TestMenuClampsScrollAndFitsSmallOutputs(t *testing.T) {
	tm := theme.Dark()
	r := software.New(1)
	st := MenuState{Items: menuItems(), Scroll: 999}

	for _, bounds := range []ui.Rect{
		ui.R(0, 0, 1280, 720),
		ui.R(0, 0, 480, 480),
		ui.R(0, 0, 320, 200),
		ui.R(0, 0, 120, 60),
	} {
		list := Menu(bounds, tm, st, r)
		for i, o := range list.Ops() {
			if o.Rect.Intersect(bounds) != o.Rect {
				t.Errorf("bounds %v: op %d (%v) rect %v escapes it", bounds, i, o.Kind, o.Rect)
			}
		}
	}
}

func TestMenuScrollRevealsLaterRows(t *testing.T) {
	tm := theme.Dark()
	r := software.New(1)
	st := MenuState{Items: menuItems(), Scroll: 2, Selected: 3}

	titles := map[string]bool{}
	for _, o := range Menu(ui.R(0, 0, 1280, 720), tm, st, r).Ops() {
		if o.Kind == ui.OpText {
			titles[o.Text.Text] = true
		}
	}
	if !titles["Terminal"] || !titles["A Very Long Application Name Indeed"] {
		t.Errorf("scrolled menu did not show the tail rows, got %v", keys(titles))
	}
}

func TestMenuTrimsLongTitles(t *testing.T) {
	tm := theme.Dark()
	r := software.New(1)
	list := Menu(ui.R(0, 0, 320, 720), tm, MenuState{Items: menuItems()}, r)

	long := "A Very Long Application Name Indeed"
	found := false
	for _, o := range list.Ops() {
		if o.Kind == ui.OpText && strings.HasPrefix(o.Text.Text, "A Very") {
			found = true
			if o.Text.Text == long || !strings.HasSuffix(o.Text.Text, "…") {
				t.Errorf("long title drew as %q, want it trimmed with an ellipsis", o.Text.Text)
			}
		}
	}
	if !found {
		t.Error("long title row was not drawn at all")
	}
}

func TestMenuClipsLabelsToWidth(t *testing.T) {
	tm := theme.Dark()
	r := software.New(1)
	list := Menu(ui.R(0, 0, 1280, 720), tm, MenuState{Items: menuItems()}, r)

	for _, o := range list.Ops() {
		if o.Kind != ui.OpText {
			continue
		}
		if w := r.Measure(o.Text.Font, o.Text.Text).W; w > o.Text.Box.W()+1 {
			t.Errorf("label %q measures %v, wider than its box %v", o.Text.Text, w, o.Text.Box)
		}
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
