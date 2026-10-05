// Package theme holds the design tokens every screen draws from. Tokens are
// values, not behaviour, so a theme file can be reloaded at runtime.
package theme

import (
	"image/color"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
)

type Theme struct {
	BarBg     color.NRGBA
	MenuBg    color.NRGBA
	ItemBg    color.NRGBA
	ItemBgOn  color.NRGBA
	Text      color.NRGBA
	TextMuted color.NRGBA
	Accent    color.NRGBA
	Border    color.NRGBA

	Radius float32
	Pad    float32
	Gap    float32

	Family string
	// Scale is the output scale factor surfaces are expected to run at, used by
	// previews to render at device resolution.
	Scale float32
}

func Dark() Theme {
	return Theme{
		BarBg:     color.NRGBA{0x14, 0x16, 0x1c, 0xe6},
		MenuBg:    color.NRGBA{0x1b, 0x1e, 0x26, 0xf5},
		ItemBg:    color.NRGBA{0x24, 0x28, 0x33, 0xff},
		ItemBgOn:  color.NRGBA{0x35, 0x3c, 0x4c, 0xff},
		Text:      color.NRGBA{0xe8, 0xec, 0xf4, 0xff},
		TextMuted: color.NRGBA{0x93, 0x9c, 0xad, 0xff},
		Accent:    color.NRGBA{0x5c, 0xa8, 0xff, 0xff},
		Border:    color.NRGBA{0x2c, 0x31, 0x3d, 0xff},
		Radius:    10,
		Pad:       8,
		Gap:       6,
		Family:    "Pico Sans",
		Scale:     2,
	}
}

func (t Theme) Font(size float32, w ui.Weight) ui.Font {
	return ui.Font{Family: t.Family, Weight: w, Size: size}
}
