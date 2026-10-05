package ui

// Caps reports what a Rasterizer can do. Screens branch on these instead of on
// GOOS, which keeps one look across platforms: a feature a backend lacks
// degrades to a bundled equivalent instead of disappearing.
type Caps struct {
	// Alpha is per-pixel surface transparency.
	Alpha bool
	// BlurBehind samples what is already on screen behind the surface.
	BlurBehind bool
	// ColorText renders colour fonts (emoji) from the system.
	ColorText bool
}

// Rasterizer presents DrawLists and answers the text metrics they need. Text
// measurement lives here rather than in ui so a layout can never disagree with
// what gets drawn.
type Rasterizer interface {
	Caps() Caps
	Measure(f Font, s string) Size
	LineHeight(f Font) float32
	Draw(s Surface, l *DrawList)
}
