package ui

// Kind identifies a shell surface. A Rasterizer or backend may use it to pick
// window attributes (layer-shell vs appbar vs popup), so screens set it rather
// than deciding themselves.
type Kind uint8

const (
	Panel   Kind = iota // the always-present bar at the bottom of the screen
	Menu                // launcher / start menu
	Popup               // small anchored flyout (calendar, tray overflow)
	Toast               // transient notification
	OSD                 // volume / brightness overlay
	Desktop             // wallpaper + desktop icons
	Lock                // lock and login screen
)

func (k Kind) String() string {
	switch k {
	case Panel:
		return "panel"
	case Menu:
		return "menu"
	case Popup:
		return "popup"
	case Toast:
		return "toast"
	case OSD:
		return "osd"
	case Desktop:
		return "desktop"
	case Lock:
		return "lock"
	}
	return "unknown"
}

// Surface is somewhere a DrawList can be presented. Backends implement it over
// a layer-shell surface (labwc), an appbar window (Windows), or in memory
// (previews and golden tests).
type Surface interface {
	Kind() Kind
	Output() string
	// Geometry is in dp, relative to the surface's own top-left.
	Geometry() Rect
	// Scale converts dp to device pixels.
	Scale() float32
	Invalidate()
	Close()
}
