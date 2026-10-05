package ui

import "image/color"

// Align positions a text run horizontally within its box.
type Align uint8

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
)

// Weight selects a face from a family. Rasterizers that have no exact match
// are expected to round to the nearest available weight.
type Weight uint8

const (
	Regular Weight = iota
	Medium
	Bold
)

func (w Weight) String() string {
	switch w {
	case Medium:
		return "medium"
	case Bold:
		return "bold"
	}
	return "regular"
}

// Font identifies a typeface at a size in dp. Resolution to a concrete face is
// the Rasterizer's job, so a screen can ask for Segoe UI on Windows and the
// Linux rasterizer can substitute without the screen knowing.
type Font struct {
	Family string
	Weight Weight
	Italic bool
	Size   float32
}

// TextRun draws a single-line string, vertically centred on box.
type TextRun struct {
	Text  string
	Font  Font
	Box   Rect
	Color color.NRGBA
	Align Align
}

type OpKind uint8

const (
	OpRect OpKind = iota
	OpRRect
	OpText
)

// Op is one resolved drawing operation. Clip is the region in force when the op
// was added, already in dp, so a Rasterizer never has to track clip state.
type Op struct {
	Kind   OpKind
	Rect   Rect
	Radius float32
	Fill   color.NRGBA
	Text   TextRun
	Clip   Rect
}

// DrawList is a retained list of drawing operations in surface-local dp,
// resolved against a clip stack as it is built. Screens produce one per frame;
// Rasterizer.Draw consumes it.
type DrawList struct {
	ops       []Op
	clipStack []Rect
	clip      Rect
}

// NewDrawList returns a DrawList whose clip starts at bounds.
func NewDrawList(bounds Rect) *DrawList {
	return &DrawList{clip: bounds, clipStack: []Rect{bounds}}
}

// Rect fills r with c.
func (d *DrawList) Rect(r Rect, c color.NRGBA) {
	d.add(Op{Kind: OpRect, Rect: r, Fill: c})
}

// RoundedRect fills r with c, rounding each corner by radius (clamped to half
// the shorter side).
func (d *DrawList) RoundedRect(r Rect, radius float32, c color.NRGBA) {
	d.add(Op{Kind: OpRRect, Rect: r, Radius: radius, Fill: c})
}

// Text draws a single line, centred vertically in box and aligned per Align.
// Lines longer than box are clipped to box.
func (d *DrawList) Text(t string, f Font, box Rect, c color.NRGBA, a Align) {
	d.add(Op{Kind: OpText, Text: TextRun{Text: t, Font: f, Box: box, Color: c, Align: a}})
}

// PushClip intersects the clip region for subsequent operations.
func (d *DrawList) PushClip(r Rect) {
	d.clipStack = append(d.clipStack, d.clip)
	d.clip = d.clip.Intersect(r)
}

// PopClip restores the region saved by the matching PushClip.
func (d *DrawList) PopClip() {
	if len(d.clipStack) < 2 {
		return
	}
	d.clip = d.clipStack[len(d.clipStack)-1]
	d.clipStack = d.clipStack[:len(d.clipStack)-1]
}

// Ops returns the operations in draw order.
func (d *DrawList) Ops() []Op { return d.ops }

func (d *DrawList) Empty() bool { return len(d.ops) == 0 }

func (d *DrawList) add(o Op) {
	if o.Kind == OpText {
		o.Rect = o.Text.Box
		// A text run is clipped to its own box, so callers get truncation
		// without wrapping it in a PushClip for every label.
		o.Clip = d.clip.Intersect(o.Text.Box)
		d.ops = append(d.ops, o)
		return
	}
	o.Clip = d.clip
	d.ops = append(d.ops, o)
}
