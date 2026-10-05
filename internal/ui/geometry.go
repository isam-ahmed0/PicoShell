// Package ui holds the platform-independent description of PicoShell's
// interface. Nothing in this package may import an OS or windowing package:
// screens are built here and handed to a Rasterizer, which is the only place
// platform rendering leaks in.
//
// All geometry is in dp (device-independent pixels). A Rasterizer converts to
// device pixels using Surface.Scale.
package ui

// Point is a position in dp.
type Point struct{ X, Y float32 }

// Size is a width/height pair in dp.
type Size struct{ W, H float32 }

// Rect is a half-open rectangle in dp: [X0,X1) x [Y0,Y1).
type Rect struct{ X0, Y0, X1, Y1 float32 }

// R builds a Rect from its edges.
func R(x0, y0, x1, y1 float32) Rect { return Rect{x0, y0, x1, y1} }

func (r Rect) W() float32  { return r.X1 - r.X0 }
func (r Rect) H() float32  { return r.Y1 - r.Y0 }
func (r Rect) Size() Size  { return Size{r.W(), r.H()} }
func (r Rect) Empty() bool { return r.X1 <= r.X0 || r.Y1 <= r.Y0 }
func (r Rect) Contains(p Point) bool {
	return p.X >= r.X0 && p.X < r.X1 && p.Y >= r.Y0 && p.Y < r.Y1
}

func (r Rect) Offset(dx, dy float32) Rect {
	return Rect{r.X0 + dx, r.Y0 + dy, r.X1 + dx, r.Y1 + dy}
}

// Inset shrinks r by dx horizontally and dy vertically, clamped at empty.
func (r Rect) Inset(dx, dy float32) Rect {
	return r.Intersect(R(r.X0+dx, r.Y0+dy, r.X1-dx, r.Y1-dy))
}

func (r Rect) Intersect(o Rect) Rect {
	return R(max(r.X0, o.X0), max(r.Y0, o.Y0), min(r.X1, o.X1), min(r.Y1, o.Y1))
}
