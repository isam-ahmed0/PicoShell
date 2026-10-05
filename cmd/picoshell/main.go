// Command picoshell is the shell entry point and the interface preview tool.
package main

import (
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/isam-ahmed0/PicoShell/internal/raster/software"
	"github.com/isam-ahmed0/PicoShell/internal/ui"
	"github.com/isam-ahmed0/PicoShell/internal/ui/screens"
	"github.com/isam-ahmed0/PicoShell/internal/ui/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "preview" {
		if err := preview(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "picoshell:", err)
			os.Exit(1)
		}
		return
	}
	usage()
}

func usage() {
	fmt.Fprint(os.Stderr, `picoshell - shell and desktop environment

  picoshell preview <screen> [-o out.png] [-scale n] [-width dp] [-height dp]
      render a screen to a PNG with the software rasterizer
      screens: panel, menu
`)
	os.Exit(2)
}

func preview(args []string) error {
	if len(args) == 0 {
		return errors.New("preview: need a screen name")
	}
	flags := parseFlags(args[1:])
	t := theme.Dark()
	scale := flags.number("scale", t.Scale)
	r := software.New(scale)

	kind, geom, list, err := render(args[0], r, t, ui.R(0, 0, flags.number("width", 1280), flags.number("height", 720)))
	if err != nil {
		return err
	}
	surface := software.NewSurface(kind, "preview", geom, scale, t.BarBg)
	software.New(surface.Scale()).Draw(surface, list)

	out := flags.str("o", args[0]+".png")
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, surface.Image); err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

// preview is the preview's fixture data: the same screens the real session
// draws, fed a fixed state so a change shows up in the PNG.
type fixture struct {
	kind ui.Kind
	geom ui.Rect
	draw func(ui.Rect) *ui.DrawList
	bg   color.NRGBA
}

func fixtures(r ui.Rasterizer, t theme.Theme, output ui.Rect) []fixture {
	return []fixture{
		{
			kind: ui.Panel,
			geom: ui.R(output.X0, output.Y1-screens.PanelHeight, output.X1, output.Y1),
			bg:   t.BarBg,
			draw: func(b ui.Rect) *ui.DrawList {
				return screens.Panel(b, t, screens.PanelState{
					Clock:  "09:41",
					Unread: 3,
					Items: []screens.PanelItem{
						{Title: "Files", Active: true},
						{Title: "Browser"},
						{Title: "Terminal"},
						{Title: "Settings"},
					},
				})
			},
		},
		{
			kind: ui.Menu,
			geom: output,
			bg:   t.BarBg,
			draw: func(b ui.Rect) *ui.DrawList {
				return screens.Menu(b, t, screens.MenuState{
					Query:    "er",
					Selected: 1,
					Items: []screens.MenuItem{
						{Title: "Files", Hint: "Folders"},
						{Title: "Browser", Hint: "Web"},
						{Title: "Terminal", Hint: "Shell"},
						{Title: "Editor", Hint: "Text"},
						{Title: "Settings", Hint: "System"},
						{Title: "Screenshot", Hint: "Capture"},
					},
				}, r)
			},
		},
	}
}

func render(name string, r ui.Rasterizer, t theme.Theme, output ui.Rect) (ui.Kind, ui.Rect, *ui.DrawList, error) {
	for _, f := range fixtures(r, t, output) {
		if name != f.kind.String() {
			continue
		}
		return f.kind, f.geom, f.draw(f.geom), nil
	}
	names := make([]string, 0, len(fixtures(r, t, output)))
	for _, f := range fixtures(r, t, output) {
		names = append(names, f.kind.String())
	}
	return 0, ui.Rect{}, nil, fmt.Errorf("unknown screen %q, want one of %s", name, strings.Join(names, ", "))
}

type flagSet map[string]string

func parseFlags(args []string) flagSet {
	f := flagSet{}
	for i := 0; i < len(args); i++ {
		key, ok := strings.CutPrefix(args[i], "-")
		if !ok {
			continue
		}
		if i+1 < len(args) {
			f[key] = args[i+1]
			i++
			continue
		}
		f[key] = "1"
	}
	return f
}

func (f flagSet) str(key, def string) string {
	if v, ok := f[key]; ok {
		return v
	}
	return def
}

func (f flagSet) number(key string, def float32) float32 {
	v, err := strconv.ParseFloat(f.str(key, ""), 32)
	if err != nil {
		return def
	}
	return float32(v)
}
