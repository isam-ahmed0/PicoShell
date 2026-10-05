package software

import (
	"image"
	"image/color"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
)

// Surface is a ui.Surface backed by an in-memory image, for previews and
// golden tests.
type Surface struct {
	kind   ui.Kind
	output string
	geom   ui.Rect
	scale  float32
	bg     color.NRGBA

	Image *image.NRGBA
	// Frames counts Invalidate calls, so tests can assert on redraw cadence.
	Frames int
}

// NewSurface allocates a surface of the given dp size.
func NewSurface(kind ui.Kind, output string, geom ui.Rect, scale float32, bg color.NRGBA) *Surface {
	if scale <= 0 {
		scale = 1
	}
	w := int(float32(geom.W()) * scale)
	h := int(float32(geom.H()) * scale)
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	drawUniform(img, bg)
	return &Surface{kind: kind, output: output, geom: geom, scale: scale, bg: bg, Image: img}
}

func (s *Surface) Kind() ui.Kind     { return s.kind }
func (s *Surface) Output() string    { return s.output }
func (s *Surface) Geometry() ui.Rect { return s.geom }
func (s *Surface) Scale() float32    { return s.scale }
func (s *Surface) Invalidate()       { s.Frames++ }
func (s *Surface) Close()            {}

// Clear restores the background, which a real backend does by repainting the
// surface before the DrawList is submitted.
func (s *Surface) Clear() {
	drawUniform(s.Image, s.bg)
	s.Frames++
}

func drawUniform(img *image.NRGBA, c color.NRGBA) {
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
}
