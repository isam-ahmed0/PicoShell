// Package software rasterizes DrawLists into an in-memory image. It is the
// reference implementation of ui.Rasterizer: no GPU, no windowing, so the whole
// interface can be built and golden-tested on any platform, including in CI.
package software

import (
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gobolditalic"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"

	"github.com/isam-ahmed0/PicoShell/internal/ui"
)

// Rasterizer draws at a fixed dp-to-pixel scale. Like every shell UI
// rasterizer, it is used from a single UI thread.
type Rasterizer struct {
	scale float32
	faces map[ui.Font]font.Face
	path  vector.Rasterizer
}

// New returns a Rasterizer where 1dp maps to scale device pixels.
func New(scale float32) *Rasterizer {
	if scale <= 0 {
		scale = 1
	}
	return &Rasterizer{scale: scale, faces: map[ui.Font]font.Face{}}
}

func (r *Rasterizer) Caps() ui.Caps {
	// No per-pixel surface transparency or backdrop blur without a compositor,
	// and the Go font stack has no colour font support.
	return ui.Caps{}
}

func (r *Rasterizer) Measure(f ui.Font, s string) ui.Size {
	face := r.face(f)
	adv := font.MeasureString(face, s)
	m := face.Metrics()
	return ui.Size{W: float32(adv) / 64, H: float32(m.Height.Ceil()) / r.scale}
}

func (r *Rasterizer) LineHeight(f ui.Font) float32 {
	return float32(r.face(f).Metrics().Height.Ceil()) / r.scale
}

// Draw renders into a Surface owned by this package. Other backends implement
// their own Draw over the same ops.
func (r *Rasterizer) Draw(s ui.Surface, l *ui.DrawList) {
	target, ok := s.(*Surface)
	if !ok || target.Image == nil {
		return
	}
	img := target.Image
	for _, o := range l.Ops() {
		switch o.Kind {
		case ui.OpRect:
			r.fillRect(img, o.Rect, o.Fill, o.Clip)
		case ui.OpRRect:
			r.fillRounded(img, o.Rect, o.Radius, o.Fill, o.Clip)
		case ui.OpText:
			r.drawText(img, o.Text, o.Clip)
		}
	}
}

// face resolves and caches a face at the requested size. DPI 72 makes one
// point equal one pixel, so dp and pixels coincide before scaling.
func (r *Rasterizer) face(f ui.Font) font.Face {
	if face, ok := r.faces[f]; ok {
		return face
	}
	var ttf []byte
	switch {
	case f.Weight == ui.Bold && f.Italic:
		ttf = gobolditalic.TTF
	case f.Weight == ui.Bold:
		ttf = gobold.TTF
	case f.Weight == ui.Medium:
		ttf = gomedium.TTF
	case f.Italic:
		ttf = goitalic.TTF
	default:
		ttf = goregular.TTF
	}
	parsed, err := opentype.Parse(ttf)
	if err != nil {
		panic(err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		DPI:     72,
		Size:    float64(f.Size * r.scale),
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}
	r.faces[f] = face
	return face
}

func (r *Rasterizer) fillRect(img *image.NRGBA, rect ui.Rect, c color.NRGBA, clip ui.Rect) {
	dr := r.device(rect).Intersect(r.device(clip))
	if dr.Empty() {
		return
	}
	draw.Draw(img, dr, image.NewUniform(c), image.Point{}, draw.Over)
}

func (r *Rasterizer) fillRounded(img *image.NRGBA, rect ui.Rect, radius float32, c color.NRGBA, clip ui.Rect) {
	dr := r.device(rect).Intersect(r.device(clip))
	if dr.Empty() {
		return
	}
	w, h := float32(dr.Dx()), float32(dr.Dy())
	radius = min(radius*r.scale, min(w, h)/2)
	k := radius * (1 - 0.5522847498307936) // control point for a quarter-circle cubic

	p := &r.path
	p.Reset(dr.Dx(), dr.Dy())
	p.MoveTo(radius, 0)
	p.LineTo(w-radius, 0)
	p.CubeTo(w-radius+k, 0, w, radius-k, w, radius)
	p.LineTo(w, h-radius)
	p.CubeTo(w, h-radius+k, w-radius+k, h, w-radius, h)
	p.LineTo(radius, h)
	p.CubeTo(radius-k, h, 0, h-radius+k, 0, h-radius)
	p.LineTo(0, radius)
	p.CubeTo(0, radius-k, radius-k, 0, radius, 0)
	p.ClosePath()

	// Rasterizer.Draw reads (r.Min.X+x) directly off dst, so the destination
	// rect must already be clipped to the image.
	dest := image.Rectangle{Min: dr.Min, Max: dr.Min.Add(dr.Size())}
	p.Draw(img, dest, image.NewUniform(c), dr.Min)
}

func (r *Rasterizer) drawText(img *image.NRGBA, run ui.TextRun, clip ui.Rect) {
	face := r.face(run.Font)
	m := face.Metrics()

	w := font.MeasureString(face, run.Text)
	switch run.Align {
	case ui.AlignCenter:
		run.Box.X0 += (run.Box.W() - float32(w)/64/r.scale) / 2
	case ui.AlignEnd:
		run.Box.X0 += run.Box.W() - float32(w)/64/r.scale
	}

	baseline := run.Box.Y0 + (run.Box.H()+float32(m.Ascent-m.Descent)/64)/2
	dot := fixed.Point26_6{
		X: fixed.I(int(run.Box.X0 * r.scale)),
		Y: fixed.I(int(baseline * r.scale)),
	}

	// One glyph per rune with no shaping: good enough to lay the interface out,
	// but the Direct2D rasterizer will shape properly and is the one that counts.
	src := image.NewUniform(run.Color)
	for _, ch := range run.Text {
		if ch == '\n' {
			continue
		}
		dr, mask, maskp, adv, ok := face.Glyph(dot, ch)
		if ok && mask != nil {
			visible := dr.Intersect(img.Bounds()).Intersect(r.device(clip))
			if !visible.Empty() {
				shift := visible.Min.Sub(dr.Min)
				draw.DrawMask(img, visible, src, visible.Min, mask, maskp.Add(shift), draw.Over)
			}
		}
		dot.X += adv
	}
}

func (r *Rasterizer) device(rect ui.Rect) image.Rectangle {
	return image.Rect(
		int(rect.X0*r.scale), int(rect.Y0*r.scale),
		int(rect.X1*r.scale), int(rect.Y1*r.scale),
	)
}
